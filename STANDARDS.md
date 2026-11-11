# go-cli-sdk standards

The judgement calls the standards-reviewer applies to go-cli-sdk, beyond what `mise run check` enforces.

- A change that alters documented behavior updates the README and the affected `doc.go` in the same change.
- `architecture/standards/go-elemental/principles/dependencies.md`: the standard library and go-core only, across `go.mod` and every import.
- `architecture/standards/go-elemental/principles/tests-and-docs.md`: the doc.go inventory of `cli`.
- `architecture/standards/go-elemental/principles/release-and-ci.md`: the check, currency, and `v<semver>` tags cut from `CHANGELOG.md`.
- `architecture/standards/go-elemental/principles/lifecycle-and-context.md`: a command builds only the graph nodes it declares and runs them under go-core's lifecycle; `Program.Run`'s exit codes, with go-core's `ExitUsage` on a usage error and on requested help.
- `architecture/standards/go-elemental/principles/construction-and-execution.md`: `Program`, whose `Run` panics naming a nil `Root` or the missing one of `Graph` and `LifecycleConfig`, and every new exported function and constructor.
- `architecture/principles/validation-first.md`: `Validate`, before a command's graph is built.
- `architecture/principles/context-architecture.md`: the README and `doc.go` are the homes.
