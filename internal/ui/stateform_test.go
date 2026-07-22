package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/rwejlgaard/org/internal/config"
)

func newTestModel(states []config.StateConfig) *uiModel {
	return &uiModel{
		config: &config.Config{States: config.StatesConfig{States: states}},
	}
}

func inputWith(val string) textinput.Model {
	ti := textinput.New()
	ti.SetValue(val)
	return ti
}

func TestApplyStateFormAddsPlainState(t *testing.T) {
	m := newTestModel(nil)
	m.stateForm = stateForm{editing: false, name: inputWith("review"), color: inputWith("202"), done: false}

	if !m.applyStateForm() {
		t.Fatalf("expected commit to succeed")
	}
	if len(m.config.States.States) != 1 {
		t.Fatalf("expected 1 state, got %d", len(m.config.States.States))
	}
	s := m.config.States.States[0]
	if s.Name != "REVIEW" || s.Color != "202" || s.Done {
		t.Fatalf("got %+v, want {REVIEW 202 false}", s)
	}
}

func TestApplyStateFormAddsDoneState(t *testing.T) {
	m := newTestModel(nil)
	m.stateForm = stateForm{editing: false, name: inputWith("cancelled"), color: inputWith("240"), done: true}

	m.applyStateForm()

	s := m.config.States.States[0]
	if !s.Done {
		t.Fatalf("expected Done=true, got %+v", s)
	}
}

func TestApplyStateFormEditUpdatesColorAndDone(t *testing.T) {
	m := newTestModel([]config.StateConfig{{Name: "TODO", Color: "202"}})
	m.stateForm = stateForm{editing: true, index: 0, name: inputWith("TODO"), color: inputWith("50"), done: true}

	if !m.applyStateForm() {
		t.Fatalf("expected commit to succeed")
	}
	s := m.config.States.States[0]
	if s.Name != "TODO" || s.Color != "50" || !s.Done {
		t.Fatalf("got %+v, want {TODO 50 true}", s)
	}
	if len(m.config.States.States) != 1 {
		t.Fatalf("edit must not add a state, got %d", len(m.config.States.States))
	}
}

func TestApplyStateFormEmptyNameRejected(t *testing.T) {
	m := newTestModel(nil)
	m.stateForm = stateForm{editing: false, name: inputWith("   "), color: inputWith("202")}

	if m.applyStateForm() {
		t.Fatalf("expected commit to fail on empty name")
	}
	if len(m.config.States.States) != 0 {
		t.Fatalf("expected no state added, got %d", len(m.config.States.States))
	}
}

func TestMoveStateFormFieldClampsForEdit(t *testing.T) {
	m := newTestModel([]config.StateConfig{{Name: "TODO", Color: "202"}})
	m.startStateForm(true, 0)

	if m.stateForm.field != 1 {
		t.Fatalf("edit form should start on color (field 1), got %d", m.stateForm.field)
	}
	m.moveStateFormField(-5)
	if m.stateForm.field != 1 {
		t.Fatalf("edit form must not move above color, got %d", m.stateForm.field)
	}
	m.moveStateFormField(5)
	if m.stateForm.field != 2 {
		t.Fatalf("field should clamp to 2 (done), got %d", m.stateForm.field)
	}
}

func TestMoveStateFormFieldAddStartsAtName(t *testing.T) {
	m := newTestModel(nil)
	m.startStateForm(false, 0)

	if m.stateForm.field != 0 {
		t.Fatalf("add form should start on name (field 0), got %d", m.stateForm.field)
	}
	m.moveStateFormField(-5)
	if m.stateForm.field != 0 {
		t.Fatalf("add form must not move above name, got %d", m.stateForm.field)
	}
}
