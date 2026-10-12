# goal · cli

- **State:** building
- **Task:** template
- **Branch:** template

## Tasks

1. [x] role-interfaces
2. [x] sdk
3. [ ] template

## Task brief · template

```
Problem       go-cli-sdk v0.1.0 has no CLI scaffold, so a new tool
              still starts from a cobra-era layout or from the
              archived spike. The scenario runner has no shared home,
              so every tool would copy it. go-cli-sdk-template holds
              no code, CI, settings, or currency command. The CLI
              conventions aren't yet in architecture.
Behaviors     go-cli-sdk (scenario package, v0.2.0)
              1. scenario.Commands(scenarios...) returns one "scenario"
                 parent, as a slice, with one leaf per scenario in the
                 order given. Each leaf takes no arguments and declares
                 its scenario's nodes with Use. The parent declares
                 none. Run alone or with -h, the parent prints its help
                 ending with the listing, exits 2, and builds nothing.
                 A usage error on a leaf builds nothing.
              2. WriteListing(w, scenarios...) writes "Scenarios:",
                 then for each scenario its name and summary and a
                 "uses <node>, <node>" line in declared order, columns
                 aligned as cli's help sections are. It builds nothing,
                 and a tool's root Footer can call it so the root help
                 ends with the same listing.
              3. A scenario's leaf runs its steps in order. Before each
                 step it writes a "[i/n] <intent>" heading, set off by
                 one blank line except before the first. Reporter.Note
                 wraps prose to 80 columns, indented. Reporter.Show
                 writes what its render function wrote, indented one
                 level further. It writes nothing when the render fails
                 and returns that error.
              4. A step that fails stops the run: later steps don't
                 run, the error reads "step <i> (<intent>): <err>", and
                 cli reports it once with exit 1. A context that ends
                 fails the step under way.
              5. A scenario builds only the nodes it declares, once
                 each, under go-core's Coordinator. A declared node
                 that fails to start fails the run before the first
                 heading, naming the node ("startup: <node>: ..."), and
                 what had started is shut down.
              6. Commands panics, naming the scenario, when given none,
                 two with one name, or one with no steps.
              7. The module still imports only the standard library,
                 go-core, and itself, and the depguard rule passes. The
                 scenario package's doc.go lists every exported name.
                 The README has a scenario section, CHANGELOG v0.2.0
                 records the addition, and the capability map marks the
                 scenario runner built.
              go-cli-sdk-template (template/v0.1.0)
              8. The module, rooted at the template subtree, requires
                 only go-core v0.8.0 and go-cli-sdk v0.2.0. Its check
                 passes and its currency command exits 0.
              9. "app" with no arguments, -h, or --help prints the root
                 help listing version and exits 2. An unknown command
                 prints the short usage to stderr and exits 2. Neither
                 builds a node or reads the environment.
              10. "app version" prints the main module's version, or
                  "(devel)" when the build has none, exits 0, and
                  builds nothing.
              11. New is cold: it reads nothing and builds nothing. The
                  App publishes Graph, Nodes, and Root. Nodes holds the
                  lifecycle config node, which reads APP_ variables
                  when a build reaches it. A command added at Root that
                  declares it builds it, and an invalid
                  APP_SHUTDOWN_TIMEOUT fails that command, naming
                  "lifecycle config", with exit 1. Run dispatches with
                  WithGraph. The domain layer defines and mounts
                  nothing, and its doc states what goes there. There is
                  no admin layer and no scenario layer.
              12. apptest.Builds records the name of each node a run
                  builds, in build order, and records nothing for a run
                  that builds nothing. apptest.Halt makes the lifecycle
                  config node fail with apptest.ErrHalted, so a command
                  that declares nodes fails before anything starts.
              13. The integration suite builds the binary once with
                  processtest.Main and runs it one shot at a time with
                  processtest.Run on an empty APP_ environment:
                  - version exits 0 with one line on stdout
                  - root help exits 2 on stdout and lists version
                  - an unknown command exits 2 with usage on stderr
              14. main derives go-core's signal context and passes it,
                  cli.Streams over the process's streams, and the
                  arguments to the App, then exits with the code
                  returned.
              15. No file in the module names the template's module
                  path except go.mod and import lines, so a gonew copy
                  at a new path builds and passes its check.
              16. CI runs the check on every pull request into main and
                  every push to it, and the integration suite on pushes
                  and on demand. The root and template copies of the
                  workflow differ only in the subtree-locating lines.
                  A pushed template/v* tag creates a GitHub release
                  from the root changelog's section.
              17. go-cli-sdk-template's GitHub settings match go-core's:
                  a "main" ruleset that requires a pull request and
                  forbids deleting or force-pushing the branch
                  (org-admin bypass), Dependabot security updates on,
                  merge commits only, and branches deleted on merge.
              18. The root README gives the standard, the gonew
                  command, the layout, and development. The starter
                  README gives running, the identity steps (rename
                  cmd/app, the root command, and APP_), and a Scenarios
                  section: opt in by adding the scenarios package and a
                  scenario layer file that mounts scenario.Commands and
                  ends the root Footer with WriteListing. internal/app's
                  doc.go lists its layer files and carries the same
                  opt-in paragraph. apptest's doc.go lists its exports.
                  CHANGELOG template/v0.1.0 records the first release.
                  The capability map drops the example domain and the
                  scenario package. STANDARDS.md's doc.go inventory
                  line drops scenario.
              Goal root
              19. go-cli-sdk's context holds the draft of the Go
                  Elemental "CLI applications" principle page:
                  - internal/app's layer files
                  - Use and inv.Get
                  - Validate building nothing
                  - PreRun flag checks
                  - exit 2 on help and usage errors
                  - bare subcommand words
                  - scenarios as opt-in, with go-cli-sdk's runner and a
                    tool's scenarios package, and no list command
                  The goal record's Pending edits list every edit sync
                  applies (Sync below).
Test seams    go-cli-sdk: cli.Run over a graph and buffers, with
              scenario.Commands mounted at a test root, read through
              the exit code, stdout, and stderr (behaviors 1-6).
              Template: App.Run over buffers with apptest fixtures
              (behaviors 9-12), and processtest.Run over the built
              binary (behaviors 13-14). The checks, the currency
              commands, CI, and gh api reads of the settings cover
              behaviors 7, 8, and 15-17.
Slices        1. go-cli-sdk scenario runner: Scenario, Step, Reporter,
                 Commands, WriteListing, and the runner, ported from
                 the spike with the domain listing removed. Its tests
                 pass over a test graph (behaviors 1-6)
              2. go-cli-sdk docs and release: the scenario doc.go
                 inventory, the README section, CHANGELOG v0.2.0, and
                 the capability map; the check and currency pass
                 (behavior 7). Then go-cli-sdk's reviews and edit
                 pass, and a go-cli-sdk release brief for the
                 architect; on accept, publish, merge, wait for ci,
                 and tag v0.2.0 before any template slice starts
              3. Template scaffolding and CI: the module on go 1.27.2,
                 mise tasks (check, currency, upgrade, build, vet,
                 test, integration, fmt, tidy, lint), lint with
                 testpackage, both CI workflow copies, the release
                 workflow, CHANGELOG [Unreleased], both .gitignore
                 copies, and the GitHub settings (behaviors 16-17)
              4. Template composition root: main, App with Nodes,
                 Graph, and Root, the lifecycle config node under APP_,
                 the empty domain layer, and version, requiring the
                 released go-cli-sdk v0.2.0 directly, with no replace
                 (behaviors 9-11, 14)
              5. apptest and the integration suite: Builds and Halt,
                 internal/app's tests over them, and the processtest
                 one-shot suite (behaviors 12-13)
              6. Template docs: both READMEs, the doc.go files, the
                 capability map, the STANDARDS.md line, CHANGELOG
                 template/v0.1.0, and the path-rewrite constraint
                 (behaviors 15, 18)
              7. Goal root: the CLI applications page draft and the
                 pending edits (behavior 19)
Out of scope  A Table or Record renderer in either repository; an
              example domain, an admin layer, or a scenario layer in
              the template; scenario Flags, Needs, or a colored
              Reporter (slab goal); any change to go-core, go-web-sdk,
              or go-web-sdk-template; applying the architecture and
              coordinator edits (cli's sync)
Door          One-way for go-cli-sdk v0.2.0 and template/v0.1.0, since
              both reach the module proxy. Everything before the tags
              is two-way: code, docs, CI, settings, and the page draft.
              Order: go-cli-sdk's slices finish, pass review, and
              are released as v0.2.0, after the architect accepts a
              go-cli-sdk release brief, before the template's slices
              start. The template pins v0.2.0 from its first slice.
              The goal record and the page draft ride a second
              go-cli-sdk pull request at SHIP, with the session
              brief, beside the template's.
Release       go-cli-sdk v0.2.0, then go-cli-sdk-template
              template/v0.1.0
```

