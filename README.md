# Chaosd

Chaosd injects faults into Linux hosts through a command-line tool or an HTTP service.
It records experiments and recovery data in SQLite so that later commands can inspect and recover supported faults.
Faults include process failures, network disruption, resource pressure, and faults in files, JVMs, and services.

See the [Chaosd documentation](https://chaos-mesh.org/docs/chaosd-overview/) for fault-specific requirements, configuration, and examples.
Use `chaosd attack --help` to list fault families and each command's `--help` for actions and flags.

## Install on Linux

Use a disposable Linux host for fault injection.
Operations often require root or Linux capabilities, and some faults need additional system tools or access to a target service.
Clock injection is unsupported on arm64.

For a prebuilt archive, choose a version and architecture from the [installation guide](https://chaos-mesh.org/docs/chaosd-overview/) or [release page](https://github.com/chaos-mesh/chaosd/releases).
Inspect the archive before installing its executable and adjacent `tools` directory.
New tagged archives built by the Ubuntu 24.04 release workflow require glibc 2.34 or later.
Older archives retain their original requirements.

To build from source, use Linux with Go, Make, Git, a C compiler, `curl`, and `tar`.
Use the Go version in [go.mod](go.mod).
SQLite requires cgo and the target architecture's compiler and system libraries.

```bash
git clone https://github.com/chaos-mesh/chaosd.git
cd chaosd
make binary
```

This builds `bin/chaosd` and prepares Linux amd64 or arm64 helpers in `bin/tools`.
Missing helper archives are downloaded into `${XDG_CACHE_HOME:-$HOME/.cache}/chaosd`.
Set `CHAOSD_TOOLS_CACHE` to use another shared directory.

Keep `tools` beside the executable when installing or moving it.
For example, install both in a dedicated directory.

```bash
sudo mkdir -p /usr/local/chaosd
sudo cp -a bin/chaosd bin/tools /usr/local/chaosd/
export PATH="/usr/local/chaosd:$PATH"
chaosd version
```

The `sudo chaosd` examples below require `/usr/local/chaosd` in sudo's command search path.
If sudo uses a different path, use `sudo /usr/local/chaosd/chaosd` instead.

For JVM faults, configure Byteman for the account that runs Chaosd.

```bash
export BYTEMAN_HOME=/usr/local/chaosd/tools/byteman
export PATH="$BYTEMAN_HOME/bin:$PATH"
```

## Build a Docker image

On a native Linux amd64 or arm64 Docker worker, build and verify the runtime image.

```bash
make image-chaosd IMAGE_TAG=dev
docker run --rm chaosd:dev /usr/local/bin/chaosd version
bash hack/verify-image.sh chaosd:dev
```

See the [Makefile](Makefile) and [Dockerfile](Dockerfile) for image options and build stages.

## Try a process fault

Help and version commands do not inject faults.

```bash
chaosd --help
chaosd attack process --help
chaosd version
```

Run the following example only in a disposable Linux environment.
It stops a `sleep` process, then resumes it.
Use the PID rather than a process name, which selects every process with that exact name.

```bash
sleep 600 &
target_pid=$!
sudo chaosd attack process stop --process "$target_pid"
```

Copy the printed experiment UID and replace `<UID>` below with that value.

```bash
sudo chaosd search '<UID>'
sudo chaosd recover '<UID>'
sudo chaosd search '<UID>'
kill "$target_pid"
wait "$target_pid" 2>/dev/null || true
```

Successful recovery resumes the process with `SIGCONT` and records `destroyed`.
Use `search --all` to list experiments and `search --help` for filters.

## Start the HTTP service

Bind to localhost for local use. The default host is `0.0.0.0`.
Plain HTTP has no authentication and allows callers to inject faults on the host.

```bash
sudo chaosd server --host 127.0.0.1 --port 31767
```

From another terminal, check health and read experiment history.

```bash
curl http://127.0.0.1:31767/api/system/health
curl http://127.0.0.1:31767/api/experiments/
```

Create faults with `POST /api/attack/<kind>` and recover them with `DELETE /api/attack/<UID>`.
Available endpoints are registered in [server.go](pkg/server/httpserver/server.go).
Requests use JSON types in [pkg/core](pkg/core), rather than CLI flags.

For TLS, pass both `--cert server.crt` and `--key server.key`.
Attack and experiment routes then use HTTPS on `--https-port`, which defaults to `31768`.
Add `--CA ca.crt` to require a client certificate signed by that CA.
Clients must trust the server certificate and provide their client certificate and key for mutual TLS.
The plain HTTP port continues to expose health and version in both TLS modes.

## Preserve experiment state

Chaosd stores `chaosd.dat` beside the executable, and that directory must be writable by the account running stateful commands.
CLI and service invocations share history when they use the same database.
Moving the executable changes the database location.
Preserve the database, its recovery configuration, and fault-created backups until recovery is complete.

Recovery depends on the fault. Shutdown and reboot cannot be reversed.
A killed process needs an explicit recovery command. File replacement reverses strings and does not restore an exact backup.
Check the target itself after recovery, as well as the recorded experiment status.

Cron schedules use UTC. Schedules and recovery timers are held in memory and do not reload after restart.
Cancel future executions through the same running service that created the schedule, using `DELETE /api/attack/<UID>`.
CLI recovery cannot cancel jobs in that service process. The current run relies on its existing recovery timer.

## Contributing

Use these targets for development. Run the full unit and integration suites only in a disposable Linux environment.
Unit tests include disk operations and subprocesses. Integration tests inject real faults.

| Target | Purpose |
| --- | --- |
| `make check` | Format Go files and check lint and module consistency. |
| `make fmt` | Apply Go formatting and import grouping. |
| `make chaosd` | Build the main binary. Use `make binary` to include helpers. |
| `SWAGGER=1 make chaosd` | Generate and include the optional Swagger UI. |
| `make unit-test` | Run the full Go test suite. |
| `make integration-test` | Build tools and run fault injection tests. |

Pull-request CI requires an Apache License 2.0 declaration in each newly added Go or shell file.

See [AGENTS.md](AGENTS.md) for maintenance constraints and safe validation guidance.
Report bugs and contribute changes through the [Chaosd repository](https://github.com/chaos-mesh/chaosd).

## License

Chaosd is licensed under [Apache License 2.0](LICENSE).
