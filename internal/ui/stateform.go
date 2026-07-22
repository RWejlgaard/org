package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
)

// stateForm backs the add/edit state form (modeSettingsStateForm).
// field: 0 = name, 1 = color, 2 = done.
type stateForm struct {
	editing bool
	index   int // index into config.States.States when editing
	name    textinput.Model
	color   textinput.Model
	done    bool
	field   int
}

// startStateForm initializes the form and switches into the form mode.
// editing=false adds a new state; editing=true edits config.States.States[index].
func (m *uiModel) startStateForm(editing bool, index int) {
	name := textinput.New()
	name.Placeholder = "STATE NAME"
	name.CharLimit = 32
	name.Width = 30

	color := textinput.New()
	color.Placeholder = "0-255"
	color.CharLimit = 3
	color.Width = 12

	sf := stateForm{editing: editing, index: index, name: name, color: color}
	if editing {
		st := m.config.States.States[index]
		sf.name.SetValue(st.Name)
		sf.color.SetValue(st.Color)
		sf.done = st.Done
		sf.field = 1 // name is read-only when editing
	} else {
		sf.color.SetValue("99")
		sf.field = 0
	}

	m.stateForm = sf
	m.focusStateFormField()
	m.mode = modeSettingsStateForm
}

// stateFormMinField is the lowest navigable field: name (0) when adding,
// color (1) when editing (name is read-only).
func (m *uiModel) stateFormMinField() int {
	if m.stateForm.editing {
		return 1
	}
	return 0
}

// moveStateFormField shifts the focused field by delta, clamped to [min, 2].
func (m *uiModel) moveStateFormField(delta int) {
	f := m.stateForm.field + delta
	if min := m.stateFormMinField(); f < min {
		f = min
	}
	if f > 2 {
		f = 2
	}
	m.stateForm.field = f
	m.focusStateFormField()
}

// focusStateFormField focuses the active text field (none for the Done field).
func (m *uiModel) focusStateFormField() {
	m.stateForm.name.Blur()
	m.stateForm.color.Blur()
	switch m.stateForm.field {
	case 0:
		m.stateForm.name.Focus()
	case 1:
		m.stateForm.color.Focus()
	}
}

// applyStateForm commits the form to config. It mutates config only (no Save).
// Returns false (without mutating) if a new state's name is empty.
func (m *uiModel) applyStateForm() bool {
	sf := &m.stateForm
	color := strings.TrimSpace(sf.color.Value())

	if sf.editing {
		st := &m.config.States.States[sf.index]
		if color != "" {
			st.Color = color
		}
		st.Done = sf.done
		m.setStatus(fmt.Sprintf("Updated state '%s' (saved)", st.Name))
		return true
	}

	name := strings.ToUpper(strings.TrimSpace(sf.name.Value()))
	if name == "" {
		m.setStatus("State name required")
		return false
	}
	if color == "" {
		color = "99"
	}
	if sf.done {
		m.config.AddDoneState(name, color)
	} else {
		m.config.AddState(name, color)
	}
	m.setStatus(fmt.Sprintf("Added state '%s' (saved)", name))
	return true
}
