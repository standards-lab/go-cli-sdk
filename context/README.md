# go-cli-sdk

go-cli-sdk is the command-line application SDK of Go Elemental, the Standards Lab
organization's Go implementation of the Elemental Architecture. Its root package `cli` is a
dispatcher on the standard library's `flag` whose commands build only the go-core graph nodes
they declare, under go-core's lifecycle. Its package `scenario` runs a tool's narrated
scenarios as `cli` commands. The module depends on the standard library and go-core only.

The README and each package's `doc.go` document this repository. The
[Go Elemental](https://github.com/standards-lab/architecture/blob/main/standards/go-elemental/README.md) standard states the principles it follows.
This context records only working knowledge the code and the README do not express.

go-cli-sdk's source is spike-cli-architecture's `cli` package
([github.com/JaimeStill/spike-cli-architecture](https://github.com/JaimeStill/spike-cli-architecture), archived).

## Capability map

- **Dispatcher and flags** (built): commands on the standard library's `flag`
  ([command.go](../command.go), [flags.go](../flags.go)).
- **Generated help and `Footer`** (built): [help.go](../help.go).
- **Per-command composition** (built): `Use`, `Get`, and `Leaves`
  ([command.go](../command.go), [invocation.go](../invocation.go)).
- **`Validate`** (built): [command.go](../command.go).
- **`Program` and `Streams`** (built): [run.go](../run.go), [invocation.go](../invocation.go).
- **Scenario runner** (built): package `scenario` ([scenario/doc.go](../scenario/doc.go)).
- **CI and release** (built): CI runs `mise run check` on every pull request and push to `main`;
  releases are `v*` tags ([ci.yml](../.github/workflows/ci.yml),
  [release.yml](../.github/workflows/release.yml)).
