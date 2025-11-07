package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewMode int

const (
	modeList viewMode = iota
	modeAgenda
	modeEdit
	modeConfirmDelete
	modeCapture
	modeAddSubTask
	modeSetDeadline
)

type model struct {
	orgFile         *OrgFile
	cursor          int
	mode            viewMode
	help            help.Model
	keys            keyMap
	width           int
	height          int
	statusMsg       string
	statusExpiry    time.Time
	editingItem     *Item
	textarea        textarea.Model
	textinput       textinput.Model
	itemToDelete    *Item
	reorderMode     bool
}

type keyMap struct {
	Up              key.Binding
	Down            key.Binding
	Left            key.Binding
	Right           key.Binding
	ShiftUp         key.Binding
	ShiftDown       key.Binding
	CycleState      key.Binding
	ToggleView      key.Binding
	Quit            key.Binding
	Help            key.Binding
	Capture         key.Binding
	AddSubTask      key.Binding
	Delete          key.Binding
	Save            key.Binding
	ToggleFold      key.Binding
	EditNotes       key.Binding
	ToggleReorder   key.Binding
	ClockIn         key.Binding
	ClockOut        key.Binding
	SetDeadline     key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "cycle state backward"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "cycle state forward"),
	),
	ShiftUp: key.NewBinding(
		key.WithKeys("shift+up"),
		key.WithHelp("shift+↑", "move item up"),
	),
	ShiftDown: key.NewBinding(
		key.WithKeys("shift+down"),
		key.WithHelp("shift+↓", "move item down"),
	),
	CycleState: key.NewBinding(
		key.WithKeys("t", " "),
		key.WithHelp("t/space", "cycle todo state"),
	),
	ToggleFold: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "fold/unfold"),
	),
	EditNotes: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "edit notes"),
	),
	ToggleView: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "toggle agenda view"),
	),
	Capture: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "capture TODO"),
	),
	AddSubTask: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "add sub-task"),
	),
	Delete: key.NewBinding(
		key.WithKeys("shift+d"),
		key.WithHelp("shift+d", "delete item"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	),
	ToggleReorder: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reorder mode"),
	),
	ClockIn: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "clock in"),
	),
	ClockOut: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "clock out"),
	),
	SetDeadline: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "set deadline"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	// This will be overridden by custom rendering in viewFullHelp
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.ToggleFold, k.EditNotes, k.ToggleReorder},
		{k.Capture, k.AddSubTask, k.Delete, k.Save},
		{k.ToggleView, k.Help, k.Quit},
	}
}

// getAllBindings returns all keybindings as a flat list
func (k keyMap) getAllBindings() []key.Binding {
	return []key.Binding{
		k.Up, k.Down, k.Left, k.Right,
		k.ToggleFold, k.EditNotes, k.ToggleReorder,
		k.Capture, k.AddSubTask, k.Delete, k.Save,
		k.ClockIn, k.ClockOut, k.SetDeadline,
		k.ToggleView, k.Help, k.Quit,
	}
}

// Styles
var (
	todoStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("202")) // Orange
	progStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("220")) // Yellow
	blockStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
	doneStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))  // Green
	cursorStyle     = lipgloss.NewStyle().Background(lipgloss.Color("240"))
	titleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	scheduledStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("141")) // Purple
	overdueStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
	statusStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	noteStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Italic(true)
	foldedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
)

