# go-cli-sdk

go-cli-sdk is the command-line application SDK of Go Elemental, the Standards Lab
organization's Go implementation of the Elemental Architecture. Its package `cli` is a
dispatcher on the standard library's `flag` whose commands build only the go-core graph nodes
they declare, under go-core's lifecycle; it depends on the standard library and go-core only.

The README and the package's `doc.go` document this repository. The
[Go Elemental](https://github.com/standards-lab/architecture/blob/main/standards/go-elemental/README.md) standard states the principles it follows.
This context records only working knowledge the code and the README do not express.

go-cli-sdk is the root of the `cli` goal, whose record is `context/goals/cli.md`. The CLI
architecture pages are drafted in this `context/` and move into `architecture` at cli's sync.
The source is spike-cli-architecture's `cli` package
([github.com/JaimeStill/spike-cli-architecture](https://github.com/JaimeStill/spike-cli-architecture), archived).

## Capability map

Nothing is built yet. The code and `doc.go` become authoritative as each capability lands, and
an unbuilt capability gains written detail when a session is about to build it.

- **Dispatcher and flags**: commands on the standard library's `flag`.
- **Generated help and `Footer`**.
- **Per-command composition**: `Use`, `WithGraph`, and `Get`, so a command builds only the
  graph nodes it declares.
- **`Validate`**.
- **`Streams` and `Run`'s exit codes**.
- **CI and release**: CI runs `mise run check` on every pull request; releases are `v*` tags.
