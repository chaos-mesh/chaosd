# Work on Chaosd

Chaosd is a Linux host fault injector with Cobra CLI commands and a Gin HTTP service.
Both call the same attack service and persist experiments and recovery configuration in SQLite.
The root Go module builds the application and release helpers, including `tools/file`.
Root `go test ./...` covers the file helper with the same dependencies used by release builds.

## Find the implementation

| Location | Responsibility |
| --- | --- |
| `cmd/main.go` | Root commands, logging, and runtime helper environment. |
| `cmd/attack`, `cmd/recover`, `cmd/search`, `cmd/server` | CLI flags, input preparation, and service startup. |
| `pkg/core` | Fault kinds, actions, request types, validation, defaults, and recovery JSON. |
| `pkg/server/chaosd` | Experiment lifecycle, execution dispatch, recovery dispatch, and fault implementations. |
| `pkg/server/httpserver` | Registered routes, JSON binding, TLS, client certificates, and system endpoints. |
| `pkg/server/server.go`, `cmd/server/module.go` | Fx providers and application wiring. |
| `pkg/crclient` | Host PID adapter for the in-process Chaos Mesh daemon. |
| `pkg/container` | Docker and containerd client adapter with its own public API. |
| `pkg/scheduler` | Live UTC cron jobs and recovery timers. |
| `pkg/store` | SQLite stores for experiments, runs, and network rules. |
| `pkg/utils` | Commands, process pools, runtime paths, and shared helpers, including public gRPC helpers. |
| `tools` | `PortOccupyTool` and the file-operation helper. |
| `Makefile`, `images`, `.github/workflows` | Builds, helper downloads, container packaging, CI, and releases. |
| `hack/generate_swagger_spec.sh`, `pkg/swaggerserver`, `docs` | Swagger generation and optional UI. |
| `test/integration_test` | Shell tests that inject real faults. |

The active daemon uses `NodeCRClient` and the Chaosd process PID to operate on the host.
`pkg/container` is not part of that active wiring. Preserve its exported interfaces when updating runtime clients unless the task explicitly changes their API.
Preserve the public gRPC helpers in `pkg/utils` unless the task explicitly changes their API.

## Change fault behavior across its callers

Start with the request and recovery types in `pkg/core`.
Follow CLI construction, HTTP binding, execution, persisted configuration, and recovery for the affected kind.
Update validation and defaults at the input adapters with any request change.
Check both dispatch paths in `pkg/server/chaosd` and the kind mapping in `pkg/core/experiment.go`.
HTTP route registration in `pkg/server/httpserver/server.go` determines availability.
An annotated handler alone does not expose a route.
Host, file, and HTTP faults have CLI commands but no registered HTTP creation routes.

Keep serialized recovery fields compatible with existing experiments.
Implementations can add captured PIDs, original permissions, or other recovery data before the final store update.
Do not discard those updates when refactoring request or lifecycle code.
Preserve the distinction between disk CLI configuration and disk service configuration.

Chaos Mesh main and API modules form a compatibility set.
Network code uses their protobufs and daemon methods. Stress code uses their validators and process manager.
Clock code uses their ptrace helpers. HTTP recovery stores their tproxy configuration.
Check packaged helper compatibility when any of these types change.
Review effective versions and root replacements when changing dependencies.
The root `go.mod` replaces the Chaos Mesh API module, gRPC, and Kubernetes modules.
Check the upstream requirement behind each replacement before removing or changing it.
The file helper shares the root module and its dependency versions.

## Build and check

Read `go.mod` and `.github/workflows/ci.yml` before choosing a Go toolchain.
SQLite requires cgo and a C compiler. The license checker requires a `python3` command in `PATH`.
Build and runtime validation need Linux facilities. Clock injection is unsupported on arm64.
Do not treat a cross-compilation result as proof that a fault works on the target architecture.

Use the repository targets for the relevant change.

```bash
make check
make chaosd
make unit-test
make chaos-tools
```

`make check` runs formatting, vet, boilerplate, lint, and tidy checks.
It reports formatting and module drift without rewriting files.
Use `make fmt` and `go mod tidy` explicitly to apply changes.
`make chaosd` builds `bin/chaosd` with cgo enabled.
`make chaos-tools` builds helpers and downloads pinned Linux tools into `bin/tools`.
`make binary` includes the main binary and tools.
The Makefile selects downloaded helpers from the effective target GOOS and GOARCH.
Cross-compiling the Go binary alone does not produce a complete tool bundle for another architecture.

Help and version are suitable smoke checks without creating experiment state.
Search and server startup can create `chaosd.dat` beside the executable.
For routine validation, select tests that only validate inputs or use mocks.
Tests can run subprocesses and create files.
Run the full unit-test suite and `make integration-test` only in a disposable Linux environment.
Never inject real host faults on a workstation as part of routine validation.
Report whether validation covered static checks, builds, selected tests, or actual fault execution.

Swagger annotations live in the HTTP handlers and reference types in `pkg/core`.
The optional Swagger UI uses the `swagger_server` build tag and generated files in `docs`.
Use `SWAGGER=1 make chaosd` to generate from `cmd/main.go` before the tagged build.
Do not hand-edit generated files in `docs`.
Keep `docs/placeholder.go`, which allows builds without generated documentation.
Review generated output separately from the default build and check the `swagger_server` build tag.

## Preserve operational state

The database is `chaosd.dat` beside the executable, not beside the source or current directory.
Keep it, its serialized recovery configuration, and fault-created backups during maintenance and cleanup.
Do not delete runtime state to make a test or dependency change pass.
Moving a binary changes the database and adjacent helper paths.
Keep `tools` beside deployed executables, or verify the configured external helper paths.

Cron jobs, recovery timers, and disk process pools are in memory.
They do not resume from SQLite after restart.
Cancel a schedule through the same running service that owns its cron entry.
CLI recovery creates a separate scheduler and cannot remove jobs in another process, even when both share the database.
An active scheduled run relies on its existing recovery timer.
Do not imply universal recovery. Host faults, killed processes, and file replacement have specific limits.
Use a fresh experiment UID unless a test explicitly exercises duplicate UID behavior.

## Reuse development caches

Before installing dependencies or building, inspect project conventions and the effective tool cache paths.
Reuse fixed project caches or the tool's default shared caches.
For Go, inspect `go env GOCACHE GOMODCACHE GOPATH` and reuse those paths.
Use `go test -count=1` to rerun tests. Do not clear or copy caches for a fresh run.
Do not create dependency caches per task, branch, worktree, agent, or review snapshot, including under `/tmp`.
Isolate source, runtime state, and test output when needed, while sharing dependency downloads.
`hack/install-tools.sh` owns pinned helper versions and architecture mapping.
It reuses versioned shared archives and matching `bin/tools` outputs.
For container builds, reuse project BuildKit cache mounts or fixed cache volumes.
Keep cache IDs tied to compatible platforms and toolchains, not tasks or dates.
If a shared cache is unwritable, use an existing writable shared cache or the available permission process.
Record the reason, path, and cleanup plan for a necessary temporary cache exception.
Pass the selected shared cache paths to any delegated work.
Clean up only this task's unused temporary outputs, containers, and mounts.
Do not clear shared caches, remove source or recovery backups, or run global Docker prune commands.