func initialModel(orgFile *OrgFile) model {
	ta := textarea.New()
	ta.Placeholder = "Enter notes here (code blocks supported)..."
	ta.ShowLineNumbers = false

	ti := textinput.New()
	ti.Placeholder = "What needs doing?"
	ti.CharLimit = 200

	h := help.New()
	h.ShowAll = false

	return model{
		orgFile:   orgFile,
		cursor:    0,
		mode:      modeList,
		help:      h,
		keys:      keys,
		textarea:  ta,
		textinput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle special modes
	switch m.mode {
	case modeEdit:
		return m.updateEditMode(msg)
	case modeConfirmDelete:
		return m.updateConfirmDelete(msg)
	case modeCapture:
		return m.updateCapture(msg)
	case modeAddSubTask:
		return m.updateAddSubTask(msg)
	case modeSetDeadline:
		return m.updateSetDeadline(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.textarea.SetWidth(msg.Width - 4)
		m.textarea.SetHeight(msg.Height - 10)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil

		case key.Matches(msg, m.keys.Up):
			if m.reorderMode {
				m.moveItemUp()
			} else {
				if m.cursor > 0 {
					m.cursor--
				}
			}

		case key.Matches(msg, m.keys.Down):
			if m.reorderMode {
				m.moveItemDown()
			} else {
				items := m.getVisibleItems()
				if m.cursor < len(items)-1 {
					m.cursor++
				}
			}

		case key.Matches(msg, m.keys.Left):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.cycleStateBackward(items[m.cursor])
				// Auto clock out when changing to DONE
				if items[m.cursor].State == StateDONE && items[m.cursor].IsClockedIn() {
					items[m.cursor].ClockOut()
				}
				m.setStatus("State changed")
			}

		case key.Matches(msg, m.keys.Right):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				items[m.cursor].CycleState()
				// Auto clock out when changing to DONE
				if items[m.cursor].State == StateDONE && items[m.cursor].IsClockedIn() {
					items[m.cursor].ClockOut()
				}
				m.setStatus("State changed")
			}

		case key.Matches(msg, m.keys.ShiftUp):
			m.moveItemUp()

		case key.Matches(msg, m.keys.ShiftDown):
			m.moveItemDown()

		case key.Matches(msg, m.keys.CycleState):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				items[m.cursor].CycleState()
				// Auto clock out when changing to DONE
				if items[m.cursor].State == StateDONE && items[m.cursor].IsClockedIn() {
					items[m.cursor].ClockOut()
				}
				m.setStatus("State changed")
			}

		case key.Matches(msg, m.keys.ToggleFold):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				items[m.cursor].ToggleFold()
				if items[m.cursor].Folded {
					m.setStatus("Folded")
				} else {
					m.setStatus("Unfolded")
				}
			}

		case key.Matches(msg, m.keys.EditNotes):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.editingItem = items[m.cursor]
				m.mode = modeEdit
				m.textarea.SetValue(strings.Join(m.editingItem.Notes, "\n"))
				m.textarea.Focus()
				return m, textarea.Blink
			}

		case key.Matches(msg, m.keys.Capture):
			m.mode = modeCapture
			m.textinput.SetValue("")
			m.textinput.Placeholder = "What needs doing?"
			m.textinput.Focus()
			return m, textinput.Blink

		case key.Matches(msg, m.keys.AddSubTask):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.editingItem = items[m.cursor]
				m.mode = modeAddSubTask
				m.textinput.SetValue("")
				m.textinput.Placeholder = "Sub-task title"
				m.textinput.Focus()
				return m, textinput.Blink
			}

		case key.Matches(msg, m.keys.Delete):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.itemToDelete = items[m.cursor]
				m.mode = modeConfirmDelete
			}

		case key.Matches(msg, m.keys.ToggleView):
			if m.mode == modeList {
				m.mode = modeAgenda
			} else {
				m.mode = modeList
			}
			m.cursor = 0

		case key.Matches(msg, m.keys.Save):
			if err := m.orgFile.Save(); err != nil {
				m.setStatus(fmt.Sprintf("Error saving: %v", err))
			} else {
				m.setStatus("Saved!")
			}

		case key.Matches(msg, m.keys.ToggleReorder):
			m.reorderMode = !m.reorderMode
			if m.reorderMode {
				m.setStatus("Reorder mode ON - Use ↑/↓ to move items, 'r' to exit")
			} else {
				m.setStatus("Reorder mode OFF")
			}

		case key.Matches(msg, m.keys.ClockIn):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				if items[m.cursor].ClockIn() {
					m.setStatus("Clocked in!")
				} else {
					m.setStatus("Already clocked in")
				}
			}

		case key.Matches(msg, m.keys.ClockOut):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				if items[m.cursor].ClockOut() {
					m.setStatus("Clocked out!")
				} else {
					m.setStatus("Not clocked in")
				}
			}

		case key.Matches(msg, m.keys.SetDeadline):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.editingItem = items[m.cursor]
				m.mode = modeSetDeadline
				m.textinput.SetValue("")
				m.textinput.Placeholder = "YYYY-MM-DD or +N (days from today)"
				m.textinput.Focus()
				return m, textinput.Blink
			}
		}
	}

	return m, nil
}

