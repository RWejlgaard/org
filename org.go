package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// TodoState represents the state of a todo item
type TodoState string

const (
	StateTODO        TodoState = "TODO"
	StatePROG        TodoState = "PROG"
	StateBLOCK       TodoState = "BLOCK"
	StateDONE        TodoState = "DONE"
	StateNone        TodoState = ""
)

// ClockEntry represents a single clock entry
type ClockEntry struct {
	Start time.Time
	End   *time.Time // nil if currently clocked in
}

// Item represents a single org-mode item (heading)
type Item struct {
	Level       int          // Heading level (number of *)
	State       TodoState    // TODO, PROG, BLOCK, DONE, or empty
	Title       string       // The main title text
	Scheduled   *time.Time
	Deadline    *time.Time
	Notes       []string     // Notes/content under the heading
	Children    []*Item      // Sub-items
	Folded      bool         // Whether the item is folded (hides notes and children)
	ClockEntries []ClockEntry // Clock in/out entries
}

// OrgFile represents a parsed org-mode file
type OrgFile struct {
	Path  string
	Items []*Item
}

// Parser patterns
var (
	headingPattern   = regexp.MustCompile(`^(\*+)\s+(?:(TODO|PROG|BLOCK|DONE)\s+)?(.+)$`)
	scheduledPattern = regexp.MustCompile(`SCHEDULED:\s*<([^>]+)>`)
	deadlinePattern  = regexp.MustCompile(`DEADLINE:\s*<([^>]+)>`)
	clockPattern     = regexp.MustCompile(`CLOCK:\s*\[([^\]]+)\](?:--\[([^\]]+)\])?`)
	drawerStart      = regexp.MustCompile(`^\s*:LOGBOOK:\s*$`)
	drawerEnd        = regexp.MustCompile(`^\s*:END:\s*$`)
	codeBlockStart   = regexp.MustCompile(`^\s*#\+BEGIN_SRC`)
	codeBlockEnd     = regexp.MustCompile(`^\s*#\+END_SRC`)
)

