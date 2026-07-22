package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	m.config.AddState(name, color, sf.done)
	m.setStatus(fmt.Sprintf("Added state '%s' (saved)", name))
	return true
}

// updateSettingsStateForm handles key input while the state form is open.
func (m *uiModel) updateSettingsStateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.Type {
	case tea.KeyEsc:
		m.mode = modeSettings
		return m, nil

	case tea.KeyEnter:
		if m.applyStateForm() {
			if err := m.config.Save(); err != nil {
				m.setStatus(fmt.Sprintf("Error auto-saving config: %v", err))
			} else {
				m.keys = newKeyMapFromConfig(m.config)
				m.styles = newStyleMapFromConfig(m.config)
			}
			m.mode = modeSettings
		}
		return m, nil

	case tea.KeyTab, tea.KeyDown:
		m.moveStateFormField(1)
		return m, nil

	case tea.KeyShiftTab, tea.KeyUp:
		m.moveStateFormField(-1)
		return m, nil

	case tea.KeySpace:
		if m.stateForm.field == 2 {
			m.stateForm.done = !m.stateForm.done
			return m, nil
		}
	}

	// Route printable input to the focused text field.
	var cmd tea.Cmd
	switch m.stateForm.field {
	case 0:
		m.stateForm.name, cmd = m.stateForm.name.Update(msg)
	case 1:
		m.stateForm.color, cmd = m.stateForm.color.Update(msg)
	}
	return m, cmd
}

// viewSettingsStateForm renders the add/edit state form.
func (m *uiModel) viewSettingsStateForm() string {
	var b strings.Builder

	title := "Add New State"
	if m.stateForm.editing {
		title = "Edit State: " + m.config.States.States[m.stateForm.index].Name
	}
	b.WriteString(m.styles.titleStyle.Render(title) + "\n\n")

	cursor := func(field int) string {
		if m.stateForm.field == field {
			return "▶ "
		}
		return "  "
	}

	// Name row: editable input when adding, read-only label when editing.
	if m.stateForm.editing {
		b.WriteString("  Name   " + m.stateForm.name.Value() + m.styles.statusStyle.Render("  (fixed)") + "\n")
	} else {
		b.WriteString(cursor(0) + "Name   " + m.stateForm.name.View() + "\n")
	}

	// Color row with a live swatch.
	colorVal := strings.TrimSpace(m.stateForm.color.Value())
	swatch := ""
	if colorVal != "" {
		label := m.stateForm.name.Value()
		if label == "" {
			label = "sample"
		}
		sw := lipgloss.NewStyle().Foreground(lipgloss.Color(colorVal))
		swatch = "  " + sw.Render("●"+label)
	}
	b.WriteString(cursor(1) + "Color  " + m.stateForm.color.View() + swatch + "\n")

	// Done row: a checkbox.
	box := "[ ] no"
	if m.stateForm.done {
		box = "[✓] yes"
	}
	b.WriteString(cursor(2) + "Done   " + box + "\n\n")

	b.WriteString(m.styles.statusStyle.Render(
		"tab/↑↓: move field • space: toggle done • enter: save • esc: cancel") + "\n")

	return b.String()
}
