package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// settingsSection represents different sections in the settings view
type settingsSection int

const (
	settingsSectionTags settingsSection = iota
	settingsSectionStates
	settingsSectionKeybindings
)

// initSettings initializes the settings state
func (m *uiModel) initSettings() {
	m.settingsCursor = 0
	m.settingsScroll = 0
}

// updateSettings handles updates in settings mode
func (m *uiModel) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If editing, handle text input
		if m.textinput.Focused() {
			switch {
			case msg.Type == tea.KeyEsc:
				m.textinput.Blur()
				return m, nil
			case msg.Type == tea.KeyEnter:
				// Save the edited value
				m.saveSettingsEdit()
				m.textinput.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.textinput, cmd = m.textinput.Update(msg)
				return m, cmd
			}
		}

		// Navigation and actions
		switch {
		case key.Matches(msg, m.keys.Quit), key.Matches(msg, m.keys.Settings):
			// Exit settings
			m.mode = modeList
			return m, nil

		case key.Matches(msg, m.keys.Up):
			if m.settingsCursor > 0 {
				m.settingsCursor--
				m.updateSettingsScrollOffset()
			}

		case key.Matches(msg, m.keys.Down):
			maxCursor := m.getSettingsItemCount() - 1
			if m.settingsCursor < maxCursor {
				m.settingsCursor++
				m.updateSettingsScrollOffset()
			}

		case key.Matches(msg, m.keys.ShiftUp):
			// Move item up
			m.moveSettingsItemUp()

		case key.Matches(msg, m.keys.ShiftDown):
			// Move item down
			m.moveSettingsItemDown()

		case key.Matches(msg, m.keys.Left):
			// Previous section
			if m.settingsSection > settingsSectionTags {
				m.settingsSection--
				m.settingsCursor = 0
				m.settingsScroll = 0
			}

		case key.Matches(msg, m.keys.Right):
			// Next section
			if m.settingsSection < settingsSectionKeybindings {
				m.settingsSection++
				m.settingsCursor = 0
				m.settingsScroll = 0
			}

		case key.Matches(msg, m.keys.EditNotes):
			// Enter edit mode
			m.startSettingsEdit()

		case key.Matches(msg, m.keys.Delete):
			// Delete tag
			m.deleteSettingsItem()

		case key.Matches(msg, m.keys.Capture):
			// Add new tag or state
			switch m.settingsSection {
			case settingsSectionTags:
				m.addNewTag()
			case settingsSectionStates:
				m.addNewState()
			case settingsSectionKeybindings:
				// Cannot add keybindings yet
			}

		case key.Matches(msg, m.keys.Save):
			// Save config to disk
			if err := m.config.Save(); err != nil {
				m.setStatus(fmt.Sprintf("Error saving config: %v", err))
			} else {
				m.setStatus("Configuration saved!")
				// Reload keybindings and styles
				m.keys = newKeyMapFromConfig(m.config)
				m.styles = newStyleMapFromConfig(m.config)
			}
		}
	}

	return m, nil
}

// getSettingsItemCount returns the number of items in the current settings view
func (m *uiModel) getSettingsItemCount() int {
	switch m.settingsSection {
	case settingsSectionTags:
		return len(m.config.Tags.Tags) + 1 // +1 for "Add new tag" option
	case settingsSectionStates:
		return len(m.config.States.States) + 2 // +1 for "Default new task state" setting, +1 for "Add new state" option
	case settingsSectionKeybindings:
		return len(m.config.GetAllKeybindings())
	default:
		return 0
	}
}