## Progress

slices 2/7 committed · go-cli-sdk: standards ✓ · spec ✓ · editor ✓ · v0.2.0 released (PR #3, 9e5a907) · template: standards — · spec — · editor —

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
- template: the template is a pure skeleton, like go-web-sdk-template: no example domain and no
  sample infrastructure; rejected an in-memory or directory-backed stand-in domain.
- template: no Table/Record renderer in go-cli-sdk or the template; a CLI's layout is each tool's
  own choice, and slab's renderer already differs from the spike's.
- template: the scenario runner moves into go-cli-sdk as the sub-package `scenario` (v0.2.0), and
  the template ships no scenario layer, since not every CLI has scenarios; rejected the tool's
  own runner package and putting the runner inside package `cli`.
- template: a CLI opts into scenarios with `scenario.Commands(scenarios...)` mounted at its root
  and `scenario.WriteListing` called from its own root Footer; rejected one `Mount` that rewrites
  the root's Footer.
- template: `scenario.Scenario` keeps the spike's fields exactly (Name, Summary, Nodes, Steps;
  Reporter with Note and Show); slab's Flags, Needs, and colored Reporter are the slab goal's.
- template: `scenario.Commands` panics on no scenarios, a duplicate name, or a scenario with no
  steps.
- template: a tool's own scenarios live in one root-level package named `scenarios`; rejected
  `demo` and per-domain tours.
