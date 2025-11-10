package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/rwejlgaard/org/internal/model"
)

// Save writes the org file back to disk
func Save(orgFile *model.OrgFile) error {
	// Check if this is a multi-file org (directory-based)
	// In multi-file mode, top-level items have SourceFile set and represent files
	isMultiFile := false
	if len(orgFile.Items) > 0 && orgFile.Items[0].SourceFile != "" {
		isMultiFile = true
	}

	if isMultiFile {
		return saveMultiFile(orgFile)
	}

	// Single file mode
	file, err := os.Create(orgFile.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, item := range orgFile.Items {
		if err := writeItem(writer, item); err != nil {
			return err
		}
	}

	return nil
}

// saveMultiFile saves items back to their individual source files
func saveMultiFile(orgFile *model.OrgFile) error {
	// Group items by source file
	fileItems := make(map[string][]*model.Item)

	for _, fileItem := range orgFile.Items {
		if fileItem.SourceFile == "" {
			continue
		}

		// The children of this file item are the actual items to save
		fileItems[fileItem.SourceFile] = fileItem.Children
	}

	// Save each file
	for filePath, items := range fileItems {
		if err := saveItemsToFile(filePath, items); err != nil {
			return err
		}
	}

	return nil
}

// saveItemsToFile writes a list of items to a specific file
func saveItemsToFile(filePath string, items []*model.Item) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, item := range items {
		// Decrement level since we're saving to individual files
		decrementedItem := decrementItemLevelForSave(item)
		if err := writeItem(writer, decrementedItem); err != nil {
			return err
		}
	}

	return nil
}

// decrementItemLevelForSave creates a copy of an item with decremented levels for saving
func decrementItemLevelForSave(item *model.Item) *model.Item {
	copied := *item
	copied.Level--

	copiedChildren := make([]*model.Item, len(item.Children))
	for i, child := range item.Children {
		copiedChildren[i] = decrementItemLevelForSave(child)
	}
	copied.Children = copiedChildren

	return &copied
}

// writeItem recursively writes an item and its children
func writeItem(writer *bufio.Writer, item *model.Item) error {
	// Write heading
	stars := strings.Repeat("*", item.Level)
	line := stars
	if item.State != model.StateNone {
		line += " " + string(item.State)
	}
	if item.Priority != model.PriorityNone {
		line += " [#" + string(item.Priority) + "]"
	}
	line += " " + item.Title

	// Add tags if present
	if len(item.Tags) > 0 {
		line += " :" + strings.Join(item.Tags, ":") + ":"
	}

	line += "\n"

	if _, err := writer.WriteString(line); err != nil {
		return err
	}

	// Write scheduling info if not already in notes
	hasScheduled := false
	hasDeadline := false
	hasLogbook := false
	hasProperties := false
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
		if strings.Contains(note, ":PROPERTIES:") {
			hasProperties = true
		}
	}

	if item.Scheduled != nil && !hasScheduled {
		scheduledLine := fmt.Sprintf("SCHEDULED: <%s>\n", FormatOrgDate(*item.Scheduled))
		if _, err := writer.WriteString(scheduledLine); err != nil {
			return err
		}
	}

	if item.Deadline != nil && !hasDeadline {
		deadlineLine := fmt.Sprintf("DEADLINE: <%s>\n", FormatOrgDate(*item.Deadline))
		if _, err := writer.WriteString(deadlineLine); err != nil {
			return err
		}
	}

	// Write effort in :PROPERTIES: drawer if not already in notes
	if item.Effort != "" && !hasProperties {
		if _, err := writer.WriteString(":PROPERTIES:\n"); err != nil {
			return err
		}
		effortLine := fmt.Sprintf(":EFFORT: %s\n", item.Effort)
		if _, err := writer.WriteString(effortLine); err != nil {
			return err
		}
		if _, err := writer.WriteString(":END:\n"); err != nil {
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
