# go-cli-sdk

go-cli-sdk is the command-line application SDK of Go Elemental, the Standards Lab
organization's Go implementation of the Elemental Architecture. Its root package `cli` is a
dispatcher on the standard library's `flag` whose commands build only the go-core graph nodes
they declare, under go-core's lifecycle. Its package `scenario` runs a tool's narrated
scenarios as `cli` commands. The module depends on the standard library and go-core only.

The README and each package's `doc.go` document this repository. The
[Go Elemental](https://github.com/standards-lab/architecture/blob/main/standards/go-elemental/README.md) standard states the principles it follows.
This context records only working knowledge the code and the README do not express.

go-cli-sdk is the root of the `cli` goal, whose record is `context/goals/cli.md`. The CLI
architecture pages are drafted in this `context/` and move into `architecture` at cli's sync.
The source is spike-cli-architecture's `cli` package
([github.com/JaimeStill/spike-cli-architecture](https://github.com/JaimeStill/spike-cli-architecture), archived).

## Capability map

Every capability below is built, as of v0.2.0. The code, `doc.go`, and the README are
authoritative for each; this list only names them.

- **Dispatcher and flags**: commands on the standard library's `flag`.
- **Generated help and `Footer`**.
- **Per-command composition**: `Use`, `WithGraph`, and `Get`, so a command builds only the
  graph nodes it declares.
- **`Validate`**.
- **`Streams` and `Run`'s exit codes**.
- **Scenario runner**: package `scenario`, whose `Commands` and `WriteListing` mount a tool's
  narrated scenarios as `<tool> scenario <name>` leaves.
- **CI and release**: CI runs `mise run check` on every pull request and push to `main`;
  releases are `v*` tags.
