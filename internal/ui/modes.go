package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rwejlgaard/org/internal/model"
	"github.com/rwejlgaard/org/internal/parser"
)

func (m uiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	case modeSetPriority:
		return m.updateSetPriority(msg)
	case modeSetEffort:
		return m.updateSetEffort(msg)
	case modeHelp:
		return m.updateHelp(msg)
	case modeSettings:
		return m.updateSettings(msg)
	case modeSettingsAddTag:
		return m.updateSettingsAddTag(msg)
	case modeSettingsAddState:
		return m.updateSettingsAddState(msg)
	case modeTagEdit:
		return m.updateTagEdit(msg)
	case modeRename:
		return m.updateRename(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.textarea.SetWidth(msg.Width - 4)
		m.textarea.SetHeight(msg.Height - 10)
		m.textinput.Width = 50 // Set a reasonable width for the text input
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.mode = modeHelp
			m.helpScroll = 0 // Reset scroll when entering help
			return m, nil

		case key.Matches(msg, m.keys.Up):
			if m.reorderMode {
				m.moveItemUp()
			} else {
				if m.cursor > 0 {
					m.cursor--
					// Update scroll to keep cursor visible
					availableHeight := m.height - 6 // Approximate
					if availableHeight < 5 {
						availableHeight = 5
					}
					m.updateScrollOffset(availableHeight)
				}
			}

		case key.Matches(msg, m.keys.Down):
			if m.reorderMode {
				m.moveItemDown()
			} else {
				items := m.getVisibleItems()
				if m.cursor < len(items)-1 {
					m.cursor++
					// Update scroll to keep cursor visible
					availableHeight := m.height - 6 // Approximate
					if availableHeight < 5 {
						availableHeight = 5
					}
					m.updateScrollOffset(availableHeight)
				}
			}

		case key.Matches(msg, m.keys.Left):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.cycleStateBackward(items[m.cursor])
				// Auto clock out when changing to DONE
				if items[m.cursor].State == model.StateDONE && items[m.cursor].IsClockedIn() {
					items[m.cursor].ClockOut()
				}
				m.setStatus("State changed")
			}

		case key.Matches(msg, m.keys.Right):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.cycleStateForward(items[m.cursor])
				// Auto clock out when changing to last state (typically DONE)
				stateNames := m.config.GetStateNames()
				if len(stateNames) > 0 && string(items[m.cursor].State) == stateNames[len(stateNames)-1] && items[m.cursor].IsClockedIn() {
					items[m.cursor].ClockOut()
				}
				m.setStatus("State changed")
			}

		case key.Matches(msg, m.keys.ShiftUp):
			m.moveItemUp()

		case key.Matches(msg, m.keys.ShiftDown):
			m.moveItemDown()

		case key.Matches(msg, m.keys.ShiftLeft):
			m.promoteItem()

		case key.Matches(msg, m.keys.ShiftRight):
			m.demoteItem()

		case key.Matches(msg, m.keys.CycleState):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.cycleStateForward(items[m.cursor])
				// Auto clock out when changing to last state (typically DONE)
				stateNames := m.config.GetStateNames()
				if len(stateNames) > 0 && string(items[m.cursor].State) == stateNames[len(stateNames)-1] && items[m.cursor].IsClockedIn() {
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
				selectedItem := items[m.cursor]

				// Check if we're in multi-file mode
				isMultiFile := len(m.orgFile.Items) > 0 && m.orgFile.Items[0].SourceFile != ""

				// Prevent editing notes for top-level file items in multi-file mode
				if isMultiFile && selectedItem.Level == 1 && selectedItem.SourceFile != "" {
					m.setStatus("Cannot add notes to file-level items")
					return m, nil
				}

				m.editingItem = selectedItem
				m.mode = modeEdit
				m.textarea.SetValue(strings.Join(m.editingItem.Notes, "\n"))
				m.textarea.Focus()
				return m, textarea.Blink
			}

		case key.Matches(msg, m.keys.Settings):
			m.mode = modeSettings
			m.initSettings()
			return m, nil

		case key.Matches(msg, m.keys.TagItem):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.editingItem = items[m.cursor]
				m.mode = modeTagEdit
				m.textinput.SetValue(strings.Join(items[m.cursor].Tags, ":"))
				m.textinput.Placeholder = "tag1:tag2:tag3"
				m.textinput.Focus()
				return m, textinput.Blink
			}

		case key.Matches(msg, m.keys.Rename):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.editingItem = items[m.cursor]
				m.mode = modeRename
				m.textinput.SetValue(items[m.cursor].Title)
				m.textinput.Placeholder = "Item title"
				m.textinput.Focus()
				return m, textinput.Blink
			}

		case key.Matches(msg, m.keys.Capture):
			m.mode = modeCapture
			m.captureCursor = m.cursor // Store current cursor position
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
			if err := parser.Save(m.orgFile); err != nil {
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

		case key.Matches(msg, m.keys.SetPriority):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.editingItem = items[m.cursor]
				m.mode = modeSetPriority
				return m, nil
			}

		case key.Matches(msg, m.keys.SetEffort):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.editingItem = items[m.cursor]
				m.mode = modeSetEffort
				m.textinput.SetValue("")
				m.textinput.Placeholder = "e.g., 8h, 2d, 1w"
				m.textinput.Focus()
				return m, textinput.Blink
			}
		}
	}

	return m, nil
}