func (m model) updateEditMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width - 4)
		m.textarea.SetHeight(msg.Height - 10)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			// Save notes and exit edit mode
			if m.editingItem != nil {
				noteText := m.textarea.Value()
				if noteText == "" {
					m.editingItem.Notes = []string{}
				} else {
					m.editingItem.Notes = strings.Split(noteText, "\n")
				}
			}
			m.mode = modeList
			m.textarea.Blur()
			m.setStatus("Notes saved")
			return m, nil
		}
	}

	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m model) updateConfirmDelete(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			// Delete the item
			m.deleteItem(m.itemToDelete)
			m.mode = modeList
			m.itemToDelete = nil
			m.setStatus("Item deleted")
			// Adjust cursor if needed
			items := m.getVisibleItems()
			if m.cursor >= len(items) && len(items) > 0 {
				m.cursor = len(items) - 1
			}
		case "n", "N", "esc":
			m.mode = modeList
			m.itemToDelete = nil
			m.setStatus("Cancelled")
		}
	}
	return m, nil
}

func (m model) updateCapture(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			title := strings.TrimSpace(m.textinput.Value())
			if title != "" {
				// Create new TODO at top level
				newItem := &Item{
					Level:    1,
					State:    StateTODO,
					Title:    title,
					Notes:    []string{},
					Children: []*Item{},
				}
				// Insert at beginning
				m.orgFile.Items = append([]*Item{newItem}, m.orgFile.Items...)
				m.setStatus("TODO captured!")
			}
			m.mode = modeList
			m.textinput.Blur()
			m.cursor = 0
			return m, nil
		case tea.KeyEsc:
			m.mode = modeList
			m.textinput.Blur()
			m.setStatus("Cancelled")
			return m, nil
		}
	}

	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m model) updateAddSubTask(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			title := strings.TrimSpace(m.textinput.Value())
			if title != "" && m.editingItem != nil {
				// Create new sub-task
				newItem := &Item{
					Level:    m.editingItem.Level + 1,
					State:    StateTODO,
					Title:    title,
					Notes:    []string{},
					Children: []*Item{},
				}
				m.editingItem.Children = append(m.editingItem.Children, newItem)
				m.editingItem.Folded = false // Unfold to show new sub-task
				m.setStatus("Sub-task added!")
			}
			m.mode = modeList
			m.textinput.Blur()
			m.editingItem = nil
			return m, nil
		case tea.KeyEsc:
			m.mode = modeList
			m.textinput.Blur()
			m.editingItem = nil
			m.setStatus("Cancelled")
			return m, nil
		}
	}

	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m model) updateSetDeadline(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			input := strings.TrimSpace(m.textinput.Value())
			if m.editingItem != nil {
				if input == "" {
					// Empty input clears the deadline
					m.editingItem.Deadline = nil
					// Remove DEADLINE line from notes (only lines starting with DEADLINE:)
					var filteredNotes []string
					for _, note := range m.editingItem.Notes {
						trimmedNote := strings.TrimSpace(note)
						if !strings.HasPrefix(trimmedNote, "DEADLINE:") {
							filteredNotes = append(filteredNotes, note)
						}
					}
					m.editingItem.Notes = filteredNotes
					m.setStatus("Deadline cleared!")
				} else {
					deadline, err := parseDeadlineInput(input)
					if err != nil {
						m.setStatus(fmt.Sprintf("Invalid date: %v", err))
					} else {
						m.editingItem.Deadline = &deadline
						// Also update or add DEADLINE line in notes
						updatedNotes := false
						for i, note := range m.editingItem.Notes {
							trimmedNote := strings.TrimSpace(note)
							if strings.HasPrefix(trimmedNote, "DEADLINE:") {
								m.editingItem.Notes[i] = fmt.Sprintf("DEADLINE: <%s>", formatOrgDate(deadline))
								updatedNotes = true
								break
							}
						}
						// If DEADLINE wasn't in notes, it will be added by writeItem
						if !updatedNotes {
							// Remove old deadline lines just to be safe
							var filteredNotes []string
							for _, note := range m.editingItem.Notes {
								trimmedNote := strings.TrimSpace(note)
								if !strings.HasPrefix(trimmedNote, "DEADLINE:") {
									filteredNotes = append(filteredNotes, note)
								}
							}
							m.editingItem.Notes = filteredNotes
						}
						m.setStatus("Deadline set!")
					}
				}
			}
			m.mode = modeList
			m.textinput.Blur()
			m.editingItem = nil
			return m, nil
		case tea.KeyEsc:
			m.mode = modeList
			m.textinput.Blur()
			m.editingItem = nil
			m.setStatus("Cancelled")
			return m, nil
		}
	}

	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