// updateSettingsScrollOffset adjusts the scroll offset to keep the cursor visible
func (m *uiModel) updateSettingsScrollOffset() {
	// Calculate available height for content
	// Reserve space for: title (2 lines), tabs (2 lines), instructions (3 lines),
	// input field if focused (3 lines), status bar, and some padding
	reservedLines := 10
	if m.textinput.Focused() {
		reservedLines += 3
	}

	availableHeight := m.height - reservedLines
	if availableHeight < 3 {
		availableHeight = 3 // Minimum visible items
	}

	// Adjust scroll to keep cursor visible
	if m.settingsCursor < m.settingsScroll {
		// Cursor is above visible area, scroll up
		m.settingsScroll = m.settingsCursor
	} else if m.settingsCursor >= m.settingsScroll+availableHeight {
		// Cursor is below visible area, scroll down
		m.settingsScroll = m.settingsCursor - availableHeight + 1
	}

	// Ensure scroll offset doesn't go negative
	if m.settingsScroll < 0 {
		m.settingsScroll = 0
	}
}

// startSettingsEdit starts editing a settings item
func (m *uiModel) startSettingsEdit() {
	switch m.settingsSection {
	case settingsSectionTags:
		if m.settingsCursor >= len(m.config.Tags.Tags) {
			return
		}
		tag := m.config.Tags.Tags[m.settingsCursor]
		m.textinput.SetValue(tag.Name + "," + tag.Color)
		m.textinput.Placeholder = "name,color (e.g., work,99)"
		m.textinput.Focus()

	case settingsSectionStates:
		// First item is the default new task state setting
		if m.settingsCursor == 0 {
			m.textinput.SetValue(m.config.States.DefaultNewTaskState)
			m.textinput.Placeholder = "Enter state name or leave empty for none"
			m.textinput.Focus()
			return
		}

		// Adjust for the default state setting offset
		stateIndex := m.settingsCursor - 1
		if stateIndex >= len(m.config.States.States) {
			return
		}
		state := m.config.States.States[stateIndex]
		m.textinput.SetValue(state.Name + "," + state.Color)
		m.textinput.Placeholder = "name,color (e.g., TODO,202)"
		m.textinput.Focus()

	case settingsSectionKeybindings:
		// Edit keybinding
		keybindings := m.config.GetAllKeybindings()

		// Convert to sorted slice
		type kbPair struct {
			action string
			keys   []string
		}
		var kbList []kbPair
		for action, keys := range keybindings {
			kbList = append(kbList, kbPair{action, keys})
		}

		// Sort alphabetically
		for i := 0; i < len(kbList)-1; i++ {
			for j := i + 1; j < len(kbList); j++ {
				if kbList[i].action > kbList[j].action {
					kbList[i], kbList[j] = kbList[j], kbList[i]
				}
			}
		}

		if m.settingsCursor >= len(kbList) {
			return
		}

		kb := kbList[m.settingsCursor]
		m.textinput.SetValue(strings.Join(kb.keys, ","))
		m.textinput.Placeholder = "Enter keys separated by commas (e.g., up,k)"
		m.textinput.Focus()
	}
}

