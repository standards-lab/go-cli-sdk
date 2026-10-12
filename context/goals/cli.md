# goal · cli

- **State:** brief ready
- **Task:** role-interfaces
- **Branch:** role-interfaces

## Tasks

1. [ ] role-interfaces
2. [ ] sdk
3. [ ] template

## Task brief · role-interfaces

```
Problem       go-core's Coordinator turns every graph node value with
              Ready() bool into a readiness check, so a value "ready" in
              another sense, such as a reactor source, joins /readyz by
              accident. v1.messaging moves those sources into go-core,
              and go-cli-sdk builds on the go-core release that settles
              the rule.
Behaviors     1. Coordinator.Checks holds one Check for each node value
                 that implements both Starter and ReadinessChecker,
                 named by its node, in layer order and then definition
                 order.
              2. A node value with Ready() bool and no Start yields no
                 Check.
              3. A Starter with Ready yields a Check whether or not it
                 also implements Stopper.
              4. A *lifecycle.Readiness yields no Check, and Readiness
                 reports its Coordinator's narrowed Checks.
              5. RegisterHealth's /readyz over a Coordinator, or a
                 Readiness bound to one, lists "lifecycle" and then only
                 the Starter checks; a non-Starter Ready node is absent
                 from the 200 and 503 bodies.
              6. go-core's lifecycle docs and go-web-sdk's RegisterHealth
                 doc state the Starter-only rule; each CHANGELOG records
                 its release, with go-core's marked Breaking.
              7. go-core exports lifecycle.ReadyStarter, Starter plus
                 ReadinessChecker, as the set Checks is inferred from,
                 with a compile-time proof; v0.8.0's CHANGELOG lists it
                 under Added.
Test seams    lifecycle.New(sys, cfg).Checks() over a built graph.System;
              RegisterHealth's /readyz over HTTP for behavior 5
Slices        1. go-core upgrade: jdx/mise-action v5.1.1 -> v5.1.2;
                 done when mise run currency exits 0 and the check
                 passes
              2. go-web-sdk upgrade: jdx/mise-action v5.1.1 -> v5.1.2;
                 done when mise run currency exits 0 and the check
                 passes
              3. go-core: checks only from Starters, with docs and
                 CHANGELOG v0.8.0. The lifecycle tests show a Ready-only
                 node missing from Checks and a Start+Ready node present
                 (behaviors 1-4, 6, 7)
              4. go-web-sdk on go-core v0.8.0: the RegisterHealth doc and
                 CHANGELOG v0.15.2. A /readyz test over a graph with a
                 Ready-only node shows it absent (behaviors 5, 6)
Out of scope  Source[T] and the reactor (v1.messaging core-reactor);
              go-web-service's Wake comment (v1.messaging
              service-reactors); web.Doctor's home; lifecycle.Subsystem;
              go-observability and go-database; go-web-sdk-template's
              requirement; middleware/rate-limit; go-cli-sdk's code and
              its mise.toml (sdk task)
Door          one-way for the go-core v0.8.0 and go-web-sdk v0.15.2 tags;
              the code and doc changes before tagging are two-way
Release       go-core v0.8.0; go-web-sdk v0.15.2
```

## Progress

slices 4/4 committed · standards done · spec done · editor done

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

## Pending edits

- architecture · `standards/go-elemental/principles/lifecycle-and-context.md`: a readiness check
  is inferred only from a Starter that implements Ready, a `lifecycle.ReadyStarter`.
- standards-lab · `context/cli-applications.md` ("Role interfaces"): add a ReadyStarter row
  (go-core lifecycle, Starter + ReadinessChecker, implemented by DB, Store, admin.Service,
  Reactor, read by `Coordinator.Checks`), point ReadinessChecker's "Read by" at it, and make the
  "Readiness checks come only from Starters" rule name `lifecycle.ReadyStarter`.