// parseDeadlineInput parses deadline input like "2024-01-15" or "+3" (3 days from now)
func parseDeadlineInput(input string) (time.Time, error) {
	// Check if it's a relative date (+N days)
	if strings.HasPrefix(input, "+") {
		daysStr := strings.TrimPrefix(input, "+")
		days := 0
		_, err := fmt.Sscanf(daysStr, "%d", &days)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid relative date format: %s", input)
		}
		return time.Now().AddDate(0, 0, days), nil
	}

	// Try parsing as absolute date
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, input); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s (use YYYY-MM-DD or +N)", input)
}

func (m *model) cycleStateBackward(item *Item) {
	switch item.State {
	case StateNone:
		item.State = StateDONE
	case StateTODO:
		item.State = StateNone
	case StatePROG:
		item.State = StateTODO
	case StateBLOCK:
		item.State = StatePROG
	case StateDONE:
		item.State = StateBLOCK
	}
}

func (m *model) deleteItem(item *Item) {
	var removeFromList func([]*Item, *Item) []*Item
	removeFromList = func(items []*Item, target *Item) []*Item {
		result := []*Item{}
		for _, it := range items {
			if it == target {
				continue
			}
			it.Children = removeFromList(it.Children, target)
			result = append(result, it)
		}
		return result
	}
	m.orgFile.Items = removeFromList(m.orgFile.Items, item)
}

func (m *model) moveItemUp() {
	items := m.getVisibleItems()
	if len(items) == 0 || m.cursor == 0 {
		return
	}

	currentItem := items[m.cursor]
	prevItem := items[m.cursor-1]

	// Can only swap items at the same level
	if currentItem.Level != prevItem.Level {
		m.setStatus("Cannot move across different levels")
		return
	}

	m.swapItems(currentItem, prevItem)
	m.cursor--
	m.setStatus("Item moved up")
}

func (m *model) moveItemDown() {
	items := m.getVisibleItems()
	if len(items) == 0 || m.cursor >= len(items)-1 {
		return
	}

	currentItem := items[m.cursor]
	nextItem := items[m.cursor+1]

	// Can only swap items at the same level
	if currentItem.Level != nextItem.Level {
		m.setStatus("Cannot move across different levels")
		return
	}

	m.swapItems(currentItem, nextItem)
	m.cursor++
	m.setStatus("Item moved down")
}

func (m *model) swapItems(item1, item2 *Item) {
	// Find parent list containing both items
	var swapInList func([]*Item) bool
	swapInList = func(items []*Item) bool {
		for i := 0; i < len(items)-1; i++ {
			if items[i] == item1 && items[i+1] == item2 {
				items[i], items[i+1] = items[i+1], items[i]
				return true
			}
			if items[i] == item2 && items[i+1] == item1 {
				items[i], items[i+1] = items[i+1], items[i]
				return true
			}
			if swapInList(items[i].Children) {
				return true
			}
		}
		if len(items) > 0 && swapInList(items[len(items)-1].Children) {
			return true
		}
		return false
	}
	swapInList(m.orgFile.Items)
}

func (m *model) setStatus(msg string) {
	m.statusMsg = msg
	m.statusExpiry = time.Now().Add(3 * time.Second)
}

// dynamicKeyMap is a helper type for rendering keybindings with dynamic layout
type dynamicKeyMap struct {
	rows [][]key.Binding
}

