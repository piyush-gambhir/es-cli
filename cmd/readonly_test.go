package cmd

import (
	"sort"
	"testing"

	"github.com/spf13/cobra"
)

// expectedMutatingPaths is the authoritative list of commands that must carry
// the Annotations["mutates"] = "true" annotation. If you add a new write
// command, add its full command path here; the test will fail until you do.
var expectedMutatingPaths = []string{
	"es index create",
	"es index delete",
	"es index open",
	"es index close",
	"es index settings",
	"es index mappings",
	"es index rollover",
	"es index reindex",
	"es index alias create",
	"es index alias delete",
	"es index template create",
	"es index template delete",
	"es index component-template create",
	"es index component-template delete",
	"es document index",
	"es document delete",
	"es document bulk",
	"es ingest create",
	"es ingest delete",
	"es ilm create",
	"es ilm delete",
}

// collectCommands walks the command tree and returns every leaf and branch
// command (excluding the root itself).
func collectCommands(root *cobra.Command) []*cobra.Command {
	var all []*cobra.Command
	var walk func(cmd *cobra.Command)
	walk = func(cmd *cobra.Command) {
		for _, c := range cmd.Commands() {
			all = append(all, c)
			walk(c)
		}
	}
	walk(root)
	return all
}

func TestMutatingAnnotations(t *testing.T) {
	root := newRootCmd()

	// Build a set of expected mutating paths for fast lookup.
	expectedSet := make(map[string]bool, len(expectedMutatingPaths))
	for _, p := range expectedMutatingPaths {
		expectedSet[p] = true
	}

	// Walk the command tree and partition into found-mutating and found-non-mutating.
	foundMutating := make(map[string]bool)

	for _, cmd := range collectCommands(root) {
		path := cmd.CommandPath()
		if cmd.Annotations != nil && cmd.Annotations["mutates"] == "true" {
			foundMutating[path] = true
		}
	}

	// 1. Every expected mutating command must actually have the annotation.
	var missing []string
	for _, p := range expectedMutatingPaths {
		if !foundMutating[p] {
			missing = append(missing, p)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Errorf("commands expected to have mutates annotation but do not:\n")
		for _, m := range missing {
			t.Errorf("  - %s", m)
		}
	}

	// 2. No unexpected commands should have the annotation.
	var unexpected []string
	for p := range foundMutating {
		if !expectedSet[p] {
			unexpected = append(unexpected, p)
		}
	}
	if len(unexpected) > 0 {
		sort.Strings(unexpected)
		t.Errorf("commands have mutates annotation but are not in the expected list (add them to expectedMutatingPaths if intentional):\n")
		for _, u := range unexpected {
			t.Errorf("  - %s", u)
		}
	}

	// 3. Spot-check that known read-only commands do NOT have the annotation.
	readOnlyPaths := []string{
		"es index list",
		"es index get",
		"es index stats",
		"es index alias list",
		"es index template list",
		"es index template get",
		"es index component-template list",
		"es index component-template get",
		"es document get",
		"es document mget",
		"es ingest list",
		"es ingest get",
		"es ingest simulate",
		"es ilm list",
		"es ilm get",
		"es ilm explain",
		"es cluster health",
		"es cluster stats",
		"es cluster settings",
		"es node list",
		"es node info",
		"es node stats",
		"es shard list",
	}

	for _, p := range readOnlyPaths {
		if foundMutating[p] {
			t.Errorf("read-only command %q should NOT have mutates annotation", p)
		}
	}
}