- template: apptest ships with the domain-free fixtures Builds and Halt; rejected inline Observe
  and Replace.
- template: no admin layer in the template; internal/app's doc.go says where one goes.
- template: the placeholder identity is "app" (cmd/app, root command "app", env prefix APP_).
- template: the CLI conventions land in architecture as one new Go Elemental principle page, "CLI
  applications", drafted in this context/; rejected folding them into existing pages.
- template: the task touches only go-cli-sdk and go-cli-sdk-template, recorded here without a
  manifest edit, since sync removes goals.cli; go-core and go-web-sdk are unchanged.
- template: go-cli-sdk v0.2.0 is published, merged, and tagged after its slices pass review and
  the architect accepts its release brief, before the template's slices start, so the template
  pins the released version with no replace directive.
- template: `scenario.WriteListing` omits the uses line for a scenario with no nodes, and names a
  node declared twice once, by Name, in first-declared order.
- template: `Reporter.Note` wraps at 80 runes, not bytes.
- template: the runner fails the step under way when the context has ended after its action
  returns nil; when the action returns an error, that error wins over the context's.
- template: `Reporter.Show` prints a blank line of output empty, with no indent.
- template: a scenario node whose constructor fails is reported with cli's and go-core's
  existing label for a Build error, not "startup:".
- template: `scenario.Commands` panics with "scenario: Commands with no scenarios",
  "scenario: duplicate scenario %q", or "scenario: scenario %q has no steps".

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
