package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/rwejlgaard/org/internal/config"
	"github.com/rwejlgaard/org/internal/model"
)

// Parser patterns
var (
	scheduledPattern      = regexp.MustCompile(`SCHEDULED:\s*<([^>]+)>`)
	deadlinePattern       = regexp.MustCompile(`DEADLINE:\s*<([^>]+)>`)
	closedPattern         = regexp.MustCompile(`CLOSED:\s*\[([^\]]+)\]`)
	clockPattern          = regexp.MustCompile(`CLOCK:\s*\[([^\]]+)\](?:--\[([^\]]+)\])?`)
	effortPattern         = regexp.MustCompile(`^\s*:EFFORT:\s*(.+)$`)
	logbookDrawerStart    = regexp.MustCompile(`^\s*:LOGBOOK:\s*$`)
	propertiesDrawerStart = regexp.MustCompile(`^\s*:PROPERTIES:\s*$`)
	drawerEnd             = regexp.MustCompile(`^\s*:END:\s*$`)
	codeBlockStart        = regexp.MustCompile(`^\s*#\+BEGIN_SRC`)
	codeBlockEnd          = regexp.MustCompile(`^\s*#\+END_SRC`)
)

// buildHeadingPattern creates a regex pattern that matches configured states
func buildHeadingPattern(cfg *config.Config) *regexp.Regexp {
	stateNames := cfg.GetStateNames()
	var statesPattern string
	if len(stateNames) > 0 {
		// Escape state names and join with |
		escapedStates := make([]string, len(stateNames))
		for i, state := range stateNames {
			escapedStates[i] = regexp.QuoteMeta(state)
		}
		statesPattern = strings.Join(escapedStates, "|")
	} else {
		// Fallback to default states if none configured
		statesPattern = "TODO|PROG|BLOCK|DONE"
	}

	pattern := `^(\*+)\s+(?:(` + statesPattern + `)\s+)?(?:\[#([A-C])\]\s+)?(.+?)(?:\s+(:[[:alnum:]_@#%:]+:)\s*)?$`
	return regexp.MustCompile(pattern)
}

// ParseOrgFile reads and parses an org-mode file
func ParseOrgFile(path string, cfg *config.Config) (*model.OrgFile, error) {
	headingPattern := buildHeadingPattern(cfg)
	file, err := os.Open(path)
	if err != nil {
		// If file doesn't exist, return empty org file
		if os.IsNotExist(err) {
			return &model.OrgFile{Path: path, Items: []*model.Item{}}, nil
		}
		return nil, err
	}
	defer file.Close()

	orgFile := &model.OrgFile{Path: path, Items: []*model.Item{}}
	scanner := bufio.NewScanner(file)

	var currentItem *model.Item
	var itemStack []*model.Item // Stack to track parent items
	var inCodeBlock bool
	var inLogbookDrawer bool
	var inPropertiesDrawer bool

	for scanner.Scan() {
		line := scanner.Text()

		// Check for drawer boundaries
		if logbookDrawerStart.MatchString(line) {
			inLogbookDrawer = true
			if currentItem != nil {
				currentItem.Notes = append(currentItem.Notes, line)
			}
			continue
		}
		if propertiesDrawerStart.MatchString(line) {
			inPropertiesDrawer = true
			if currentItem != nil {
				currentItem.Notes = append(currentItem.Notes, line)
			}
			continue
		}
		if drawerEnd.MatchString(line) {
			if inLogbookDrawer {
				inLogbookDrawer = false
				if currentItem != nil {
					currentItem.Notes = append(currentItem.Notes, line)
				}
				continue
			}
			if inPropertiesDrawer {
				inPropertiesDrawer = false
				if currentItem != nil {
					currentItem.Notes = append(currentItem.Notes, line)
				}
				continue
			}
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
			state := model.TodoState(matches[2])
			priority := model.Priority(matches[3])
			title := strings.TrimSpace(matches[4])
			tagsStr := matches[5]

			// Parse tags from :tag1:tag2: format
			var tags []string
			if tagsStr != "" {
				tagsStr = strings.Trim(tagsStr, ":")
				if tagsStr != "" {
					tags = strings.Split(tagsStr, ":")
				}
			}

			item := &model.Item{
				Level:    level,
				State:    state,
				Priority: priority,
				Title:    title,
				Tags:     tags,
				Notes:    []string{},
				Children: []*model.Item{},
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

			// Check for CLOSED
			if matches := closedPattern.FindStringSubmatch(line); matches != nil {
				if t, err := parseClockTimestamp(matches[1]); err == nil {
					currentItem.Closed = &t
				}
			}

			// Check for EFFORT (inside PROPERTIES drawer)
			if matches := effortPattern.FindStringSubmatch(line); matches != nil {
				currentItem.Effort = strings.TrimSpace(matches[1])
			}

			// Check for CLOCK (can be inside or outside drawer)
			if matches := clockPattern.FindStringSubmatch(line); matches != nil {
				if startTime, err := parseClockTimestamp(matches[1]); err == nil {
					entry := model.ClockEntry{Start: startTime}
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

// ParseMultipleOrgFiles loads all .org files in a directory and wraps them as top-level items
func ParseMultipleOrgFiles(dirPath string, cfg *config.Config) (*model.OrgFile, error) {
	// Find all .org files in the directory
	matches, err := filepath.Glob(filepath.Join(dirPath, "*.org"))
	if err != nil {
		return nil, err
	}

	// Sort files alphabetically
	sort.Strings(matches)

	// Create a virtual org file
	multiOrgFile := &model.OrgFile{
		Path:  dirPath, // Store directory path
		Items: []*model.Item{},
	}

	// Parse each file and wrap it as a top-level item
	for _, filePath := range matches {
		orgFile, err := ParseOrgFile(filePath, cfg)
		if err != nil {
			// Skip files that can't be parsed
			continue
		}

		// Create a wrapper item for this file
		fileName := filepath.Base(filePath)
		fileItem := &model.Item{
			Level:      1,
			State:      model.StateNone,
			Priority:   model.PriorityNone,
			Title:      fileName,
			Tags:       []string{},
			Notes:      []string{},
			Children:   []*model.Item{},
			SourceFile: filePath,
		}

		// Increment the level of all items from this file and add as children
		for _, item := range orgFile.Items {
			incrementItemLevel(item)
			setSourceFileRecursive(item, filePath)
			fileItem.Children = append(fileItem.Children, item)
		}

		multiOrgFile.Items = append(multiOrgFile.Items, fileItem)
	}

	return multiOrgFile, nil
}

// incrementItemLevel recursively increments the level of an item and its children
func incrementItemLevel(item *model.Item) {
	item.Level++
	for _, child := range item.Children {
		incrementItemLevel(child)
	}
}

// setSourceFileRecursive sets the source file for an item and all its descendants
func setSourceFileRecursive(item *model.Item, filePath string) {
	item.SourceFile = filePath
	for _, child := range item.Children {
		setSourceFileRecursive(child, filePath)
	}
}