func (m uiModel) updateEditMode(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m uiModel) updateConfirmDelete(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m uiModel) updateCapture(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textinput.Width = 50

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			title := strings.TrimSpace(m.textinput.Value())
			if title != "" {
				// Get default state from config
				defaultState := model.TodoState(m.config.GetDefaultNewTaskState())

				// Create new TODO at top level
				newItem := &model.Item{
					Level:    1,
					State:    defaultState,
					Title:    title,
					Notes:    []string{},
					Children: []*model.Item{},
				}

				// Check if we're in multi-file mode
				isMultiFile := len(m.orgFile.Items) > 0 && m.orgFile.Items[0].SourceFile != ""

				if isMultiFile {
					// In multi-file mode, add to the file of the highlighted item (using stored cursor position)
					items := m.getVisibleItems()
					targetFileItem := m.findTopLevelFileItem(items, m.captureCursor)

					if targetFileItem != nil {
						// Set the source file for the new item
						newItem.SourceFile = targetFileItem.SourceFile
						newItem.Level = 2 // Children of file items are level 2

						// Insert at the beginning of the file item's children
						targetFileItem.Children = append([]*model.Item{newItem}, targetFileItem.Children...)
						targetFileItem.Folded = false // Unfold to show the new item
						m.setStatus("TODO captured to " + targetFileItem.Title)
					} else {
						m.setStatus("Error: Could not find file to add to")
					}
				} else {
					// Single file mode: insert at beginning
					m.orgFile.Items = append([]*model.Item{newItem}, m.orgFile.Items...)
					m.setStatus("TODO captured!")
				}
			}
			m.mode = modeList
			m.textinput.Blur()
			// Don't reset cursor, keep it where it was
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

// findTopLevelFileItem finds the top-level file item that contains the item at the given cursor position
func (m *uiModel) findTopLevelFileItem(items []*model.Item, cursorPos int) *model.Item {
	if cursorPos < 0 || cursorPos >= len(items) {
		// Fallback to first file if cursor out of bounds
		if len(m.orgFile.Items) > 0 {
			return m.orgFile.Items[0]
		}
		return nil
	}

	selectedItem := items[cursorPos]

	// Check if we're in multi-file mode
	isMultiFile := len(m.orgFile.Items) > 0 && m.orgFile.Items[0].SourceFile != ""

	if !isMultiFile {
		// Not in multi-file mode, return nil
		return nil
	}

	// If the selected item itself is a file item (level 1 with SourceFile), return it
	if selectedItem.SourceFile != "" && selectedItem.Level == 1 {
		return selectedItem
	}

	// Otherwise, find which top-level file item this item belongs to
	// by checking the SourceFile field
	if selectedItem.SourceFile != "" {
		for _, fileItem := range m.orgFile.Items {
			if fileItem.SourceFile == selectedItem.SourceFile {
				return fileItem
			}
		}
	}

	// Fallback: return the first file item
	if len(m.orgFile.Items) > 0 {
		return m.orgFile.Items[0]
	}

	return nil
}

func (m uiModel) updateAddSubTask(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textinput.Width = 50

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			title := strings.TrimSpace(m.textinput.Value())
			if title != "" && m.editingItem != nil {
				// Get default state from config
				defaultState := model.TodoState(m.config.GetDefaultNewTaskState())

				// Create new sub-task
				newItem := &model.Item{
					Level:      m.editingItem.Level + 1,
					State:      defaultState,
					Title:      title,
					Notes:      []string{},
					Children:   []*model.Item{},
					SourceFile: m.editingItem.SourceFile, // Inherit source file from parent
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

func (m uiModel) updateSetDeadline(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textinput.Width = 50

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
								m.editingItem.Notes[i] = fmt.Sprintf("DEADLINE: <%s>", parser.FormatOrgDate(deadline))
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

func (m uiModel) updateSetPriority(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "A", "a":
			if m.editingItem != nil {
				m.editingItem.Priority = model.PriorityA
				m.setStatus("Priority set to A")
			}
			m.mode = modeList
			m.editingItem = nil
			return m, nil
		case "B", "b":
			if m.editingItem != nil {
				m.editingItem.Priority = model.PriorityB
				m.setStatus("Priority set to B")
			}
			m.mode = modeList
			m.editingItem = nil
			return m, nil
		case "C", "c":
			if m.editingItem != nil {
				m.editingItem.Priority = model.PriorityC
				m.setStatus("Priority set to C")
			}
			m.mode = modeList
			m.editingItem = nil
			return m, nil
		case " ", "enter":
			// Clear priority
			if m.editingItem != nil {
				m.editingItem.Priority = model.PriorityNone
				m.setStatus("Priority cleared")
			}
			m.mode = modeList
			m.editingItem = nil
			return m, nil
		case "esc":
			m.mode = modeList
			m.editingItem = nil
			m.setStatus("Cancelled")
			return m, nil
		}
	}
	return m, nil
}

func (m uiModel) updateSetEffort(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textinput.Width = 50

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			input := strings.TrimSpace(m.textinput.Value())
			if m.editingItem != nil {
				if input == "" {
					// Empty input clears the effort
					m.editingItem.Effort = ""
					m.setStatus("Effort cleared!")
				} else {
					// Set the effort value
					m.editingItem.Effort = input
					m.setStatus("Effort set!")
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

func (m *uiModel) cycleStateForward(item *model.Item) {
	stateNames := m.config.GetStateNames()
	if len(stateNames) == 0 {
		return
	}

	// Find current state index
	currentIndex := -1
	currentState := string(item.State)
	lastStateIndex := len(stateNames) - 1

	// Handle empty state
	if currentState == "" {
		currentIndex = -1
	} else {
		for i, name := range stateNames {
			if name == currentState {
				currentIndex = i
				break
			}
		}
	}

	// Store the old state to check if we're transitioning to/from DONE
	oldState := currentState
	var newState string

	// Cycle forward
	if currentIndex < 0 || currentIndex >= len(stateNames)-1 {
		if currentIndex == len(stateNames)-1 {
			newState = "" // Back to empty
		} else {
			newState = stateNames[0] // First state
		}
	} else {
		newState = stateNames[currentIndex+1]
	}

	// Update the item state
	item.State = model.TodoState(newState)

	// Manage CLOSED timestamp
	wasInDoneState := (oldState == stateNames[lastStateIndex])
	isInDoneState := (newState == stateNames[lastStateIndex])

	if isInDoneState && !wasInDoneState {
		// Moving TO done state - add CLOSED timestamp
		now := time.Now()
		item.Closed = &now
		// Remove any existing CLOSED line from notes
		var filteredNotes []string
		for _, note := range item.Notes {
			if !strings.HasPrefix(strings.TrimSpace(note), "CLOSED:") {
				filteredNotes = append(filteredNotes, note)
			}
		}
		item.Notes = filteredNotes
	} else if wasInDoneState && !isInDoneState {
		// Moving FROM done state - remove CLOSED timestamp
		item.Closed = nil
		// Remove any existing CLOSED line from notes
		var filteredNotes []string
		for _, note := range item.Notes {
			if !strings.HasPrefix(strings.TrimSpace(note), "CLOSED:") {
				filteredNotes = append(filteredNotes, note)
			}
		}
		item.Notes = filteredNotes
	}
}

func (m *uiModel) cycleStateBackward(item *model.Item) {
	stateNames := m.config.GetStateNames()
	if len(stateNames) == 0 {
		return
	}

	// Find current state index
	currentIndex := -1
	currentState := string(item.State)
	lastStateIndex := len(stateNames) - 1

	// Handle empty state
	if currentState == "" {
		currentIndex = len(stateNames) // One past the last state
	} else {
		for i, name := range stateNames {
			if name == currentState {
				currentIndex = i
				break
			}
		}
	}

	// Store the old state to check if we're transitioning to/from DONE
	oldState := currentState
	var newState string

	// Cycle backward
	if currentIndex <= 0 {
		newState = "" // Empty state
	} else if currentIndex > len(stateNames) {
		newState = stateNames[len(stateNames)-1]
	} else {
		newState = stateNames[currentIndex-1]
	}

	// Update the item state
	item.State = model.TodoState(newState)

	// Manage CLOSED timestamp
	wasInDoneState := (oldState == stateNames[lastStateIndex])
	isInDoneState := (newState == stateNames[lastStateIndex])

	if isInDoneState && !wasInDoneState {
		// Moving TO done state - add CLOSED timestamp
		now := time.Now()
		item.Closed = &now
		// Remove any existing CLOSED line from notes
		var filteredNotes []string
		for _, note := range item.Notes {
			if !strings.HasPrefix(strings.TrimSpace(note), "CLOSED:") {
				filteredNotes = append(filteredNotes, note)
			}
		}
		item.Notes = filteredNotes
	} else if wasInDoneState && !isInDoneState {
		// Moving FROM done state - remove CLOSED timestamp
		item.Closed = nil
		// Remove any existing CLOSED line from notes
		var filteredNotes []string
		for _, note := range item.Notes {
			if !strings.HasPrefix(strings.TrimSpace(note), "CLOSED:") {
				filteredNotes = append(filteredNotes, note)
			}
		}
		item.Notes = filteredNotes
	}
}

func (m *uiModel) deleteItem(item *model.Item) {
	var removeFromList func([]*model.Item, *model.Item) []*model.Item
	removeFromList = func(items []*model.Item, target *model.Item) []*model.Item {
		result := []*model.Item{}
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

func (m *uiModel) moveItemUp() {
	items := m.getVisibleItems()
	if len(items) == 0 || m.cursor == 0 {
		return
	}

	currentItem := items[m.cursor]

	// Find the previous sibling (not child, not parent's sibling)
	prevSibling := m.findPreviousSibling(currentItem)
	if prevSibling == nil {
		m.setStatus("Cannot move - already at top of list")
		return
	}

	m.swapItems(currentItem, prevSibling)
	m.setStatus("Item moved up")

	// Update cursor to follow the item
	items = m.getVisibleItems()
	for i, item := range items {
		if item == currentItem {
			m.cursor = i
			break
		}
	}
}

func (m *uiModel) moveItemDown() {
	items := m.getVisibleItems()
	if len(items) == 0 || m.cursor >= len(items)-1 {
		return
	}

	currentItem := items[m.cursor]

	// Find the next sibling (not child, not parent's sibling)
	nextSibling := m.findNextSibling(currentItem)
	if nextSibling == nil {
		m.setStatus("Cannot move - already at bottom of list")
		return
	}

	m.swapItems(currentItem, nextSibling)
	m.setStatus("Item moved down")

	// Update cursor to follow the item
	items = m.getVisibleItems()
	for i, item := range items {
		if item == currentItem {
			m.cursor = i
			break
		}
	}
}

func (m *uiModel) findPreviousSibling(item *model.Item) *model.Item {
	// Find the parent list containing this item and return the previous sibling
	var findInList func([]*model.Item) *model.Item
	findInList = func(items []*model.Item) *model.Item {
		for i, it := range items {
			if it == item && i > 0 {
				// Found the item and there's a previous sibling
				return items[i-1]
			}
			// Recursively check children
			if result := findInList(it.Children); result != nil {
				return result
			}
		}
		return nil
	}
	return findInList(m.orgFile.Items)
}

func (m *uiModel) findNextSibling(item *model.Item) *model.Item {
	// Find the parent list containing this item and return the next sibling
	var findInList func([]*model.Item) *model.Item
	findInList = func(items []*model.Item) *model.Item {
		for i, it := range items {
			if it == item && i < len(items)-1 {
				// Found the item and there's a next sibling
				return items[i+1]
			}
			// Recursively check children
			if result := findInList(it.Children); result != nil {
				return result
			}
		}
		return nil
	}
	return findInList(m.orgFile.Items)
}

func (m *uiModel) swapItems(item1, item2 *model.Item) {
	// Find parent list containing both items
	var swapInList func([]*model.Item) bool
	swapInList = func(items []*model.Item) bool {
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

func (m *uiModel) promoteItem() {
	items := m.getVisibleItems()
	if len(items) == 0 || m.cursor >= len(items) {
		return
	}

	currentItem := items[m.cursor]

	// Can't promote a top-level item
	if currentItem.Level <= 1 {
		m.setStatus("Cannot promote - already at top level")
		return
	}

	// Find the parent of this item
	parent := m.findParent(currentItem)
	if parent == nil {
		m.setStatus("Cannot promote - no parent found")
		return
	}

	// Remove item from parent's children
	for i, child := range parent.Children {
		if child == currentItem {
			parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
			break
		}
	}

	// Find grandparent to insert this item after the parent
	grandparent := m.findParent(parent)
	if grandparent != nil {
		// Insert after parent in grandparent's children
		for i, child := range grandparent.Children {
			if child == parent {
				// Decrease level and update all descendants
				m.adjustItemLevels(currentItem, -1)
				grandparent.Children = append(grandparent.Children[:i+1], append([]*model.Item{currentItem}, grandparent.Children[i+1:]...)...)
				break
			}
		}
	} else {
		// Parent is at top level, insert after parent in m.orgFile.Items
		for i, item := range m.orgFile.Items {
			if item == parent {
				// Decrease level and update all descendants
				m.adjustItemLevels(currentItem, -1)
				m.orgFile.Items = append(m.orgFile.Items[:i+1], append([]*model.Item{currentItem}, m.orgFile.Items[i+1:]...)...)
				break
			}
		}
	}

	m.setStatus("Item promoted")

	// Update cursor to follow the item
	items = m.getVisibleItems()
	for i, item := range items {
		if item == currentItem {
			m.cursor = i
			break
		}
	}
}

func (m *uiModel) demoteItem() {
	items := m.getVisibleItems()
	if len(items) == 0 || m.cursor >= len(items) {
		return
	}

	currentItem := items[m.cursor]

	// Find the previous sibling to make this item its child
	prevSibling := m.findPreviousSibling(currentItem)
	if prevSibling == nil {
		m.setStatus("Cannot demote - no previous sibling")
		return
	}

	// Remove item from its current parent's children
	parent := m.findParent(currentItem)
	if parent != nil {
		for i, child := range parent.Children {
			if child == currentItem {
				parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
				break
			}
		}
	} else {
		// Item is at top level
		for i, item := range m.orgFile.Items {
			if item == currentItem {
				m.orgFile.Items = append(m.orgFile.Items[:i], m.orgFile.Items[i+1:]...)
				break
			}
		}
	}

	// Increase level and update all descendants
	m.adjustItemLevels(currentItem, 1)

	// Add as child of previous sibling
	prevSibling.Children = append(prevSibling.Children, currentItem)
	prevSibling.Folded = false // Unfold to show the demoted item

	m.setStatus("Item demoted")

	// Update cursor to follow the item
	items = m.getVisibleItems()
	for i, item := range items {
		if item == currentItem {
			m.cursor = i
			break
		}
	}
}

func (m *uiModel) findParent(target *model.Item) *model.Item {
	var findInList func([]*model.Item) *model.Item
	findInList = func(items []*model.Item) *model.Item {
		for _, item := range items {
			// Check if target is a direct child
			for _, child := range item.Children {
				if child == target {
					return item
				}
			}
			// Recursively check children
			if result := findInList(item.Children); result != nil {
				return result
			}
		}
		return nil
	}
	return findInList(m.orgFile.Items)
}

func (m *uiModel) adjustItemLevels(item *model.Item, delta int) {
	item.Level += delta
	for _, child := range item.Children {
		m.adjustItemLevels(child, delta)
	}
}

func (m uiModel) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "?", "esc", "q":
			m.mode = modeList
			m.helpScroll = 0 // Reset scroll when exiting
			return m, nil
		case "up", "k":
			if m.helpScroll > 0 {
				m.helpScroll--
			}
			return m, nil
		case "down", "j":
			m.helpScroll++
			// The view will handle clamping to max scroll
			return m, nil
		case "pageup":
			m.helpScroll -= 10
			if m.helpScroll < 0 {
				m.helpScroll = 0
			}
			return m, nil
		case "pagedown":
			m.helpScroll += 10
			return m, nil
		case "home", "g":
			m.helpScroll = 0
			return m, nil
		}
	}
	return m, nil
}

// updateTagEdit handles tag editing mode
func (m *uiModel) updateTagEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.mode = modeList
			m.textinput.Blur()
			return m, nil

		case msg.Type == tea.KeyEnter:
			if m.editingItem != nil {
				// Parse tags from input (colon-separated)
				tagsStr := m.textinput.Value()
				var tags []string
				if tagsStr != "" {
					tags = strings.Split(tagsStr, ":")
					// Remove empty strings
					var filteredTags []string
					for _, tag := range tags {
						tag = strings.TrimSpace(tag)
						if tag != "" {
							filteredTags = append(filteredTags, tag)
						}
					}
					tags = filteredTags
				}
				m.editingItem.Tags = tags
				m.setStatus("Tags updated")
			}
			m.mode = modeList
			m.textinput.Blur()
			return m, nil

		default:
			var cmd tea.Cmd
			m.textinput, cmd = m.textinput.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

// updateRename handles item rename mode
func (m *uiModel) updateRename(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.mode = modeList
			m.textinput.Blur()
			m.editingItem = nil
			return m, nil

		case msg.Type == tea.KeyEnter:
			if m.editingItem != nil {
				newTitle := strings.TrimSpace(m.textinput.Value())
				if newTitle != "" {
					m.editingItem.Title = newTitle
					m.setStatus("Item renamed")
				} else {
					m.setStatus("Cannot rename to empty title")
				}
			}
			m.mode = modeList
			m.textinput.Blur()
			m.editingItem = nil
			return m, nil

		case msg.Type == tea.KeyEsc:
			m.mode = modeList
			m.textinput.Blur()
			m.editingItem = nil
			m.setStatus("Cancelled")
			return m, nil

		default:
			var cmd tea.Cmd
			m.textinput, cmd = m.textinput.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}