// saveSettingsEdit saves the edited value and auto-saves to disk
func (m *uiModel) saveSettingsEdit() {
	switch m.settingsSection {
	case settingsSectionTags:
		if m.settingsCursor >= len(m.config.Tags.Tags) {
			return
		}
		// Parse "name,color" format
		parts := strings.Split(m.textinput.Value(), ",")
		if len(parts) >= 2 {
			tag := &m.config.Tags.Tags[m.settingsCursor]
			tag.Name = strings.TrimSpace(parts[0])
			tag.Color = strings.TrimSpace(parts[1])
			m.setStatus(fmt.Sprintf("Updated tag '%s' (saved)", tag.Name))
		} else {
			m.setStatus("Invalid format. Use: name,color")
			return
		}

	case settingsSectionStates:
		// First item is the default new task state setting
		if m.settingsCursor == 0 {
			newDefault := strings.TrimSpace(m.textinput.Value())
			// Convert to uppercase
			newDefault = strings.ToUpper(newDefault)
			m.config.States.DefaultNewTaskState = newDefault
			if newDefault == "" {
				m.setStatus("Default new task state set to 'none' (saved)")
			} else {
				m.setStatus(fmt.Sprintf("Default new task state set to '%s' (saved)", newDefault))
			}
			// Auto-save
			if err := m.config.Save(); err != nil {
				m.setStatus(fmt.Sprintf("Error auto-saving config: %v", err))
			}
			return
		}

		// Adjust for the default state setting offset
		stateIndex := m.settingsCursor - 1
		if stateIndex >= len(m.config.States.States) {
			return
		}
		// Parse "name,color" format
		parts := strings.Split(m.textinput.Value(), ",")
		if len(parts) >= 2 {
			state := &m.config.States.States[stateIndex]
			state.Name = strings.TrimSpace(parts[0])
			state.Color = strings.TrimSpace(parts[1])
			m.setStatus(fmt.Sprintf("Updated state '%s' (saved)", state.Name))
		} else {
			m.setStatus("Invalid format. Use: name,color")
			return
		}

	case settingsSectionKeybindings:
		// Save keybinding
		keybindings := m.config.GetAllKeybindings()

		// Convert to sorted slice
		type kbPair struct {
			action string
			keys   []string
		}
		var kbList []kbPair
		for action, keys := range keybindings {
			kbList = append(kbList, kbPair{action, keys})
		}

		// Sort alphabetically
		for i := 0; i < len(kbList)-1; i++ {
			for j := i + 1; j < len(kbList); j++ {
				if kbList[i].action > kbList[j].action {
					kbList[i], kbList[j] = kbList[j], kbList[i]
				}
			}
		}

		if m.settingsCursor >= len(kbList) {
			return
		}

		kb := kbList[m.settingsCursor]
		// Parse comma-separated keys
		keysStr := m.textinput.Value()
		var newKeys []string
		for _, k := range strings.Split(keysStr, ",") {
			k = strings.TrimSpace(k)
			if k != "" {
				newKeys = append(newKeys, k)
			}
		}

		if len(newKeys) > 0 {
			if err := m.config.UpdateKeybinding(kb.action, newKeys); err != nil {
				m.setStatus(fmt.Sprintf("Error updating keybinding: %v", err))
				return
			} else {
				m.setStatus(fmt.Sprintf("Updated keybinding for '%s' (saved)", kb.action))
			}
		}
	}

	// Auto-save configuration to disk
	if err := m.config.Save(); err != nil {
		m.setStatus(fmt.Sprintf("Error auto-saving config: %v", err))
	} else {
		// Reload keybindings and styles from updated config
		m.keys = newKeyMapFromConfig(m.config)
		m.styles = newStyleMapFromConfig(m.config)
	}
}

// deleteSettingsItem deletes the current settings item and auto-saves
func (m *uiModel) deleteSettingsItem() {
	switch m.settingsSection {
	case settingsSectionTags:
		if m.settingsCursor >= len(m.config.Tags.Tags) {
			return
		}
		tag := m.config.Tags.Tags[m.settingsCursor]
		m.config.RemoveTag(tag.Name)
		m.setStatus(fmt.Sprintf("Deleted tag '%s' (saved)", tag.Name))

		// Adjust cursor if needed
		if m.settingsCursor >= len(m.config.Tags.Tags) {
			m.settingsCursor = len(m.config.Tags.Tags) - 1
			if m.settingsCursor < 0 {
				m.settingsCursor = 0
			}
		}

	case settingsSectionStates:
		// Cannot delete the default new task state setting (first item)
		if m.settingsCursor == 0 {
			m.setStatus("Cannot delete default state setting (use Enter to edit)")
			return
		}

		// Adjust for the default state setting offset
		stateIndex := m.settingsCursor - 1
		if stateIndex >= len(m.config.States.States) {
			return
		}
		state := m.config.States.States[stateIndex]
		m.config.RemoveState(state.Name)
		m.setStatus(fmt.Sprintf("Deleted state '%s' (saved)", state.Name))

		// Adjust cursor if needed
		// +1 for the default state setting
		if m.settingsCursor >= len(m.config.States.States)+1 {
			m.settingsCursor = len(m.config.States.States)
			if m.settingsCursor < 1 {
				m.settingsCursor = 1
			}
		}

	case settingsSectionKeybindings:
		// Keybindings cannot be deleted
		return
	}

	// Auto-save configuration to disk
	if err := m.config.Save(); err != nil {
		m.setStatus(fmt.Sprintf("Error auto-saving config: %v", err))
	}
}

