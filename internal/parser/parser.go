package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/rwejlgaard/org/internal/config"
	"github.com/rwejlgaard/org/internal/model"
)

// Parser patterns
var (
	scheduledPattern    = regexp.MustCompile(`SCHEDULED:\s*<([^>]+)>`)
	deadlinePattern     = regexp.MustCompile(`DEADLINE:\s*<([^>]+)>`)
	clockPattern        = regexp.MustCompile(`CLOCK:\s*\[([^\]]+)\](?:--\[([^\]]+)\])?`)
	effortPattern       = regexp.MustCompile(`^\s*:EFFORT:\s*(.+)$`)
	logbookDrawerStart  = regexp.MustCompile(`^\s*:LOGBOOK:\s*$`)
	propertiesDrawerStart = regexp.MustCompile(`^\s*:PROPERTIES:\s*$`)
	drawerEnd           = regexp.MustCompile(`^\s*:END:\s*$`)
	codeBlockStart      = regexp.MustCompile(`^\s*#\+BEGIN_SRC`)
	codeBlockEnd        = regexp.MustCompile(`^\s*#\+END_SRC`)
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
