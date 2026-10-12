# goal · cli

- **State:** building
- **Task:** sdk
- **Branch:** sdk

## Tasks

1. [x] role-interfaces
2. [ ] sdk
3. [ ] template

## Task brief · sdk

```
Problem       The Elemental Architecture's second application type
              has no SDK. CLIs still run on cobra, and the template
              and slab wait on a released go-cli-sdk. The spike
              proved the API, and go-core v0.8.0 now has the graph
              and lifecycle it needs. go-cli-sdk holds no code, no
              check, no CI, and no currency command.
Behaviors     1. A leaf accepts its flags before, between, and
                 after its positional arguments, up to a "--".
                 Flags defined on the root are accepted at any
                 depth and set one value. A command below the root
                 that redefines a root flag panics at dispatch.
              2. A dispatch to a leaf checks Args, then Require,
                 then Exclusive, then Validate, then runs the
                 root's PreRun, builds the declared nodes, and
                 runs the leaf. It stops at the first failure, and
                 a failure before the build builds nothing.
              3. -h or --help, or a parent run with no subcommand,
                 prints the generated help to stdout and returns
                 go-core's ExitUsage. The help lists the
                 subcommands, the command's own flags, and the
                 global flags as --name, with required, repeatable,
                 and default markers, then what the command's
                 Footer writes. There is no help or completion
                 command and no shorthand flag.
              4. These print to stderr with the command's short
                 usage and return ExitUsage: an unknown command or
                 flag, a malformed flag value, a rejected argument
                 count, a missing required flag, two flags of one
                 exclusive group, any Validate error, and a
                 UsageError anywhere in PreRun's or Run's error
                 chain. Any other error prints once and returns
                 ExitFailure. Success returns ExitOK.
              5. Invocation.Changed reports a flag given on the
                 command line, even at its default value. A
                 StringsVar flag appends each occurrence, and the
                 first occurrence replaces the default.
              6. A leaf whose path declares nodes with Use builds
                 their union once each, plus WithGraph's lifecycle
                 config node, from WithGraph's graph. It runs under
                 go-core's Coordinator and shuts down in reverse
                 order on success, on error, and when the context
                 ends. A leaf whose path declares nothing builds
                 nothing. Use anywhere in a tree with no WithGraph
                 panics at dispatch, whichever command is selected.
              7. Invocation.Get returns a declared node's value. It
                 panics, naming the command and the node, on a nil
                 node, on a node the path didn't declare (even one
                 the build reached as a dependency), and on a read
                 before the build.
              8. A build or startup failure is reported once, with
                 go-core's labels ("startup: <node>: <err>",
                 "shutdown: ..."), returns ExitFailure, and shuts
                 down whatever had started. A Monitored node's
                 failure ends the running command's context.
              9. The module requires go-core v0.8.0 and imports
                 only the standard library, go-core, and itself. A
                 depguard rule in the check fails any other
                 import.
              10. The check runs in CI on every pull request into
                  main and every push to it. A pushed v* tag
                  creates a GitHub release from that version's
                  changelog section. The currency command exits 0.
                  The repository's merge waits on
                  `gh pr checks --watch`, and its ci wait follows
                  main's push run.
              11. go-cli-sdk's GitHub settings match go-core's: a
                  "main" ruleset that requires a pull request and
                  forbids deleting or force-pushing the branch
                  (org-admin bypass), Dependabot security updates
                  on, merge commits only, and branches deleted on
                  merge.
              12. The package comment lists every exported name.
                  The README gives the standard, the package, and
                  development. CHANGELOG v0.1.0 records the first
                  release. The context capability map marks each
                  capability built.
Test seams    cli.Run(ctx, root, args, Streams over buffers,
              WithGraph(...)), read through its exit code and the
              stdout and stderr buffers (behaviors 1-8). The check,
              the currency command, CI, and `gh api` reads of the
              repository settings cover behaviors 9-11.
Slices        1. Scaffolding and CI: the module on go 1.27 and
                 go-core v0.8.0, mise tasks (check, currency,
                 upgrade, conveniences), lint with depguard, the CI
                 and release workflows, CHANGELOG [Unreleased], the
                 restored CI-gated merge and ci wait, and the GitHub
                 settings. Done when mise run currency exits 0 and
                 the check passes (behaviors 9-11)
              2. Dispatcher: the command tree, flags, root flags,
                 StringsVar, NoArgs and ExactArgs, Require and
                 Exclusive, Changed, UsageError, generated help and
                 Footer, Streams, and Run's exit codes. The ported
                 dispatcher tests pass over buffers (behaviors 1-5)
              3. Per-command composition: Use, WithGraph, Get, and
                 Validate before the build, under go-core's
                 Coordinator. The ported composition tests show
                 declared-only builds, reverse shutdown, and
                 "startup:" reporting (behaviors 6-8)
              4. Docs and release: the package comment inventory,
                 the README, CHANGELOG v0.1.0, and the capability
                 map (behavior 12)
Out of scope  go-cli-sdk-template and its settings (template task);
              an App type or a scenario runner in the SDK; a help
              or completion command, shorthand flags, or a Long
              counterpart; any change to go-core or go-web-sdk;
              moving the CLI architecture pages (cli's sync); slab
              (slab goal)
Door          one-way for the go-cli-sdk v0.1.0 tag, since its
              import path and API reach the module proxy. Code,
              docs, CI, and the GitHub settings before the tag are
              two-way.
Release       go-cli-sdk v0.1.0
```