// ShortHelp for dynamicKeyMap
func (d dynamicKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

// FullHelp for dynamicKeyMap
func (d dynamicKeyMap) FullHelp() [][]key.Binding {
	return d.rows
}

// renderFullHelp renders the help with width-aware layout
func (m model) renderFullHelp() string {
	bindings := m.keys.getAllBindings()

	// Estimate the width needed for each keybinding (key + desc + padding)
	// Average is roughly 20-25 chars per binding
	const estimatedBindingWidth = 22
	const minWidth = 40 // Minimum width before stacking

	var columnsPerRow int
	if m.width < minWidth {
		columnsPerRow = 1 // Stack vertically on very narrow terminals
	} else if m.width < 80 {
		columnsPerRow = 2 // Two columns on narrow terminals
	} else if m.width < 120 {
		columnsPerRow = 3 // Three columns on medium terminals
	} else {
		columnsPerRow = 4 // Four columns on wide terminals
	}

	// Build rows based on columns per row
	var rows [][]key.Binding
	for i := 0; i < len(bindings); i += columnsPerRow {
		end := i + columnsPerRow
		if end > len(bindings) {
			end = len(bindings)
		}
		rows = append(rows, bindings[i:end])
	}

	// Use the help model to render with our dynamic layout
	h := help.New()
	h.Width = m.width
	h.ShowAll = true

	// Create a temporary keyMap for rendering
	dkm := dynamicKeyMap{rows: rows}

	return h.View(dkm)
}

func (m model) getVisibleItems() []*Item {
	if m.mode == modeAgenda {
		return m.getAgendaItems()
	}
	return m.orgFile.GetAllItems()
}

func (m model) getAgendaItems() []*Item {
	var items []*Item
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfWeek := startOfDay.AddDate(0, 0, 7)

	// Get all items regardless of folding for agenda view
	var getAllItems func([]*Item)
	getAllItems = func(list []*Item) {
		for _, item := range list {
			if item.Scheduled != nil && item.Scheduled.Before(endOfWeek) {
				items = append(items, item)
			}
			if item.Deadline != nil && item.Deadline.Before(endOfWeek) {
				items = append(items, item)
			}
			getAllItems(item.Children)
		}
	}
	getAllItems(m.orgFile.Items)

	return items
}

func (m model) View() string {
	switch m.mode {
	case modeEdit:
		return m.viewEditMode()
	case modeConfirmDelete:
		return m.viewConfirmDelete()
	case modeCapture:
		return m.viewCapture()
	case modeAddSubTask:
		return m.viewAddSubTask()
	case modeSetDeadline:
		return m.viewSetDeadline()
	}

	// Build footer (status + help)
	var footer strings.Builder

	// Status message
	if time.Now().Before(m.statusExpiry) {
		footer.WriteString(statusStyle.Render(m.statusMsg))
		footer.WriteString("\n")
	}

	// Help
	if m.help.ShowAll {
		footer.WriteString(m.renderFullHelp())
	} else {
		footer.WriteString(m.help.View(m.keys))
	}

	footerHeight := lipgloss.Height(footer.String())

	// Build main content
	var content strings.Builder

	// Title
	title := "Org Mode - List View"
	if m.mode == modeAgenda {
		title = "Org Mode - Agenda View (Next 7 Days)"
	}
	if m.reorderMode {
		reorderIndicator := lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(" [REORDER MODE]")
		content.WriteString(titleStyle.Render(title))
		content.WriteString(reorderIndicator)
	} else {
		content.WriteString(titleStyle.Render(title))
	}
	content.WriteString("\n\n")

	// Calculate available height for items (total - title - footer)
	availableHeight := m.height - 3 - footerHeight // 3 for title + spacing
	if availableHeight < 5 {
		availableHeight = 5 // Minimum height
	}

	// Items
	items := m.getVisibleItems()
	if len(items) == 0 {
		content.WriteString("No items. Press 'c' to capture a new TODO.\n")
	}

	itemLines := 0
	for i, item := range items {
		if itemLines >= availableHeight {
			break // Don't render more items than fit
		}

		line := m.renderItem(item, i == m.cursor)
		content.WriteString(line)
		content.WriteString("\n")
		itemLines++

		// Show notes if not folded
		if !item.Folded && len(item.Notes) > 0 && m.mode == modeList {
			indent := strings.Repeat("  ", item.Level)
			// Filter out LOGBOOK drawer and apply syntax highlighting to notes
			filteredNotes := filterLogbookDrawer(item.Notes)
			highlightedNotes := renderNotesWithHighlighting(filteredNotes)
			for _, note := range highlightedNotes {
				if itemLines >= availableHeight {
					break
				}
				content.WriteString(indent)
				content.WriteString("  " + note)
				content.WriteString("\n")
				itemLines++
			}
		}
	}

	// Combine content and footer with padding
	contentHeight := lipgloss.Height(content.String())
	paddingNeeded := m.height - contentHeight - footerHeight
	if paddingNeeded < 0 {
		paddingNeeded = 0
	}

	var result strings.Builder
	result.WriteString(content.String())
	if paddingNeeded > 0 {
		result.WriteString(strings.Repeat("\n", paddingNeeded))
	}
	result.WriteString(footer.String())

	return result.String()
}

func (m model) viewConfirmDelete() string {
	var b strings.Builder

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 2).
		Width(60)

	var content strings.Builder
	content.WriteString(titleStyle.Render("⚠ Delete Item"))
	content.WriteString("\n\n")

	if m.itemToDelete != nil {
		itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("202")).Bold(true)
		content.WriteString(itemStyle.Render(m.itemToDelete.Title))
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(statusStyle.Render("This will delete the item and all sub-tasks."))
	content.WriteString("\n\n")
	content.WriteString("Press Y to confirm • N or ESC to cancel")

	dialog := dialogStyle.Render(content.String())

	// Center the dialog
	if m.height > 0 {
		verticalPadding := (m.height - lipgloss.Height(dialog)) / 2
		if verticalPadding > 0 {
			b.WriteString(strings.Repeat("\n", verticalPadding))
		}
	}
	b.WriteString(dialog)

	return b.String()
}

