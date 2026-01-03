package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/rwejlgaard/org/internal/config"
)

type keyMap struct {
	Up            key.Binding
	Down          key.Binding
	Left          key.Binding
	Right         key.Binding
	ShiftUp       key.Binding
	ShiftDown     key.Binding
	ShiftLeft     key.Binding
	ShiftRight    key.Binding
	Rename        key.Binding
	CycleState    key.Binding
	ToggleView    key.Binding
	Quit          key.Binding
	Help          key.Binding
	Capture       key.Binding
	AddSubTask    key.Binding
	Delete        key.Binding
	Save          key.Binding
	ToggleFold    key.Binding
	ToggleFoldAll key.Binding
	EditNotes     key.Binding
	ToggleReorder key.Binding
	ClockIn       key.Binding
	ClockOut      key.Binding
	SetDeadline   key.Binding
	SetScheduled  key.Binding
	SetPriority   key.Binding
	SetEffort     key.Binding
	Settings      key.Binding
	TagItem       key.Binding
}

// newKeyMapFromConfig creates a keyMap from configuration
func newKeyMapFromConfig(cfg *config.Config) keyMap {
	kb := cfg.Keybindings

	return keyMap{
		Up: key.NewBinding(
			key.WithKeys(kb.Up...),
			key.WithHelp(formatKeyHelp(kb.Up), "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys(kb.Down...),
			key.WithHelp(formatKeyHelp(kb.Down), "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys(kb.Left...),
			key.WithHelp(formatKeyHelp(kb.Left), "cycle state backward"),
		),
		Right: key.NewBinding(
			key.WithKeys(kb.Right...),
			key.WithHelp(formatKeyHelp(kb.Right), "cycle state forward"),
		),
		ShiftUp: key.NewBinding(
			key.WithKeys(kb.ShiftUp...),
			key.WithHelp(formatKeyHelp(kb.ShiftUp), "move item up"),
		),
		ShiftDown: key.NewBinding(
			key.WithKeys(kb.ShiftDown...),
			key.WithHelp(formatKeyHelp(kb.ShiftDown), "move item down"),
		),
		ShiftLeft: key.NewBinding(
			key.WithKeys(kb.ShiftLeft...),
			key.WithHelp(formatKeyHelp(kb.ShiftLeft), "promote item"),
		),
		ShiftRight: key.NewBinding(
			key.WithKeys(kb.ShiftRight...),
			key.WithHelp(formatKeyHelp(kb.ShiftRight), "demote item"),
		),
		Rename: key.NewBinding(
			key.WithKeys(kb.Rename...),
			key.WithHelp(formatKeyHelp(kb.Rename), "rename item"),
		),
		CycleState: key.NewBinding(
			key.WithKeys(kb.CycleState...),
			key.WithHelp(formatKeyHelp(kb.CycleState), "cycle todo state"),
		),
		ToggleFold: key.NewBinding(
			key.WithKeys(kb.ToggleFold...),
			key.WithHelp(formatKeyHelp(kb.ToggleFold), "fold/unfold"),
		),
		ToggleFoldAll: key.NewBinding(
			key.WithKeys(kb.ToggleFoldAll...),
			key.WithHelp(formatKeyHelp(kb.ToggleFoldAll), "fold/unfold all"),
		),
		EditNotes: key.NewBinding(
			key.WithKeys(kb.EditNotes...),
			key.WithHelp(formatKeyHelp(kb.EditNotes), "edit notes"),
		),
		ToggleView: key.NewBinding(
			key.WithKeys(kb.ToggleView...),
			key.WithHelp(formatKeyHelp(kb.ToggleView), "toggle agenda view"),
		),
		Capture: key.NewBinding(
			key.WithKeys(kb.Capture...),
			key.WithHelp(formatKeyHelp(kb.Capture), "capture TODO"),
		),
		AddSubTask: key.NewBinding(
			key.WithKeys(kb.AddSubTask...),
			key.WithHelp(formatKeyHelp(kb.AddSubTask), "add sub-task"),
		),
		Delete: key.NewBinding(
			key.WithKeys(kb.Delete...),
			key.WithHelp(formatKeyHelp(kb.Delete), "delete item"),
		),
		Save: key.NewBinding(
			key.WithKeys(kb.Save...),
			key.WithHelp(formatKeyHelp(kb.Save), "save"),
		),
		ToggleReorder: key.NewBinding(
			key.WithKeys(kb.ToggleReorder...),
			key.WithHelp(formatKeyHelp(kb.ToggleReorder), "reorder mode"),
		),
		ClockIn: key.NewBinding(
			key.WithKeys(kb.ClockIn...),
			key.WithHelp(formatKeyHelp(kb.ClockIn), "clock in"),
		),
		ClockOut: key.NewBinding(
			key.WithKeys(kb.ClockOut...),
			key.WithHelp(formatKeyHelp(kb.ClockOut), "clock out"),
		),
		SetDeadline: key.NewBinding(
			key.WithKeys(kb.SetDeadline...),
			key.WithHelp(formatKeyHelp(kb.SetDeadline), "set deadline"),
		),
		SetScheduled: key.NewBinding(
			key.WithKeys(kb.SetScheduled...),
			key.WithHelp(formatKeyHelp(kb.SetScheduled), "set scheduled"),
		),
		SetPriority: key.NewBinding(
			key.WithKeys(kb.SetPriority...),
			key.WithHelp(formatKeyHelp(kb.SetPriority), "set priority"),
		),
		SetEffort: key.NewBinding(
			key.WithKeys(kb.SetEffort...),
			key.WithHelp(formatKeyHelp(kb.SetEffort), "set effort"),
		),
		Help: key.NewBinding(
			key.WithKeys(kb.Help...),
			key.WithHelp(formatKeyHelp(kb.Help), "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys(kb.Quit...),
			key.WithHelp(formatKeyHelp(kb.Quit), "quit"),
		),
		Settings: key.NewBinding(
			key.WithKeys(kb.Settings...),
			key.WithHelp(formatKeyHelp(kb.Settings), "settings"),
		),
		TagItem: key.NewBinding(
			key.WithKeys(kb.TagItem...),
			key.WithHelp(formatKeyHelp(kb.TagItem), "add/edit tags"),
		),
	}
}

// formatKeyHelp formats a slice of keys for display in help
func formatKeyHelp(keys []string) string {
	if len(keys) == 0 {
		return ""
	}
	// Take first two keys for display
	if len(keys) == 1 {
		return formatKey(keys[0])
	}
	return formatKey(keys[0]) + "/" + formatKey(keys[1])
}

// formatKey formats a single key for display
func formatKey(k string) string {
	// Convert key names to symbols where appropriate
	k = strings.ReplaceAll(k, "up", "↑")
	k = strings.ReplaceAll(k, "down", "↓")
	k = strings.ReplaceAll(k, "left", "←")
	k = strings.ReplaceAll(k, "right", "→")
	return k
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	// This will be overridden by custom rendering in viewFullHelp
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.ToggleFold, k.ToggleFoldAll, k.EditNotes, k.ToggleReorder},
		{k.Capture, k.AddSubTask, k.Delete, k.Save},
		{k.ToggleView, k.Help, k.Quit},
	}
}

// getAllBindings returns all keybindings as a flat list
func (k keyMap) getAllBindings() []key.Binding {
	return []key.Binding{
		k.Up, k.Down, k.Left, k.Right,
		k.ToggleFold, k.ToggleFoldAll, k.EditNotes, k.ToggleReorder,
		k.Capture, k.AddSubTask, k.Delete, k.Save,
		k.ClockIn, k.ClockOut, k.SetDeadline, k.SetScheduled, k.SetPriority, k.SetEffort,
		k.TagItem, k.Settings, k.ToggleView, k.Help, k.Quit,
	}
}