## Progress

slices 4/4 committed · standards ✓ · spec ✓ · editor ✓

## Decisions

- plan: go-cli-sdk is cli's root, not go-core: the CLI architecture pages are drafted in the
  root's `context/`, which belongs to the CLI SDK.
- plan: go-cli-sdk merges with a plain `gh pr merge --merge --delete-branch` and no `ci` until the
  sdk task's CI lands: role-interfaces' record-only pull request there has no checks, and `gh pr
  checks` exits 1 on none.
- plan: go-cli-sdk and go-cli-sdk-template joined `order`, the references catalog, and the
  exclusive group at cli's first plan, not at sync, since LOCATE finds a repository only through
  `order`.
- role-interfaces: go-core infers a readiness check only from a node value that is also a
  `lifecycle.Starter`; rejected inferring one from any value with `Ready() bool`, which lets
  reactor sources join /readyz by accident.
- role-interfaces: `web.Doctor` stays in go-web-sdk; rejected moving it to go-core's lifecycle,
  since RegisterHealth is its only reader and a CLI has no probes.
- role-interfaces: `lifecycle.Subsystem` is kept as the name for Starter plus Stopper, like
  `io.ReadWriteCloser`; rejected removing it, which breaks compile-time proofs in five
  repositories.
- role-interfaces: `Source[T]` keeps `Ready()`; rejected renaming it, since non-Starters no
  longer join /readyz.
- role-interfaces: go-observability and go-database left cli's repos: the Telemetry.Shutdown fix
  shipped in go-observability v0.2.0 and the `lc.Add` doc examples left in go-database v0.8.0;
  their next goal's currency takes go-core v0.8.0.
- role-interfaces: go-web-sdk releases v0.15.2; rejected v0.16.0, since its code and importers'
  builds are unchanged.
- sdk: builds on go-core v0.8.0.
- role-interfaces: upgrade slices take jdx/mise-action v5.1.2 in go-core and go-web-sdk, a patch
  with nothing to adapt.
- role-interfaces: go-cli-sdk's failing currency (no mise.toml) is left to the sdk task.
- role-interfaces: go-web-sdk builds against the local go-core branch through a temporary
  `replace => ../go-core` (go.sum emptied by tidy); SHIP drops it once go-core v0.8.0 is tagged,
  then `go get go-core@v0.8.0` and `go mod tidy` before go-web-sdk publishes.
- role-interfaces: go-core exports `lifecycle.ReadyStarter` (Starter + ReadinessChecker) in
  v0.8.0, so a value that must stay in /readyz can prove it at compile time; rejected adding it
  later or never, since the rule now hinges on the pair. Named ReadyStarter, not StartReadier.
