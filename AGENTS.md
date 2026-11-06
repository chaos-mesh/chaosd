# Work on Chaosd

Chaosd is a Linux host fault injector. Its Cobra CLI and Gin HTTP service share the attack service and SQLite stores.
The root Go module includes the application and release helpers, including `tools/file`.

## Find the implementation

| Location | Responsibility |
| --- | --- |
| `cmd` | CLI input, logging, runtime environment, and startup. |
| `pkg/core` | Fault kinds, request validation, defaults, and recovery JSON. |
| `pkg/server/chaosd` | Experiment lifecycle, fault execution, and recovery. |
| `pkg/server/httpserver` | HTTP routes, JSON binding, and TLS. |
| `pkg/server/server.go`, `cmd/server/module.go` | Fx providers and application wiring. |
| `pkg/scheduler`, `pkg/store` | In-memory scheduling and persistent experiment data. |
| `pkg/crclient`, `pkg/container`, `pkg/utils` | Runtime adapters and shared helpers. |
| `Makefile`, `Dockerfile`, `hack`, `.github/workflows`, `tools` | Builds, dependencies, packaging, and CI. |

The active daemon uses `pkg/crclient.NodeCRClient` with the Chaosd process PID to operate on the host.
`pkg/container` is not part of that wiring. Preserve its exported interfaces and the public gRPC helpers in `pkg/utils` unless the task explicitly changes their APIs.

## Change fault behavior

Follow the affected kind through `pkg/core`, CLI construction, HTTP binding, execution, persisted configuration, and recovery.
Update input validation and defaults, both dispatch paths in `pkg/server/chaosd`, and the kind mapping in `pkg/core/experiment.go` as needed.
HTTP availability depends on route registration in `pkg/server/httpserver/server.go`, not handler annotations alone.

Keep serialized recovery fields compatible with existing experiments.
Preserve data captured during execution, such as PIDs and original permissions, through the final store update.
Keep disk CLI configuration distinct from disk service configuration.

Treat the Chaos Mesh main and API modules, their daemon interfaces, and packaged helpers as a compatibility set.
Before changing dependencies or root `go.mod` replacements, check effective versions and the upstream requirements behind each replacement.
The file helper shares the root module and dependency versions, and is covered by root `go test ./...`.

## Validate safely

See [README.md](README.md#contributing) for build and check targets.
Read `go.mod` and `.github/workflows/ci.yml` before selecting a toolchain.
SQLite requires cgo and a C compiler.

`make check` formats files in place. Inspect the diff afterward, as CI rejects formatting drift.
Use `make fmt` for formatting. The tidy check reports drift without rewriting files. Use `go mod tidy` to apply module changes.
Cross-compiling the binary proves neither fault execution nor a complete helper bundle for the target architecture.
Clock injection is unsupported on arm64.

For routine validation, select input-validation tests or mocks.
The full `make unit-test` suite and `make integration-test` require a disposable Linux environment because they can operate on disks, run subprocesses, or inject real faults.
Never inject real host faults on a workstation for routine validation.
Help and version are suitable smoke checks without experiment state. Search and server startup can create `chaosd.dat`.
Report whether validation covered static checks, builds, selected tests, or actual fault execution.

For Swagger changes, update HTTP handler annotations and their `pkg/core` types, then run `SWAGGER=1 make chaosd`.
Review generated output and the `swagger_server` build separately from the default build.
Do not hand-edit generated files in `docs` or remove `docs/placeholder.go`, which supports builds without generated documentation.

## Keep tests focused

- Test critical behavior, regressions, and important boundaries with representative cases. Avoid repeating shared validation or unchanged paths.
- Assert observable outcomes. Keep fixtures, setup, and helpers small and simple.
- Preserve checks that catch key failures when simplifying tests. Prioritize useful coverage over test counts or coverage percentages.

Use a fresh experiment UID unless testing duplicates.

## Preserve operational state

The database is `chaosd.dat` beside the executable, not in the working directory.
Preserve it, serialized recovery configuration, and fault-created backups during maintenance and cleanup.
Do not delete runtime state to make checks pass. Moving the binary changes the database and adjacent helper paths.
Keep `tools` beside deployed executables, or verify configured external helper paths.

Cron jobs, recovery timers, and disk process pools are in memory and do not resume from SQLite after restart.
Cancel schedules through the same running service that owns them. CLI recovery uses a separate scheduler and cannot cancel another process's jobs.
Cancellation does not stop the current run, which relies on its existing recovery timer.
Do not imply universal recovery. Host faults, killed processes, and file replacement have specific limits described in [README.md](README.md#preserve-experiment-state).

## Reuse development caches

Before installing or building, inspect project cache conventions and effective paths, including `go env GOCACHE GOMODCACHE GOPATH`.
Reuse fixed project or default shared caches. Isolate source, test output, and runtime state, not dependency downloads.
Do not create caches per task, branch, worktree, agent, or review snapshot, including under `/tmp`.
Use `go test -count=1` to rerun tests without clearing or copying caches.

`hack/dependency-versions.sh` owns pinned helper versions. `hack/install-tools.sh` owns installation and architecture mapping.
Reuse its versioned archives and matching `bin/tools` outputs. Reuse project BuildKit cache mounts or fixed volumes, with IDs tied to compatible platforms and toolchains.
If a cache is unwritable, use an existing writable shared cache or request permission for the required directory.
A temporary cache exception requires a reason, path, and cleanup plan, and is limited to unavailable or corrupted shared caches or explicit cold-cache validation.
Pass selected cache paths to delegated work.
Clean up only this task's unused temporary outputs, exceptional caches, containers, and mounts.
Do not clear shared caches, remove source or recovery backups, or run global Docker prune commands.

## Prepare changes for delivery

Use `git commit -s` to sign off commits.
Keep changelog entries under `Unreleased` with Keep a Changelog categories until a release is selected.
Describe major capabilities, significant behavior fixes, and necessary upgrade or compatibility information by their user-visible result.
Summarize dependency upgrades together. Omit routine documentation, CI, cache, fixture, and internal changes unless they materially affect users.
For work merged through pull requests, end each item with links to the corresponding PRs.
