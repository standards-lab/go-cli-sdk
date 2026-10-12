# go-cli-sdk

go-cli-sdk is the command-line application SDK for Standards Lab's Go Elemental standard. It
provides a command dispatcher on the standard library's `flag` whose commands build only the
go-core graph nodes they declare, under go-core's lifecycle.

`github.com/standards-lab/go-cli-sdk` is a single Go module with two packages: `cli`, at its
root, and `scenario`, which runs a tool's narrated scenarios on it.

## Standard

`go-cli-sdk` is the command-line application SDK of
[Go Elemental](https://github.com/standards-lab/architecture/blob/main/standards/go-elemental/README.md), the
minimal-dependency Go standard. This README and each package's `doc.go` document the repository.
It depends on the standard library and
[go-core](https://github.com/standards-lab/go-core) alone, a line `depguard` holds in the check.

## Packages

- `cli` — a dispatcher of command trees on `flag`, with generated help, required and exclusive
  flags, input validation, and per-command composition: a command declares the graph nodes it
  needs, and `Run` builds only those and runs the command under go-core's lifecycle, returning
  go-core's exit codes. Its `doc.go` lists every exported name.
- `scenario` — a runner of narrated scenarios, a tool's guided tours of its own commands, as
  `<tool> scenario <name>` leaves. It holds no scenarios itself: a tool keeps its own in one
  root-level package named `scenarios`, mounts `scenario.Commands(scenarios...)` at its root
  with `root.Add`, and calls `scenario.WriteListing` from its root's `Footer` so the root's help
  ends with the listing. Each scenario builds only the graph nodes it declares, and a `Reporter`
  narrates its steps as plain text. Its `doc.go` lists every exported name.

## Development

Tasks run through [mise](https://mise.jdx.dev):

```
mise run check      # build, vet, format, fix, tidy, test, and lint; writes nothing
mise run currency   # report requirements, Go, tools, and actions behind their latest
mise run upgrade    # upgrade the go directive, requirements, and tools to their latest
```

## License

[Apache License 2.0](LICENSE).