// addNewTag adds a new tag
func (m *uiModel) addNewTag() {
	m.textinput.SetValue("")
	m.textinput.Placeholder = "Enter tag name"
	m.textinput.Focus()
	m.textinput.Blur() // Will be refocused when user types

	// Prompt for tag name first, then color
	m.mode = modeSettingsAddTag
}

// addNewState adds a new state
func (m *uiModel) addNewState() {
	m.textinput.SetValue("")
	m.textinput.Placeholder = "Enter state name"
	m.textinput.Focus()
	m.textinput.Blur() // Will be refocused when user types

	// Prompt for state name first, then color
	m.mode = modeSettingsAddState
}

// viewSettings renders the settings view
func (m *uiModel) viewSettings() string {
	var content strings.Builder

	// Title
	title := m.styles.titleStyle.Render("⚙ Settings")
	content.WriteString(title + "\n\n")

	// Tab selector
	tabStyle := lipgloss.NewStyle().Padding(0, 2)
	activeTabStyle := lipgloss.NewStyle().Padding(0, 2).Bold(true).Foreground(lipgloss.Color(m.config.Colors.Title))

	tabs := ""
	if m.settingsSection == settingsSectionTags {
		tabs += activeTabStyle.Render("[Tags]")
	} else {
		tabs += tabStyle.Render("Tags")
	}
	tabs += " "
	if m.settingsSection == settingsSectionStates {
		tabs += activeTabStyle.Render("[States]")
	} else {
		tabs += tabStyle.Render("States")
	}
	tabs += " "
	if m.settingsSection == settingsSectionKeybindings {
		tabs += activeTabStyle.Render("[Keybindings]")
	} else {
		tabs += tabStyle.Render("Keybindings")
	}
	content.WriteString(tabs + "\n\n")

	// Instructions
	var instructions string
	switch m.settingsSection {
	case settingsSectionTags:
		instructions = "←/→: Switch tabs • ↑/↓: Navigate • Enter: Edit • D: Delete\nc: Add new tag • ctrl+s: Save • q/,: Exit"
	case settingsSectionStates:
		instructions = "←/→: Switch tabs • ↑/↓: Navigate • shift+↑/↓: Reorder • Enter: Edit\nD: Delete • c: Add new state • ctrl+s: Save • q/,: Exit"
	case settingsSectionKeybindings:
		instructions = "←/→: Switch tabs • ↑/↓: Navigate • Enter: Edit keybinding\nctrl+s: Save • q/,: Exit"
	}
	content.WriteString(m.styles.statusStyle.Render(instructions) + "\n\n")

	// Render the appropriate section
	switch m.settingsSection {
	case settingsSectionTags:
		content.WriteString(m.viewSettingsTags())
	case settingsSectionStates:
		content.WriteString(m.viewSettingsStates())
	case settingsSectionKeybindings:
		content.WriteString(m.viewSettingsKeybindings())
	}

	// If editing, show input
	if m.textinput.Focused() {
		content.WriteString("\n")
		content.WriteString(m.textinput.View() + "\n")
		content.WriteString(m.styles.statusStyle.Render("Enter: Save • ESC: Cancel") + "\n")
	}

	return content.String()
}