func (m model) viewCapture() string {
	var b strings.Builder

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 2).
		Width(60)

	var content strings.Builder
	content.WriteString(titleStyle.Render("Capture TODO"))
	content.WriteString("\n\n")
	content.WriteString(m.textinput.View())
	content.WriteString("\n\n")
	content.WriteString(statusStyle.Render("Press Enter to save • ESC to cancel"))

	dialog := dialogStyle.Render(content.String())

	// Center the dialog
	if m.height > 0 {
		verticalPadding := (m.height - lipgloss.Height(dialog)) / 2
		if verticalPadding > 0 {
			b.WriteString(strings.Repeat("\n", verticalPadding))
		}
	}
	b.WriteString(dialog)

	return b.String()
}

func (m model) viewAddSubTask() string {
	var b strings.Builder

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 2).
		Width(60)

	var content strings.Builder
	content.WriteString(titleStyle.Render("Add Sub-Task"))
	content.WriteString("\n")
	if m.editingItem != nil {
		content.WriteString(statusStyle.Render(fmt.Sprintf("Under: %s", m.editingItem.Title)))
	}
	content.WriteString("\n\n")
	content.WriteString(m.textinput.View())
	content.WriteString("\n\n")
	content.WriteString(statusStyle.Render("Press Enter to save • ESC to cancel"))

	dialog := dialogStyle.Render(content.String())

	// Center the dialog
	if m.height > 0 {
		verticalPadding := (m.height - lipgloss.Height(dialog)) / 2
		if verticalPadding > 0 {
			b.WriteString(strings.Repeat("\n", verticalPadding))
		}
	}
	b.WriteString(dialog)

	return b.String()
}

