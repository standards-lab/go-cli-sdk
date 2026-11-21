# go-cli-sdk

go-cli-sdk is the command-line application SDK of Go Elemental, the Standards Lab
organization's Go implementation of the Elemental Architecture. Its root package `cli` is a
dispatcher on the standard library's `flag` whose commands build only the go-core graph nodes
they declare, under go-core's lifecycle. Its package `scenario` runs a tool's narrated
scenarios as `cli` commands. The module depends on the standard library and go-core only.

The README, each package's `doc.go`, and `docs/` document this repository. The
[Go Elemental](https://github.com/standards-lab/architecture/blob/main/standards/go-elemental/README.md) standard states the principles it follows.

## Capability map

- **Dispatcher and flags** (built): command trees on the standard library's `flag`, with root
  flags accepted at any depth and required and exclusive flags ([doc.go](../doc.go)).
- **Generated help and `Footer`** (built): each command's help, generated from its tree and flags,
  with the command's own text appended ([doc.go](../doc.go)).
- **Per-command composition** (built): `Use`, `Get`, and `Leaves` build only the graph nodes a
  command declares ([doc.go](../doc.go)).
- **`Validate`** (built): a leaf's input is checked before anything is built ([doc.go](../doc.go)).
- **`Program` and `Streams`** (built): dispatch a command tree and return go-core's exit codes
  ([doc.go](../doc.go)).
- **Scenario runner** (built): runs a tool's narrated scenarios as `cli` commands
  ([scenario/doc.go](../scenario/doc.go)).
- **CI and release** (built): CI runs `mise run check` on every pull request and push to `main`;
  releases are `v*` tags ([ci.yml](../.github/workflows/ci.yml),
  [release.yml](../.github/workflows/release.yml)).