// viewSettingsTags renders the tags section
func (m *uiModel) viewSettingsTags() string {
	var content strings.Builder

	// Calculate visible window
	reservedLines := 10
	if m.textinput.Focused() {
		reservedLines += 3
	}
	availableHeight := m.height - reservedLines
	if availableHeight < 3 {
		availableHeight = 3
	}

	endIdx := m.settingsScroll + availableHeight
	totalItems := len(m.config.Tags.Tags) + 1 // +1 for "Add new tag"

	for i, tag := range m.config.Tags.Tags {
		// Skip items outside visible window
		if i < m.settingsScroll || i >= endIdx {
			continue
		}

		line := ""

		// Cursor
		if i == m.settingsCursor && !m.textinput.Focused() {
			line += "▶ "
		} else {
			line += "  "
		}

		// Tag name with its color
		tagStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(tag.Color))
		line += fmt.Sprintf(":%s: ", tag.Name)
		line += tagStyle.Render(fmt.Sprintf("(color: %s)", tag.Color))

		content.WriteString(line + "\n")
	}

	// Add new tag option
	addNewIdx := len(m.config.Tags.Tags)
	if addNewIdx >= m.settingsScroll && addNewIdx < endIdx {
		if m.settingsCursor == len(m.config.Tags.Tags) && !m.textinput.Focused() {
			content.WriteString("▶ ")
		} else {
			content.WriteString("  ")
		}
		content.WriteString(m.styles.statusStyle.Render("+ Add new tag (press 'c')") + "\n")
	}

	// Add scroll indicator if needed
	if totalItems > availableHeight {
		scrollInfo := fmt.Sprintf("\n[Showing %d-%d of %d]",
			m.settingsScroll+1,
			min(endIdx, totalItems),
			totalItems)
		content.WriteString(m.styles.statusStyle.Render(scrollInfo))
	}

	return content.String()
}

// viewSettingsStates renders the states section
func (m *uiModel) viewSettingsStates() string {
	var content strings.Builder

	// Calculate visible window
	reservedLines := 10
	if m.textinput.Focused() {
		reservedLines += 3
	}
	availableHeight := m.height - reservedLines
	if availableHeight < 3 {
		availableHeight = 3
	}

	endIdx := m.settingsScroll + availableHeight
	totalItems := len(m.config.States.States) + 2 // +1 for default state, +1 for "Add new state"

	// First show the default new task state setting (item 0)
	if 0 >= m.settingsScroll && 0 < endIdx {
		line := ""
		if m.settingsCursor == 0 && !m.textinput.Focused() {
			line += "▶ "
		} else {
			line += "  "
		}
		line += "Default new task state: "
		if m.config.States.DefaultNewTaskState == "" {
			line += m.styles.statusStyle.Render("(none)")
		} else {
			// Try to get the color for this state
			color := m.config.GetStateColor(m.config.States.DefaultNewTaskState)
			stateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
			line += stateStyle.Render(m.config.States.DefaultNewTaskState)
		}
		content.WriteString(line + "\n\n")
	}

	// Then show all configured states
	for i, state := range m.config.States.States {
		itemIdx := i + 1 // Offset by 1 for the default state setting

		// Skip items outside visible window
		if itemIdx < m.settingsScroll || itemIdx >= endIdx {
			continue
		}

		line := ""

		// Cursor
		if itemIdx == m.settingsCursor && !m.textinput.Focused() {
			line += "▶ "
		} else {
			line += "  "
		}

		// State name with its color
		stateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(state.Color))
		line += stateStyle.Render(state.Name)
		line += fmt.Sprintf(" (color: %s)", state.Color)

		content.WriteString(line + "\n")
	}

	// Add new state option
	addNewIdx := len(m.config.States.States) + 1
	if addNewIdx >= m.settingsScroll && addNewIdx < endIdx {
		if m.settingsCursor == addNewIdx && !m.textinput.Focused() {
			content.WriteString("▶ ")
		} else {
			content.WriteString("  ")
		}
		content.WriteString(m.styles.statusStyle.Render("+ Add new state (press 'c')") + "\n")
	}

	// Add scroll indicator if needed
	if totalItems > availableHeight {
		scrollInfo := fmt.Sprintf("\n[Showing %d-%d of %d]",
			m.settingsScroll+1,
			min(endIdx, totalItems),
			totalItems)
		content.WriteString(m.styles.statusStyle.Render(scrollInfo))
	}

	return content.String()
}

