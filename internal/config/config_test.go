package config

import "testing"

func TestIsDoneState(t *testing.T) {
	c := &Config{States: StatesConfig{States: []StateConfig{
		{Name: "TODO", Color: "202"},
		{Name: "DONE", Color: "34", Done: true},
		{Name: "CANCELLED", Color: "240", Done: true},
	}}}

	cases := map[string]bool{"TODO": false, "DONE": true, "CANCELLED": true, "": false, "NOPE": false}
	for name, want := range cases {
		if got := c.IsDoneState(name); got != want {
			t.Errorf("IsDoneState(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestFillDefaultsMarksLastStateDoneWhenNoneFlagged(t *testing.T) {
	c := &Config{States: StatesConfig{States: []StateConfig{
		{Name: "TODO", Color: "202"},
		{Name: "DONE", Color: "34"},
	}}}
	c.fillDefaults()

	if c.States.States[len(c.States.States)-1].Name != "DONE" {
		t.Fatalf("states reordered unexpectedly")
	}
	if !c.States.States[len(c.States.States)-1].Done {
		t.Error("expected last state (DONE) to be flagged Done after migration")
	}
	if c.States.States[0].Done {
		t.Error("expected first state (TODO) to remain not-done")
	}
}

func TestFillDefaultsLeavesExplicitDoneFlagsAlone(t *testing.T) {
	c := &Config{States: StatesConfig{States: []StateConfig{
		{Name: "DONE", Color: "34", Done: true},
		{Name: "TODO", Color: "202"},
	}}}
	c.fillDefaults()

	if c.States.States[0].Name != "DONE" || !c.States.States[0].Done {
		t.Error("explicit done flag on DONE should be preserved")
	}
	if c.States.States[1].Done {
		t.Error("last state (TODO) must NOT be auto-flagged when a done state already exists")
	}
}
