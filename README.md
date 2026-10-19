# Chaosd

Chaosd injects faults into Linux hosts through a command-line tool or an HTTP service.
It records experiments and recovery data in SQLite so that later commands can inspect and recover supported faults.

See the [Chaosd documentation](https://chaos-mesh.org/docs/chaosd-overview/) for fault-specific configuration and examples.

## Choose a fault

The CLI exposes these fault families. Use each command's `--help` for its actions and flags.

| Command | Faults |
| --- | --- |
| `process` | Stop or kill processes. |
| `network` | Delay, loss, corruption, duplication, partition, DNS, bandwidth, port occupation, flood, and interface failure. |
| `stress` | CPU and memory pressure. |
| `disk` | Read load, write load, and disk fill. |
| `host` | Shutdown and reboot. |
| `jvm` | Method latency, return values, exceptions, stress, garbage collection, and Byteman rules. |
| `clock` | Time offsets in a target process. |
| `kafka` | Disk fill, message flood, and broker I/O faults. |
| `redis` | Sentinel faults and cache faults. |
| `file` | Create, modify, delete, rename, append, and replace file contents. |
| `http` | Abort or delay HTTP traffic, apply proxy configuration, and send requests. |
| `vm` | Stop a virtual machine through libvirt. |
| `user-defined` | Run custom attack and recovery commands. |

The HTTP service creates faults at `POST /api/attack/<kind>` for `process`, `network`, `stress`, `disk`, `jvm`, `clock`, `kafka`, `redis`, `http`, `vm`, and `user_defined`.
Host and file faults are available through the CLI only.
HTTP requests use JSON fields defined in `pkg/core`. They do not accept CLI flags.

## Prepare a Linux host

Build with Go, Make, Git, and a C compiler. Use the Go version configured in the [CI workflow](.github/workflows/ci.yml).
The build downloads helper tools with `curl` and extracts archives with `tar`.
The SQLite driver requires cgo. The Makefile enables cgo for the main binary and helper builds.
Use a Linux build environment with the target architecture's compiler and system libraries.

Runtime requirements depend on the fault you choose.

- Network faults use Linux tools such as `tc`, `ipset`, and `iptables`. Some actions also require `ifconfig` or `iperf`.
- CPU and memory stress use `stress-ng` and `memStress`.
- JVM faults require a compatible Java process and Byteman.
- HTTP proxy faults use `tproxy`. File faults use `FileTool`, and port occupation uses `PortOccupyTool`.
- Disk faults use `dd` or `fallocate`. VM faults use `virsh`.
- Kafka and Redis faults require access to their target services and any configured reload or recovery commands.

Run faults with the privileges required by the target operation, often root or Linux capabilities.
Clock injection is unsupported on arm64.
Use a disposable Linux host to try fault injection or run integration tests.

## Build and install

Clone the repository and build the executable.

```bash
git clone https://github.com/chaos-mesh/chaosd.git
cd chaosd
make chaosd
```

Build the local helpers and download the pinned Linux tools.
This step supports Linux amd64 and arm64. Downloads reuse versioned archives under `${XDG_CACHE_HOME:-$HOME/.cache}/chaosd`; set `CHAOSD_TOOLS_CACHE` to use another shared directory. Network access is needed for missing archives.

```bash
make chaos-tools
```

The output is `bin/chaosd` with helpers under `bin/tools`.
Keep `tools` beside the executable when you install or move it.
Chaosd adds that directory to its helper search path.
For JVM faults, set `BYTEMAN_HOME` to the installed Byteman directory and include its `bin` directory in `PATH`.

For example, install the executable and helpers in a dedicated directory.

```bash
sudo mkdir -p /usr/local/chaosd
sudo cp -a bin/chaosd bin/tools /usr/local/chaosd/
export PATH="/usr/local/chaosd:$PATH"
chaosd version
```

For JVM faults, configure Byteman for the account that runs Chaosd.

```bash
export BYTEMAN_HOME=/usr/local/chaosd/tools/byteman
export PATH="$BYTEMAN_HOME/bin:$PATH"
```

To use a prebuilt archive, choose its version and architecture from the [installation guide](https://chaos-mesh.org/docs/chaosd-overview/) or [release page](https://github.com/chaos-mesh/chaosd/releases).
Download and inspect the selected archive before installing its executable and adjacent `tools` directory.
Replace `<archive-url>` with the selected download URL.

```bash
curl -fL '<archive-url>' -o chaosd.tar.gz
tar -tzf chaosd.tar.gz
tar -xzf chaosd.tar.gz
```

## Build a Docker image

Build the runtime image with the current checkout's version information.
Docker builds support native Linux amd64 and arm64 workers.

```bash
make image-chaosd IMAGE_TAG=dev
docker run --rm chaosd:dev /usr/local/bin/chaosd version
bash hack/verify-image.sh chaosd:dev
```

Set `IMAGE_NAME`, `IMAGE_TAG`, and `DOCKER_REGISTRY_PREFIX` to change the image name, for example `DOCKER_REGISTRY_PREFIX=registry.example.com/team/`.
Use `DOCKER_BUILD_ARGS` for Docker options and `SWAGGER=1` for the Swagger UI.

The root Dockerfile uses official Go and Debian images and needs no prebuilt Chaosd images.
Direct builds omit Git version information because the Docker context excludes `.git`.

```bash
docker build -t chaosd:dev .
docker run --rm chaosd:dev /usr/local/bin/chaosd --help
```

Use `make image-build-base` for the compiler image `chaosd-build-base:local`.
Use `make image-binary` for the scratch artifact image `chaosd-binary:local`, which contains `/bin/chaosd` and `/bin/tools`.
The runtime image installs the executable and helpers under `/usr/local/bin`.
All image targets reuse shared BuildKit caches.

## Inspect commands and recover a process

Help and version commands do not inject faults.

```bash
./bin/chaosd --help
./bin/chaosd attack --help
./bin/chaosd attack network --help
./bin/chaosd version
```

The following example stops a disposable `sleep` process, then resumes it.
Run it only in a disposable Linux environment.
Use the PID rather than a process name. A name selects every process with that exact name.

```bash
sleep 600 &
target_pid=$!
sudo ./bin/chaosd attack process stop --process "$target_pid"
```

Copy the printed experiment UID. Replace `<UID>` in the commands below with that value.

```bash
sudo ./bin/chaosd search '<UID>'
sudo ./bin/chaosd recover '<UID>'
sudo ./bin/chaosd search '<UID>'
kill "$target_pid"
wait "$target_pid" 2>/dev/null || true
```

Successful recovery resumes the process with `SIGCONT` and records `destroyed`.
Use `search --all` to list experiments and `search --help` to inspect filters.
Use `search --status scheduled` to find experiments recorded as scheduled.

## Start the HTTP service

Bind to localhost for local use. The default host is `0.0.0.0`.
Plain HTTP has no authentication and allows callers to inject faults.

```bash
sudo ./bin/chaosd server --host 127.0.0.1 --port 31767
```

From another terminal, check health and read experiment history.

```bash
curl http://127.0.0.1:31767/api/system/health
curl http://127.0.0.1:31767/api/experiments/
```

Recover an experiment with `DELETE /api/attack/<UID>`.
Read scheduled run history at `GET /api/experiments/<UID>/runs`.
Use the local platform. The active service operates on the host.

For TLS, pass both `--cert server.crt` and `--key server.key`.
Attack and experiment routes then use HTTPS on `--https-port`, which defaults to `31768`.
Add `--CA ca.crt` to require a client certificate signed by that CA.
Clients must trust the server certificate and, for mutual TLS, provide their client certificate and key.
The plain HTTP port continues to expose health and version in both TLS modes.

## Preserve experiment state

Chaosd stores `chaosd.dat` beside the executable, rather than in the current working directory.
The executable directory must be writable by the account that runs stateful commands.
CLI and service invocations share history when they use the same database.
Moving the executable to another directory changes the database location.
Preserve the database, its recovery configuration, and fault-created backups until recovery is complete.

Recovery depends on the fault. Shutdown and reboot cannot be reversed.
A killed process needs an explicit recovery command. File replacement reverses strings and does not restore an exact backup.
Check the target itself after recovery, as well as the recorded experiment status.

Schedules and recovery timers exist only in the running service, and cron schedules use UTC.
A service restart does not reload schedules or timers from the database.
To cancel future scheduled executions, send `DELETE /api/attack/<UID>` to the same running service that created the schedule.
CLI recovery uses a separate scheduler and cannot cancel jobs in that service process.
The current run relies on its existing recovery timer.

## Develop and validate

Run `make check` for formatting, vet, license headers, lint, and module consistency.
The license checker requires Python 3.
`make check` reports formatting and module drift without changing files. Use `make fmt` or `go mod tidy` to apply the corresponding changes.
Inspect the resulting diff before committing.
Run `make chaosd` to build the main binary, or `make binary` to include helper tools.
Use `SWAGGER=1 make chaosd` to generate and include the optional Swagger UI.
`make image-chaosd` builds the runtime image with the current checkout version.

Run `make unit-test` in a disposable Linux environment.
The suite includes disk operations and real subprocesses.
`make integration-test` builds tools and injects real faults, including network, process, stress, and file faults.
Use targeted package tests for routine validation and `go test -count=1` when a fresh test run is required.

See [AGENTS.md](AGENTS.md) for the source map and maintenance constraints.
Report bugs and contribute changes through the [Chaosd repository](https://github.com/chaos-mesh/chaosd).
Chaosd is licensed under [Apache License 2.0](LICENSE).