// ParseOrgFile reads and parses an org-mode file
func ParseOrgFile(path string) (*OrgFile, error) {
	file, err := os.Open(path)
	if err != nil {
		// If file doesn't exist, return empty org file
		if os.IsNotExist(err) {
			return &OrgFile{Path: path, Items: []*Item{}}, nil
		}
		return nil, err
	}
	defer file.Close()

	orgFile := &OrgFile{Path: path, Items: []*Item{}}
	scanner := bufio.NewScanner(file)

	var currentItem *Item
	var itemStack []*Item // Stack to track parent items
	var inCodeBlock bool
	var inLogbookDrawer bool

	for scanner.Scan() {
		line := scanner.Text()

		// Check for drawer boundaries
		if drawerStart.MatchString(line) {
			inLogbookDrawer = true
			if currentItem != nil {
				currentItem.Notes = append(currentItem.Notes, line)
			}
			continue
		}
		if drawerEnd.MatchString(line) && inLogbookDrawer {
			inLogbookDrawer = false
			if currentItem != nil {
				currentItem.Notes = append(currentItem.Notes, line)
			}
			continue
		}

		// Check for code block boundaries
		if codeBlockStart.MatchString(line) {
			inCodeBlock = true
			if currentItem != nil {
				currentItem.Notes = append(currentItem.Notes, line)
			}
			continue
		}
		if codeBlockEnd.MatchString(line) {
			inCodeBlock = false
			if currentItem != nil {
				currentItem.Notes = append(currentItem.Notes, line)
			}
			continue
		}

		// If in code block, add line to notes
		if inCodeBlock {
			if currentItem != nil {
				currentItem.Notes = append(currentItem.Notes, line)
			}
			continue
		}

		// Try to match heading
		if matches := headingPattern.FindStringSubmatch(line); matches != nil {
			level := len(matches[1])
			state := TodoState(matches[2])
			title := matches[3]

			item := &Item{
				Level:    level,
				State:    state,
				Title:    title,
				Notes:    []string{},
				Children: []*Item{},
			}

			// Find parent based on level
			for len(itemStack) > 0 && itemStack[len(itemStack)-1].Level >= level {
				itemStack = itemStack[:len(itemStack)-1]
			}

			if len(itemStack) == 0 {
				// Top-level item
				orgFile.Items = append(orgFile.Items, item)
			} else {
				// Child item
				parent := itemStack[len(itemStack)-1]
				parent.Children = append(parent.Children, item)
			}

			itemStack = append(itemStack, item)
			currentItem = item
		} else if currentItem != nil {
			// This is content under the current item
			trimmed := strings.TrimSpace(line)

			// Check for SCHEDULED
			if matches := scheduledPattern.FindStringSubmatch(line); matches != nil {
				if t, err := parseOrgDate(matches[1]); err == nil {
					currentItem.Scheduled = &t
				}
			}

			// Check for DEADLINE
			if matches := deadlinePattern.FindStringSubmatch(line); matches != nil {
				if t, err := parseOrgDate(matches[1]); err == nil {
					currentItem.Deadline = &t
				}
			}

			// Check for CLOCK (can be inside or outside drawer)
			if matches := clockPattern.FindStringSubmatch(line); matches != nil {
				if startTime, err := parseClockTimestamp(matches[1]); err == nil {
					entry := ClockEntry{Start: startTime}
					if len(matches) > 2 && matches[2] != "" {
						if endTime, err := parseClockTimestamp(matches[2]); err == nil {
							entry.End = &endTime
						}
					}
					currentItem.ClockEntries = append(currentItem.ClockEntries, entry)
				}
			}

			// Add all lines as notes (including scheduling lines and drawer content for proper serialization)
			if trimmed != "" || len(currentItem.Notes) > 0 {
				currentItem.Notes = append(currentItem.Notes, line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return orgFile, nil
}

// parseOrgDate parses org-mode date format
func parseOrgDate(dateStr string) (time.Time, error) {
	// Org-mode format: 2024-01-15 Mon 10:00
	formats := []string{
		"2006-01-02 Mon 15:04",
		"2006-01-02 Mon",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// parseClockTimestamp parses org-mode clock timestamp format
func parseClockTimestamp(timestampStr string) (time.Time, error) {
	// Org-mode clock format: [2024-01-15 Mon 10:00]
	formats := []string{
		"2006-01-02 Mon 15:04",
		"2006-01-02 Mon 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timestampStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse clock timestamp: %s", timestampStr)
}

// formatClockTimestamp formats a time as org-mode clock timestamp
func formatClockTimestamp(t time.Time) string {
	return t.Format("2006-01-02 Mon 15:04")
}

// formatOrgDate formats a time as org-mode date
func formatOrgDate(t time.Time) string {
	return t.Format("2006-01-02 Mon")
}

// Save writes the org file back to disk
func (of *OrgFile) Save() error {
	file, err := os.Create(of.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, item := range of.Items {
		if err := writeItem(writer, item); err != nil {
			return err
		}
	}

	return nil
}

// writeItem recursively writes an item and its children
func writeItem(writer *bufio.Writer, item *Item) error {
	// Write heading
	stars := strings.Repeat("*", item.Level)
	line := stars
	if item.State != StateNone {
		line += " " + string(item.State)
	}
	line += " " + item.Title + "\n"

	if _, err := writer.WriteString(line); err != nil {
		return err
	}

	// Write scheduling info if not already in notes
	hasScheduled := false
	hasDeadline := false
	hasLogbook := false
	for _, note := range item.Notes {
		if strings.Contains(note, "SCHEDULED:") {
			hasScheduled = true
		}
		if strings.Contains(note, "DEADLINE:") {
			hasDeadline = true
		}
		if strings.Contains(note, ":LOGBOOK:") {
			hasLogbook = true
		}
	}

	if item.Scheduled != nil && !hasScheduled {
		scheduledLine := fmt.Sprintf("SCHEDULED: <%s>\n", formatOrgDate(*item.Scheduled))
		if _, err := writer.WriteString(scheduledLine); err != nil {
			return err
		}
	}

	if item.Deadline != nil && !hasDeadline {
		deadlineLine := fmt.Sprintf("DEADLINE: <%s>\n", formatOrgDate(*item.Deadline))
		if _, err := writer.WriteString(deadlineLine); err != nil {
			return err
		}
	}

	// Write clock entries in :LOGBOOK: drawer if not already in notes
	if len(item.ClockEntries) > 0 && !hasLogbook {
		if _, err := writer.WriteString(":LOGBOOK:\n"); err != nil {
			return err
		}
		for _, entry := range item.ClockEntries {
			clockLine := fmt.Sprintf("CLOCK: [%s]", formatClockTimestamp(entry.Start))
			if entry.End != nil {
				clockLine += fmt.Sprintf("--[%s]", formatClockTimestamp(*entry.End))
			}
			clockLine += "\n"
			if _, err := writer.WriteString(clockLine); err != nil {
				return err
			}
		}
		if _, err := writer.WriteString(":END:\n"); err != nil {
			return err
		}
	}

	// Write notes
	for _, note := range item.Notes {
		if _, err := writer.WriteString(note + "\n"); err != nil {
			return err
		}
	}

	// Write children
	for _, child := range item.Children {
		if err := writeItem(writer, child); err != nil {
			return err
		}
	}

	return nil
}

// GetAllItems returns a flattened list of all items (for UI display)
// Respects folding - folded items don't show their children
func (of *OrgFile) GetAllItems() []*Item {
	var items []*Item
	var flatten func([]*Item)
	flatten = func(list []*Item) {
		for _, item := range list {
			items = append(items, item)
			if !item.Folded {
				flatten(item.Children)
			}
		}
	}
	flatten(of.Items)
	return items
}

// ToggleFold toggles the folded state of an item
func (item *Item) ToggleFold() {
	item.Folded = !item.Folded
}

// CycleState cycles through todo states
func (item *Item) CycleState() {
	switch item.State {
	case StateNone:
		item.State = StateTODO
	case StateTODO:
		item.State = StatePROG
	case StatePROG:
		item.State = StateBLOCK
	case StateBLOCK:
		item.State = StateDONE
	case StateDONE:
		item.State = StateNone
	}
}

// ClockIn starts a new clock entry
func (item *Item) ClockIn() bool {
	// Check if already clocked in
	if item.IsClockedIn() {
		return false
	}

	entry := ClockEntry{
		Start: time.Now(),
		End:   nil,
	}
	item.ClockEntries = append(item.ClockEntries, entry)
	return true
}

// ClockOut ends the current clock entry
func (item *Item) ClockOut() bool {
	// Find the most recent open clock entry
	for i := len(item.ClockEntries) - 1; i >= 0; i-- {
		if item.ClockEntries[i].End == nil {
			now := time.Now()
			item.ClockEntries[i].End = &now
			return true
		}
	}
	return false
}

// IsClockedIn returns true if there's an active clock entry
func (item *Item) IsClockedIn() bool {
	for _, entry := range item.ClockEntries {
		if entry.End == nil {
			return true
		}
	}
	return false
}

// GetCurrentClockDuration returns the duration of the current clock entry
func (item *Item) GetCurrentClockDuration() time.Duration {
	for _, entry := range item.ClockEntries {
		if entry.End == nil {
			return time.Since(entry.Start)
		}
	}
	return 0
}

// GetTotalClockDuration returns the total duration of all clock entries
func (item *Item) GetTotalClockDuration() time.Duration {
	var total time.Duration
	for _, entry := range item.ClockEntries {
		if entry.End != nil {
			// Completed clock entry
			total += entry.End.Sub(entry.Start)
		} else {
			// Currently clocked in
			total += time.Since(entry.Start)
		}
	}
	return total
}
