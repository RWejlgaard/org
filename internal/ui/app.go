package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rwejlgaard/org/internal/config"
	"github.com/rwejlgaard/org/internal/model"
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
	modeSetPriority
	modeSetEffort
	modeHelp
	modeSettings
	modeTagEdit
	modeRename
)

type uiModel struct {
	orgFile        *model.OrgFile
	cursor         int
	scrollOffset   int // Track the scroll position
	helpScroll     int // Track scroll position in help mode
	mode           viewMode
	help           help.Model
	keys           keyMap
	styles         styleMap
	config         *config.Config
	width          int
	height         int
	statusMsg      string
	statusExpiry   time.Time
	editingItem    *model.Item
	textarea       textarea.Model
	textinput      textinput.Model
	itemToDelete     *model.Item
	reorderMode      bool
	settingsCursor   int             // Cursor position in settings view
	settingsScroll   int             // Scroll position in settings view
	settingsSection  settingsSection // Current settings section/tab
	captureCursor    int             // Store cursor position when entering capture mode
}

func InitialModel(orgFile *model.OrgFile, cfg *config.Config) uiModel {
	ta := textarea.New()
	ta.Placeholder = "Enter notes here (code blocks supported)..."
	ta.ShowLineNumbers = false

	ti := textinput.New()
	ti.Placeholder = "What needs doing?"
	ti.CharLimit = 200

	h := help.New()
	h.ShowAll = false

	return uiModel{
		orgFile:   orgFile,
		cursor:    0,
		mode:      modeList,
		help:      h,
		keys:      newKeyMapFromConfig(cfg),
		styles:    newStyleMapFromConfig(cfg),
		config:    cfg,
		textarea:  ta,
		textinput: ti,
	}
}

func (m uiModel) Init() tea.Cmd {
	return nil
}

func (m *uiModel) setStatus(msg string) {
	m.statusMsg = msg
	m.statusExpiry = time.Now().Add(3 * time.Second)
}

func (m uiModel) getVisibleItems() []*model.Item {
	if m.mode == modeAgenda {
		return m.getAgendaItems()
	}
	return m.orgFile.GetAllItems()
}

func (m *uiModel) updateScrollOffset(availableHeight int) {
	items := m.getVisibleItems()
	if len(items) == 0 {
		return
	}

	// Build line count for each item
	itemLineCount := make([]int, len(items))
	for i, item := range items {
		lineCount := 1 // The item itself
		if !item.Folded && len(item.Notes) > 0 && m.mode == modeList {
			// Count note lines with wrapping
			indent := strings.Repeat("  ", item.Level)
			noteIndent := indent + "  "
			filteredNotes := filterLogbookDrawer(item.Notes)
			wrappedNotes := wrapNoteLines(filteredNotes, m.width, noteIndent)
			highlightedNotes := renderNotesWithHighlighting(wrappedNotes)
			lineCount += len(highlightedNotes)
		}
		itemLineCount[i] = lineCount
	}

	// Calculate total lines up to cursor
	totalLinesBeforeCursor := 0
	for i := 0; i < m.cursor && i < len(itemLineCount); i++ {
		totalLinesBeforeCursor += itemLineCount[i]
	}

	// Adjust scroll offset to keep cursor visible
	if totalLinesBeforeCursor < m.scrollOffset {
		// Cursor is above visible area, scroll up
		m.scrollOffset = totalLinesBeforeCursor
	} else if totalLinesBeforeCursor >= m.scrollOffset+availableHeight {
		// Cursor is below visible area, scroll down
		m.scrollOffset = totalLinesBeforeCursor - availableHeight + 1
	}

	// Ensure scroll offset doesn't go negative
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// RunUI starts the terminal UI
func RunUI(orgFile *model.OrgFile, cfg *config.Config) error {
	p := tea.NewProgram(InitialModel(orgFile, cfg), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