// viewSettingsKeybindings renders the keybindings section
func (m *uiModel) viewSettingsKeybindings() string {
	var content strings.Builder

	// Calculate visible window
	reservedLines := 10
	if m.textinput.Focused() {
		reservedLines += 3
	}
	availableHeight := m.height - reservedLines
	if availableHeight < 3 {
		availableHeight = 3
	}

	// Get all keybindings
	keybindings := m.config.GetAllKeybindings()

	// Convert to sorted slice for consistent display
	type kbPair struct {
		action string
		keys   []string
	}
	var kbList []kbPair
	for action, keys := range keybindings {
		kbList = append(kbList, kbPair{action, keys})
	}

	// Simple alphabetical sort by action name
	for i := 0; i < len(kbList)-1; i++ {
		for j := i + 1; j < len(kbList); j++ {
			if kbList[i].action > kbList[j].action {
				kbList[i], kbList[j] = kbList[j], kbList[i]
			}
		}
	}

	endIdx := m.settingsScroll + availableHeight
	totalItems := len(kbList)

	for i, kb := range kbList {
		// Skip items outside visible window
		if i < m.settingsScroll || i >= endIdx {
			continue
		}

		line := ""

		// Cursor
		if i == m.settingsCursor && !m.textinput.Focused() {
			line += "▶ "
		} else {
			line += "  "
		}

		// Format keybinding
		keysStr := strings.Join(kb.keys, ", ")
		line += fmt.Sprintf("%-20s : %s", kb.action, keysStr)

		content.WriteString(line + "\n")
	}

	// Add scroll indicator if needed
	if totalItems > availableHeight {
		scrollInfo := fmt.Sprintf("\n[Showing %d-%d of %d]",
			m.settingsScroll+1,
			min(endIdx, totalItems),
			totalItems)
		content.WriteString(m.styles.statusStyle.Render(scrollInfo))
	}

	return content.String()
}

// modeSettingsAddTag is a special mode for adding tags
const modeSettingsAddTag viewMode = 100

// modeSettingsAddState is a special mode for adding states
const modeSettingsAddState viewMode = 101

