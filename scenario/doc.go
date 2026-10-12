// Package scenario runs narrated scenarios: a tool's tours of its own
// commands, each an ordered list of steps over the graph nodes it
// declares. [Commands] builds the "scenario" parent with one leaf per
// scenario, which a tool mounts under its root, and [WriteListing] writes
// the listing of the scenarios that the parent's help ends with.
package scenario
