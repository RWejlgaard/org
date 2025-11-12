package model

import "time"

// Priority represents org-mode priority levels
type Priority string

const (
	PriorityNone Priority = ""
	PriorityA    Priority = "A"
	PriorityB    Priority = "B"
	PriorityC    Priority = "C"
)

// Item represents a single org-mode item (heading)
type Item struct {
	Level        int          // Heading level (number of *)
	State        TodoState    // TODO, PROG, BLOCK, DONE, or empty
	Priority     Priority     // Priority: A, B, C, or empty
	Title        string       // The main title text
	Tags         []string     // Tags for this item (e.g., :work:urgent:)
	Scheduled    *time.Time
	Deadline     *time.Time
	Closed       *time.Time   // Closed timestamp (when task was marked as done)
	Effort       string       // Effort estimate (e.g., "8h", "2d")
	Notes        []string     // Notes/content under the heading
	Children     []*Item      // Sub-items
	Folded       bool         // Whether the item is folded (hides notes and children)
	ClockEntries []ClockEntry // Clock in/out entries
	SourceFile   string       // Source file path (used in multi-file mode)
}

// OrgFile represents a parsed org-mode file
type OrgFile struct {
	Path  string
	Items []*Item
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