func (m model) viewSetDeadline() string {
	var b strings.Builder

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("141")).
		Padding(1, 2).
		Width(60)

	var content strings.Builder
	content.WriteString(titleStyle.Render("Set Deadline"))
	content.WriteString("\n")
	if m.editingItem != nil {
		content.WriteString(statusStyle.Render(fmt.Sprintf("For: %s", m.editingItem.Title)))
	}
	content.WriteString("\n\n")
	content.WriteString(m.textinput.View())
	content.WriteString("\n\n")
	content.WriteString(statusStyle.Render("Examples: 2025-12-31, +7 (7 days from now)"))
	content.WriteString("\n")
	content.WriteString(statusStyle.Render("Leave empty to clear deadline"))
	content.WriteString("\n")
	content.WriteString(statusStyle.Render("Press Enter to save • ESC to cancel"))

	dialog := dialogStyle.Render(content.String())

	// Center the dialog
	if m.height > 0 {
		verticalPadding := (m.height - lipgloss.Height(dialog)) / 2
		if verticalPadding > 0 {
			b.WriteString(strings.Repeat("\n", verticalPadding))
		}
	}
	b.WriteString(dialog)

	return b.String()
}

func (m model) viewEditMode() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Editing Notes"))
	b.WriteString("\n")
	if m.editingItem != nil {
		b.WriteString(fmt.Sprintf("Item: %s\n", m.editingItem.Title))
	}
	b.WriteString(statusStyle.Render("Press ESC to save and exit"))
	b.WriteString("\n\n")

	b.WriteString(m.textarea.View())

	return b.String()
}

// filterLogbookDrawer removes LOGBOOK drawer content and scheduling metadata from notes
func filterLogbookDrawer(notes []string) []string {
	var filtered []string
	inLogbook := false

	for _, note := range notes {
		trimmed := strings.TrimSpace(note)

		// Check for start of LOGBOOK drawer
		if trimmed == ":LOGBOOK:" {
			inLogbook = true
			continue
		}

		// Check for end of LOGBOOK drawer
		if trimmed == ":END:" && inLogbook {
			inLogbook = false
			continue
		}

		// Skip lines inside LOGBOOK drawer
		if inLogbook {
			continue
		}

		// Skip SCHEDULED and DEADLINE lines
		if strings.HasPrefix(trimmed, "SCHEDULED:") || strings.HasPrefix(trimmed, "DEADLINE:") {
			continue
		}

		filtered = append(filtered, note)
	}

	return filtered
}

// renderNotesWithHighlighting renders notes with syntax highlighting for code blocks
func renderNotesWithHighlighting(notes []string) []string {
	if len(notes) == 0 {
		return notes
	}

	var result []string
	var inCodeBlock bool
	var codeLanguage string
	var codeLines []string
	var codeBlockDelimiter string // Track whether we're in #+BEGIN_SRC or ``` block

	for _, note := range notes {
		trimmed := strings.TrimSpace(note)

		// Check for org-mode style code block start
		if strings.HasPrefix(trimmed, "#+BEGIN_SRC") {
			inCodeBlock = true
			codeBlockDelimiter = "org"
			// Extract language
			parts := strings.Fields(trimmed)
			if len(parts) > 1 {
				codeLanguage = strings.ToLower(parts[1])
			} else {
				codeLanguage = "text"
			}
			result = append(result, note) // Keep the delimiter visible
			codeLines = []string{}
			continue
		}

		// Check for markdown style code block start
		if strings.HasPrefix(trimmed, "```") {
			if !inCodeBlock {
				// Starting a code block
				inCodeBlock = true
				codeBlockDelimiter = "markdown"
				// Extract language
				lang := strings.TrimPrefix(trimmed, "```")
				if lang != "" {
					codeLanguage = strings.ToLower(lang)
				} else {
					codeLanguage = "text"
				}
				result = append(result, note) // Keep the delimiter visible
				codeLines = []string{}
				continue
			} else if codeBlockDelimiter == "markdown" {
				// Ending a markdown code block
				inCodeBlock = false
				// Highlight and add the code
				if len(codeLines) > 0 {
					highlighted := highlightCode(strings.Join(codeLines, "\n"), codeLanguage)
					highlightedLines := strings.Split(highlighted, "\n")
					result = append(result, highlightedLines...)
				}
				result = append(result, note) // Keep the delimiter visible
				codeLines = []string{}
				codeLanguage = ""
				codeBlockDelimiter = ""
				continue
			}
		}

		// Check for org-mode style code block end
		if strings.HasPrefix(trimmed, "#+END_SRC") {
			inCodeBlock = false
			// Highlight and add the code
			if len(codeLines) > 0 {
				highlighted := highlightCode(strings.Join(codeLines, "\n"), codeLanguage)
				highlightedLines := strings.Split(highlighted, "\n")
				result = append(result, highlightedLines...)
			}
			result = append(result, note) // Keep the delimiter visible
			codeLines = []string{}
			codeLanguage = ""
			codeBlockDelimiter = ""
			continue
		}

		// If in code block, accumulate lines
		if inCodeBlock {
			codeLines = append(codeLines, note)
		} else {
			result = append(result, note)
		}
	}

	// Handle case where code block wasn't closed
	if inCodeBlock && len(codeLines) > 0 {
		highlighted := highlightCode(strings.Join(codeLines, "\n"), codeLanguage)
		highlightedLines := strings.Split(highlighted, "\n")
		result = append(result, highlightedLines...)
	}

	return result
}