// updateSettingsAddTag handles the add tag flow
func (m *uiModel) updateSettingsAddTag(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.textinput.Focused() {
			m.textinput.Focus()
		}

		switch {
		case msg.Type == tea.KeyEsc:
			m.textinput.Blur()
			m.mode = modeSettings
			return m, nil

		case msg.Type == tea.KeyEnter:
			tagName := m.textinput.Value()
			if tagName != "" {
				// Default color
				m.config.AddTag(tagName, "99")
				// Auto-save
				if err := m.config.Save(); err != nil {
					m.setStatus(fmt.Sprintf("Error saving: %v", err))
				} else {
					m.setStatus(fmt.Sprintf("Added tag '%s' (saved)", tagName))
				}
			}
			m.textinput.Blur()
			m.mode = modeSettings
			return m, nil

		default:
			var cmd tea.Cmd
			m.textinput, cmd = m.textinput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// viewSettingsAddTag renders the add tag view
func (m *uiModel) viewSettingsAddTag() string {
	var content strings.Builder

	content.WriteString(m.styles.titleStyle.Render("Add New Tag") + "\n\n")
	content.WriteString(m.textinput.View() + "\n\n")
	content.WriteString(m.styles.statusStyle.Render("Enter tag name • Press Enter to add • ESC to cancel") + "\n")

	return content.String()
}

// updateSettingsAddState handles the add state flow
func (m *uiModel) updateSettingsAddState(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.textinput.Focused() {
			m.textinput.Focus()
		}

		switch {
		case msg.Type == tea.KeyEsc:
			m.textinput.Blur()
			m.mode = modeSettings
			return m, nil

		case msg.Type == tea.KeyEnter:
			stateName := m.textinput.Value()
			if stateName != "" {
				// Default color
				m.config.AddState(stateName, "99")
				// Auto-save
				if err := m.config.Save(); err != nil {
					m.setStatus(fmt.Sprintf("Error saving: %v", err))
				} else {
					m.setStatus(fmt.Sprintf("Added state '%s' (saved)", stateName))
				}
			}
			m.textinput.Blur()
			m.mode = modeSettings
			return m, nil

		default:
			var cmd tea.Cmd
			m.textinput, cmd = m.textinput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// viewSettingsAddState renders the add state view
func (m *uiModel) viewSettingsAddState() string {
	var content strings.Builder

	content.WriteString(m.styles.titleStyle.Render("Add New State") + "\n\n")
	content.WriteString(m.textinput.View() + "\n\n")
	content.WriteString(m.styles.statusStyle.Render("Enter state name • Press Enter to add • ESC to cancel") + "\n")

	return content.String()
}

// moveSettingsItemUp moves the current settings item up and auto-saves
func (m *uiModel) moveSettingsItemUp() {
	switch m.settingsSection {
	case settingsSectionTags:
		if m.settingsCursor > 0 && m.settingsCursor < len(m.config.Tags.Tags) {
			// Swap with previous item
			m.config.Tags.Tags[m.settingsCursor], m.config.Tags.Tags[m.settingsCursor-1] =
				m.config.Tags.Tags[m.settingsCursor-1], m.config.Tags.Tags[m.settingsCursor]
			m.settingsCursor--
			// Auto-save
			if err := m.config.Save(); err != nil {
				m.setStatus(fmt.Sprintf("Error saving: %v", err))
			} else {
				m.setStatus("Reordered (saved)")
			}
		}

	case settingsSectionStates:
		// Cannot reorder the default state setting (first item)
		if m.settingsCursor <= 1 {
			return
		}

		// Adjust for the default state setting offset
		stateIndex := m.settingsCursor - 1
		if stateIndex > 0 && stateIndex < len(m.config.States.States) {
			// Swap with previous item
			m.config.States.States[stateIndex], m.config.States.States[stateIndex-1] =
				m.config.States.States[stateIndex-1], m.config.States.States[stateIndex]
			m.settingsCursor--
			// Auto-save
			if err := m.config.Save(); err != nil {
				m.setStatus(fmt.Sprintf("Error saving: %v", err))
			} else {
				m.setStatus("Reordered (saved)")
			}
		}

	case settingsSectionKeybindings:
		// Keybindings cannot be reordered
		return
	}
}

// moveSettingsItemDown moves the current settings item down and auto-saves
func (m *uiModel) moveSettingsItemDown() {
	switch m.settingsSection {
	case settingsSectionTags:
		if m.settingsCursor >= 0 && m.settingsCursor < len(m.config.Tags.Tags)-1 {
			// Swap with next item
			m.config.Tags.Tags[m.settingsCursor], m.config.Tags.Tags[m.settingsCursor+1] =
				m.config.Tags.Tags[m.settingsCursor+1], m.config.Tags.Tags[m.settingsCursor]
			m.settingsCursor++
			// Auto-save
			if err := m.config.Save(); err != nil {
				m.setStatus(fmt.Sprintf("Error saving: %v", err))
			} else {
				m.setStatus("Reordered (saved)")
			}
		}

	case settingsSectionStates:
		// Cannot reorder the default state setting (first item)
		if m.settingsCursor == 0 {
			return
		}

		// Adjust for the default state setting offset
		stateIndex := m.settingsCursor - 1
		if stateIndex >= 0 && stateIndex < len(m.config.States.States)-1 {
			// Swap with next item
			m.config.States.States[stateIndex], m.config.States.States[stateIndex+1] =
				m.config.States.States[stateIndex+1], m.config.States.States[stateIndex]
			m.settingsCursor++
			// Auto-save
			if err := m.config.Save(); err != nil {
				m.setStatus(fmt.Sprintf("Error saving: %v", err))
			} else {
				m.setStatus("Reordered (saved)")
			}
		}

	case settingsSectionKeybindings:
		// Keybindings cannot be reordered
		return
	}
}