- role-interfaces: go-web-sdk's readiness tests define `toggleStarter`, a `toggleChecker` with a
  no-op `Start`, for every node that must appear on /readyz, and keep `toggleChecker` as the
  Ready-only node that must not.
- role-interfaces: the `var _ lifecycle.ReadyStarter = probed(nil)` proof lives in
  `lifecycle/readiness_test.go`, the external `lifecycle_test` package, beside the package's other
  compile-time interface proofs.

- sdk: package `cli` sits at go-cli-sdk's module root (github.com/standards-lab/go-cli-sdk), as
  go-web-sdk's `web` does; rejected a `cli/` subdirectory, which adds a path segment to every
  import of an application SDK whose packages keep short names.
- sdk: v0.1.0 ports spike-cli-architecture's exported `cli` API as it stands, re-pointed at
  go-core v0.8.0's graph and lifecycle; rejected revising it before the first tag, since the spike
  proved it and nothing in it has a standard library replacement (flag.Func can't replace
  StringsVar's default on first set or mark it repeatable).
- sdk: architecture's dependencies.md names go-cli-sdk's depguard as the one place a lint enforces
  the dependency line; rejected leaving the page saying no lint does.
- sdk: go-cli-sdk's GitHub settings (the "main" ruleset, Dependabot security updates, merge
  commits only, delete branch on merge) are applied in the CI slice before v0.1.0 is tagged;
  rejected handing them to the session as a note, since release-and-ci.md requires the ruleset in
  every repository.
- sdk: go.mod gains its go-core v0.8.0 require in the dispatcher slice, not the scaffolding
  slice, since the check's `go mod tidy -diff` rejects a require nothing imports.
- sdk: go-core v0.8.0's labels replace the spike's in reported errors: "startup: <node>: <err>",
  "shutdown: drain timeout after ...", and "run: ..." for a Monitored failure.
- sdk: a tree may be dispatched more than once. Each dispatch parses a fresh flag set, so
  Changed, Require, and Exclusive see only its flags, and first restores, newest first, only the
  flags the previous dispatch set, or that a failed parse wrote, to their pre-dispatch values. A
  single dispatch calls no Set beyond the command line's and keeps program-assigned values;
  flag.Func, flag.BoolFunc, and other Values without flag.Getter are never restored; a restore
  whose Set rejects the captured String panics, as Run's doc comment states. Rejected resetting
  every flag to its DefValue, which overwrote program-assigned values on a single dispatch.
- sdk: the Monitored-failure test's leaf returns nil once its context ends, so stderr holds one
  "run:" line; returning ctx.Err() would join "context canceled" to it.
- sdk: the capability map and CHANGELOG say CI runs on pull requests and on pushes to main,
  matching ci.yml.

## Pending edits

- architecture · `standards/go-elemental/principles/lifecycle-and-context.md`: a readiness check
  is inferred only from a Starter that implements Ready, a `lifecycle.ReadyStarter`.
- standards-lab · `context/cli-applications.md` ("Role interfaces"): add a ReadyStarter row
  (go-core lifecycle, Starter + ReadinessChecker, implemented by DB, Store, admin.Service,
  Reactor, read by `Coordinator.Checks`), point ReadinessChecker's "Read by" at it, and make the
  "Readiness checks come only from Starters" rule name `lifecycle.ReadyStarter`.
- architecture · `standards/go-elemental/principles/dependencies.md`: the paragraph saying no
  lint, build check, or CI step enforces the line names go-cli-sdk as the exception. A depguard
  rule in its check holds the module to the standard library and go-core, so an import outside
  that line fails the check and CI. Every other repository holds the line by discipline and review.
- standards-lab · `context/cli-applications.md` ("The plan", the go-cli-sdk bullet): go-cli-sdk
  v0.1.0 is built, so the bullet states that it provides package `cli` over the standard library
  and go-core, links its README and `doc.go` for the API and the left-out features in place of
  listing them, and keeps only why the `App` type stays in the template's `internal/app` and why
  the scenario runner is the tool's own package, promoted only on fit.