// highlightCode applies syntax highlighting to code
func highlightCode(code, language string) string {
	if code == "" {
		return code
	}

	var buf strings.Builder
	err := quick.Highlight(&buf, code, language, "terminal256", "monokai")
	if err != nil {
		// If highlighting fails, return the original code
		return code
	}

	return strings.TrimRight(buf.String(), "\n")
}

func (m model) renderItem(item *Item, isCursor bool) string {
	var b strings.Builder

	// Indentation for level
	indent := strings.Repeat("  ", item.Level-1)
	b.WriteString(indent)

	// Fold indicator
	if len(item.Children) > 0 || len(item.Notes) > 0 {
		if item.Folded {
			b.WriteString(foldedStyle.Render("▶ "))
		} else {
			b.WriteString(foldedStyle.Render("▼ "))
		}
	} else {
		b.WriteString("  ")
	}

	// State
	stateStr := ""
	switch item.State {
	case StateTODO:
		stateStr = todoStyle.Render("[TODO] ")
	case StatePROG:
		stateStr = progStyle.Render("[PROG] ")
	case StateBLOCK:
		stateStr = blockStyle.Render("[BLOCK]")
	case StateDONE:
		stateStr = doneStyle.Render("[DONE] ")
	default:
		stateStr = "       " // Empty space for alignment
	}
	b.WriteString(stateStr)
	b.WriteString(" ")

	// Title
	b.WriteString(item.Title)

	// Clock status
	if item.IsClockedIn() {
		duration := item.GetCurrentClockDuration()
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		clockStr := fmt.Sprintf(" [CLOCKED IN: %dh %dm]", hours, minutes)
		clockStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true) // Bright green
		b.WriteString(clockStyle.Render(clockStr))
	}

	// Total clocked time (show if there are any clock entries)
	if len(item.ClockEntries) > 0 {
		totalDuration := item.GetTotalClockDuration()
		totalHours := int(totalDuration.Hours())
		totalMinutes := int(totalDuration.Minutes()) % 60

		// Format the time display based on magnitude
		var timeStr string
		if totalHours > 0 {
			timeStr = fmt.Sprintf("%dh %dm", totalHours, totalMinutes)
		} else {
			timeStr = fmt.Sprintf("%dm", totalMinutes)
		}

		totalTimeStr := fmt.Sprintf(" (Time: %s)", timeStr)
		totalTimeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141")) // Purple, similar to scheduled
		b.WriteString(totalTimeStyle.Render(totalTimeStr))
	}

	// Scheduling info
	now := time.Now()
	if item.Scheduled != nil {
		schedStr := fmt.Sprintf(" (Scheduled: %s)", formatOrgDate(*item.Scheduled))
		if item.Scheduled.Before(now) {
			b.WriteString(overdueStyle.Render(schedStr))
		} else {
			b.WriteString(scheduledStyle.Render(schedStr))
		}
	}
	if item.Deadline != nil {
		deadlineStr := fmt.Sprintf(" (Deadline: %s)", formatOrgDate(*item.Deadline))
		if item.Deadline.Before(now) {
			b.WriteString(overdueStyle.Render(deadlineStr))
		} else {
			b.WriteString(scheduledStyle.Render(deadlineStr))
		}
	}

	line := b.String()
	if isCursor {
		return cursorStyle.Render(line)
	}
	return line
}

func runUI(orgFile *OrgFile) error {
	p := tea.NewProgram(initialModel(orgFile), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
