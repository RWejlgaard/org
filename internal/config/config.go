package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/bubbles/key"
)

// Config represents the application configuration
type Config struct {
	Keybindings KeybindingsConfig `toml:"keybindings"`
	Colors      ColorsConfig      `toml:"colors"`
	Tags        TagsConfig        `toml:"tags"`
	States      StatesConfig      `toml:"states"`
	UI          UIConfig          `toml:"ui"`
}

// KeybindingsConfig holds all keybinding configurations
type KeybindingsConfig struct {
	Up            []string `toml:"up"`
	Down          []string `toml:"down"`
	Left          []string `toml:"left"`
	Right         []string `toml:"right"`
	ShiftUp       []string `toml:"shift_up"`
	ShiftDown     []string `toml:"shift_down"`
	ShiftLeft     []string `toml:"shift_left"`
	ShiftRight    []string `toml:"shift_right"`
	Rename        []string `toml:"rename"`
	CycleState    []string `toml:"cycle_state"`
	ToggleFold    []string `toml:"toggle_fold"`
	ToggleFoldAll []string `toml:"toggle_fold_all"`
	EditNotes     []string `toml:"edit_notes"`
	ToggleView    []string `toml:"toggle_view"`
	Capture       []string `toml:"capture"`
	AddSubTask    []string `toml:"add_subtask"`
	Delete        []string `toml:"delete"`
	Save          []string `toml:"save"`
	ToggleReorder []string `toml:"toggle_reorder"`
	ClockIn       []string `toml:"clock_in"`
	ClockOut      []string `toml:"clock_out"`
	SetDeadline   []string `toml:"set_deadline"`
	SetScheduled  []string `toml:"set_scheduled"`
	SetPriority   []string `toml:"set_priority"`
	SetEffort     []string `toml:"set_effort"`
	Help          []string `toml:"help"`
	Quit          []string `toml:"quit"`
	Settings      []string `toml:"settings"`
	TagItem       []string `toml:"tag_item"`
}

// ColorsConfig holds color configurations
type ColorsConfig struct {
	Todo      string `toml:"todo"`
	Progress  string `toml:"progress"`
	Blocked   string `toml:"blocked"`
	Done      string `toml:"done"`
	Cursor    string `toml:"cursor"`
	Title     string `toml:"title"`
	Scheduled string `toml:"scheduled"`
	Overdue   string `toml:"overdue"`
	Status    string `toml:"status"`
	Note      string `toml:"note"`
	Folded    string `toml:"folded"`
}

// TagConfig represents a single tag configuration
type TagConfig struct {
	Name  string `toml:"name"`
	Color string `toml:"color"`
}

// TagsConfig holds tag configurations
type TagsConfig struct {
	Enabled    bool        `toml:"enabled"`
	Tags       []TagConfig `toml:"tags"`
	DefaultTag string      `toml:"default_tag"`
}

// StateConfig represents a single TODO state configuration
type StateConfig struct {
	Name  string `toml:"name"`
	Color string `toml:"color"`
}

// StatesConfig holds TODO state configurations
type StatesConfig struct {
	States              []StateConfig `toml:"states"`
	DefaultNewTaskState string        `toml:"default_new_task_state"`
}

// UIConfig holds UI-related configurations
type UIConfig struct {
	HelpTextWidth         int    `toml:"help_text_width"`
	MinTerminalWidth      int    `toml:"min_terminal_width"`
	AgendaDays            int    `toml:"agenda_days"`
	OrgSyntaxHighlighting bool   `toml:"org_syntax_highlighting"`
	ShowIndentationGuides bool   `toml:"show_indentation_guides"`
	IndentationGuideColor string `toml:"indentation_guide_color"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Keybindings: KeybindingsConfig{
			Up:            []string{"up", "k"},
			Down:          []string{"down", "j"},
			Left:          []string{"left", "h"},
			Right:         []string{"right", "l"},
			ShiftUp:       []string{"shift+up"},
			ShiftDown:     []string{"shift+down"},
			ShiftLeft:     []string{"shift+left"},
			ShiftRight:    []string{"shift+right"},
			Rename:        []string{"R"},
			CycleState:    []string{"t", " "},
			ToggleFold:    []string{"tab"},
			ToggleFoldAll: []string{"shift+tab", "backtab"},
			EditNotes:     []string{"enter"},
			ToggleView:    []string{"a"},
			Capture:       []string{"c"},
			AddSubTask:    []string{"s"},
			Delete:        []string{"D"},
			Save:          []string{"ctrl+s"},
			ToggleReorder: []string{"r"},
			ClockIn:       []string{"i"},
			ClockOut:      []string{"o"},
			SetDeadline:   []string{"d"},
			SetScheduled:  []string{"S"},
			SetPriority:   []string{"p"},
			SetEffort:     []string{"e"},
			Help:          []string{"?"},
			Quit:          []string{"q", "ctrl+c"},
			Settings:      []string{","},
			TagItem:       []string{"#"},
		},
		Colors: ColorsConfig{
			Todo:      "202",
			Progress:  "220",
			Blocked:   "196",
			Done:      "34",
			Cursor:    "240",
			Title:     "99",
			Scheduled: "141",
			Overdue:   "196",
			Status:    "241",
			Note:      "246",
			Folded:    "243",
		},
		Tags: TagsConfig{
			Enabled:    true,
			DefaultTag: "work",
			Tags: []TagConfig{
				{Name: "work", Color: "99"},
				{Name: "personal", Color: "141"},
				{Name: "urgent", Color: "196"},
				{Name: "important", Color: "220"},
			},
		},
		States: StatesConfig{
			States: []StateConfig{
				{Name: "TODO", Color: "202"},
				{Name: "PROG", Color: "220"},
				{Name: "BLOCK", Color: "196"},
				{Name: "DONE", Color: "34"},
			},
			DefaultNewTaskState: "TODO",
		},
		UI: UIConfig{
			HelpTextWidth:         22,
			MinTerminalWidth:      40,
			AgendaDays:            7,
			OrgSyntaxHighlighting: true,
			ShowIndentationGuides: true,
			IndentationGuideColor: "245",
		},
	}
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	orgConfigDir := filepath.Join(configDir, "org")
	configPath := filepath.Join(orgConfigDir, "config.toml")

	return configPath, nil
}

// LoadConfig loads the configuration from the config file
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, create it with defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultCfg := DefaultConfig()
		if err := defaultCfg.Save(); err != nil {
			// If we can't save, just return defaults
			return defaultCfg, nil
		}
		return defaultCfg, nil
	}

	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge with defaults for any missing values
	config.fillDefaults()

	return &config, nil
}

// Save saves the configuration to the config file
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create or truncate the file
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	// Encode config to TOML
	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// fillDefaults fills in any missing config values with defaults
func (c *Config) fillDefaults() {
	defaults := DefaultConfig()

	// Fill keybindings if empty
	if len(c.Keybindings.Up) == 0 {
		c.Keybindings.Up = defaults.Keybindings.Up
	}
	if len(c.Keybindings.Down) == 0 {
		c.Keybindings.Down = defaults.Keybindings.Down
	}
	if len(c.Keybindings.Left) == 0 {
		c.Keybindings.Left = defaults.Keybindings.Left
	}
	if len(c.Keybindings.Right) == 0 {
		c.Keybindings.Right = defaults.Keybindings.Right
	}
	if len(c.Keybindings.ShiftUp) == 0 {
		c.Keybindings.ShiftUp = defaults.Keybindings.ShiftUp
	}
	if len(c.Keybindings.ShiftDown) == 0 {
		c.Keybindings.ShiftDown = defaults.Keybindings.ShiftDown
	}
	if len(c.Keybindings.ShiftLeft) == 0 {
		c.Keybindings.ShiftLeft = defaults.Keybindings.ShiftLeft
	}
	if len(c.Keybindings.ShiftRight) == 0 {
		c.Keybindings.ShiftRight = defaults.Keybindings.ShiftRight
	}
	if len(c.Keybindings.Rename) == 0 {
		c.Keybindings.Rename = defaults.Keybindings.Rename
	}
	if len(c.Keybindings.CycleState) == 0 {
		c.Keybindings.CycleState = defaults.Keybindings.CycleState
	}
	if len(c.Keybindings.ToggleFold) == 0 {
		c.Keybindings.ToggleFold = defaults.Keybindings.ToggleFold
	}
	if len(c.Keybindings.ToggleFoldAll) == 0 {
		c.Keybindings.ToggleFoldAll = defaults.Keybindings.ToggleFoldAll
	}
	if len(c.Keybindings.EditNotes) == 0 {
		c.Keybindings.EditNotes = defaults.Keybindings.EditNotes
	}
	if len(c.Keybindings.ToggleView) == 0 {
		c.Keybindings.ToggleView = defaults.Keybindings.ToggleView
	}
	if len(c.Keybindings.Capture) == 0 {
		c.Keybindings.Capture = defaults.Keybindings.Capture
	}
	if len(c.Keybindings.AddSubTask) == 0 {
		c.Keybindings.AddSubTask = defaults.Keybindings.AddSubTask
	}
	if len(c.Keybindings.Delete) == 0 {
		c.Keybindings.Delete = defaults.Keybindings.Delete
	}
	if len(c.Keybindings.Save) == 0 {
		c.Keybindings.Save = defaults.Keybindings.Save
	}
	if len(c.Keybindings.ToggleReorder) == 0 {
		c.Keybindings.ToggleReorder = defaults.Keybindings.ToggleReorder
	}
	if len(c.Keybindings.ClockIn) == 0 {
		c.Keybindings.ClockIn = defaults.Keybindings.ClockIn
	}
	if len(c.Keybindings.ClockOut) == 0 {
		c.Keybindings.ClockOut = defaults.Keybindings.ClockOut
	}
	if len(c.Keybindings.SetDeadline) == 0 {
		c.Keybindings.SetDeadline = defaults.Keybindings.SetDeadline
	}
	if len(c.Keybindings.SetScheduled) == 0 {
		c.Keybindings.SetScheduled = defaults.Keybindings.SetScheduled
	}
	if len(c.Keybindings.SetPriority) == 0 {
		c.Keybindings.SetPriority = defaults.Keybindings.SetPriority
	}
	if len(c.Keybindings.SetEffort) == 0 {
		c.Keybindings.SetEffort = defaults.Keybindings.SetEffort
	}
	if len(c.Keybindings.Help) == 0 {
		c.Keybindings.Help = defaults.Keybindings.Help
	}
	if len(c.Keybindings.Quit) == 0 {
		c.Keybindings.Quit = defaults.Keybindings.Quit
	}
	if len(c.Keybindings.Settings) == 0 {
		c.Keybindings.Settings = defaults.Keybindings.Settings
	}
	if len(c.Keybindings.TagItem) == 0 {
		c.Keybindings.TagItem = defaults.Keybindings.TagItem
	}

	// Fill colors if empty
	if c.Colors.Todo == "" {
		c.Colors.Todo = defaults.Colors.Todo
	}
	if c.Colors.Progress == "" {
		c.Colors.Progress = defaults.Colors.Progress
	}
	if c.Colors.Blocked == "" {
		c.Colors.Blocked = defaults.Colors.Blocked
	}
	if c.Colors.Done == "" {
		c.Colors.Done = defaults.Colors.Done
	}
	if c.Colors.Cursor == "" {
		c.Colors.Cursor = defaults.Colors.Cursor
	}
	if c.Colors.Title == "" {
		c.Colors.Title = defaults.Colors.Title
	}
	if c.Colors.Scheduled == "" {
		c.Colors.Scheduled = defaults.Colors.Scheduled
	}
	if c.Colors.Overdue == "" {
		c.Colors.Overdue = defaults.Colors.Overdue
	}
	if c.Colors.Status == "" {
		c.Colors.Status = defaults.Colors.Status
	}
	if c.Colors.Note == "" {
		c.Colors.Note = defaults.Colors.Note
	}
	if c.Colors.Folded == "" {
		c.Colors.Folded = defaults.Colors.Folded
	}

	// Fill tags if empty
	if len(c.Tags.Tags) == 0 {
		c.Tags.Tags = defaults.Tags.Tags
	}
	if c.Tags.DefaultTag == "" {
		c.Tags.DefaultTag = defaults.Tags.DefaultTag
	}

	// Fill states if empty
	if len(c.States.States) == 0 {
		c.States.States = defaults.States.States
		// Also set the default new task state since the entire states section is missing
		c.States.DefaultNewTaskState = defaults.States.DefaultNewTaskState
	}
	// Note: We don't fill DefaultNewTaskState if States.States is non-empty because
	// an empty string is a valid intentional value meaning "no default state".

	// Fill UI if zero values
	if c.UI.HelpTextWidth == 0 {
		c.UI.HelpTextWidth = defaults.UI.HelpTextWidth
	}
	if c.UI.MinTerminalWidth == 0 {
		c.UI.MinTerminalWidth = defaults.UI.MinTerminalWidth
	}
	if c.UI.AgendaDays == 0 {
		c.UI.AgendaDays = defaults.UI.AgendaDays
	}
	if c.UI.IndentationGuideColor == "" {
		c.UI.IndentationGuideColor = defaults.UI.IndentationGuideColor
	}
}

// BuildKeyBinding creates a key.Binding from config
func BuildKeyBinding(keys []string, help string, description string) key.Binding {
	return key.NewBinding(
		key.WithKeys(keys...),
		key.WithHelp(help, description),
	)
}

// GetTagColor returns the color for a given tag name
func (c *Config) GetTagColor(tagName string) string {
	for _, tag := range c.Tags.Tags {
		if tag.Name == tagName {
			return tag.Color
		}
	}
	// Return a default color if tag not found
	return "99"
}

// AddTag adds a new tag to the configuration
func (c *Config) AddTag(name, color string) {
	// Check if tag already exists
	for i, tag := range c.Tags.Tags {
		if tag.Name == name {
			c.Tags.Tags[i].Color = color
			return
		}
	}
	c.Tags.Tags = append(c.Tags.Tags, TagConfig{Name: name, Color: color})
}

// RemoveTag removes a tag from the configuration
func (c *Config) RemoveTag(name string) {
	for i, tag := range c.Tags.Tags {
		if tag.Name == name {
			c.Tags.Tags = append(c.Tags.Tags[:i], c.Tags.Tags[i+1:]...)
			return
		}
	}
}

// UpdateTagColor updates the color of an existing tag
func (c *Config) UpdateTagColor(name, color string) {
	for i, tag := range c.Tags.Tags {
		if tag.Name == name {
			c.Tags.Tags[i].Color = color
			return
		}
	}
}

// GetStateColor returns the color for a given state name
func (c *Config) GetStateColor(stateName string) string {
	for _, state := range c.States.States {
		if state.Name == stateName {
			return state.Color
		}
	}
	// Return a default color if state not found
	return "99"
}

// AddState adds a new state to the configuration
func (c *Config) AddState(name, color string) {
	// Check if state already exists
	for i, state := range c.States.States {
		if state.Name == name {
			c.States.States[i].Color = color
			return
		}
	}
	c.States.States = append(c.States.States, StateConfig{Name: name, Color: color})
}

// RemoveState removes a state from the configuration
func (c *Config) RemoveState(name string) {
	for i, state := range c.States.States {
		if state.Name == name {
			c.States.States = append(c.States.States[:i], c.States.States[i+1:]...)
			return
		}
	}
}

// UpdateStateColor updates the color of an existing state
func (c *Config) UpdateStateColor(name, color string) {
	for i, state := range c.States.States {
		if state.Name == name {
			c.States.States[i].Color = color
			return
		}
	}
}

// GetStateNames returns all configured state names
func (c *Config) GetStateNames() []string {
	names := make([]string, len(c.States.States))
	for i, state := range c.States.States {
		names[i] = state.Name
	}
	return names
}

// UpdateKeybinding updates a keybinding in the configuration
func (c *Config) UpdateKeybinding(action string, keys []string) error {
	// Use reflection would be complex, so we handle specific cases
	switch action {
	case "up":
		c.Keybindings.Up = keys
	case "down":
		c.Keybindings.Down = keys
	case "left":
		c.Keybindings.Left = keys
	case "right":
		c.Keybindings.Right = keys
	case "cycle_state":
		c.Keybindings.CycleState = keys
	case "toggle_fold":
		c.Keybindings.ToggleFold = keys
	case "toggle_fold_all":
		c.Keybindings.ToggleFoldAll = keys
	case "edit_notes":
		c.Keybindings.EditNotes = keys
	case "capture":
		c.Keybindings.Capture = keys
	case "add_subtask":
		c.Keybindings.AddSubTask = keys
	case "delete":
		c.Keybindings.Delete = keys
	case "tag_item":
		c.Keybindings.TagItem = keys
	case "settings":
		c.Keybindings.Settings = keys
	case "toggle_view":
		c.Keybindings.ToggleView = keys
	case "save":
		c.Keybindings.Save = keys
	case "help":
		c.Keybindings.Help = keys
	case "quit":
		c.Keybindings.Quit = keys
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
	return nil
}

// GetAllKeybindings returns a map of all keybindings
func (c *Config) GetAllKeybindings() map[string][]string {
	return map[string][]string{
		"up":              c.Keybindings.Up,
		"down":            c.Keybindings.Down,
		"left":            c.Keybindings.Left,
		"right":           c.Keybindings.Right,
		"shift_up":        c.Keybindings.ShiftUp,
		"shift_down":      c.Keybindings.ShiftDown,
		"shift_left":      c.Keybindings.ShiftLeft,
		"shift_right":     c.Keybindings.ShiftRight,
		"rename":          c.Keybindings.Rename,
		"cycle_state":     c.Keybindings.CycleState,
		"toggle_fold":     c.Keybindings.ToggleFold,
		"toggle_fold_all": c.Keybindings.ToggleFoldAll,
		"edit_notes":      c.Keybindings.EditNotes,
		"toggle_view":     c.Keybindings.ToggleView,
		"capture":         c.Keybindings.Capture,
		"add_subtask":     c.Keybindings.AddSubTask,
		"delete":          c.Keybindings.Delete,
		"save":            c.Keybindings.Save,
		"toggle_reorder":  c.Keybindings.ToggleReorder,
		"clock_in":        c.Keybindings.ClockIn,
		"clock_out":       c.Keybindings.ClockOut,
		"set_deadline":    c.Keybindings.SetDeadline,
		"set_scheduled":   c.Keybindings.SetScheduled,
		"set_priority":    c.Keybindings.SetPriority,
		"set_effort":      c.Keybindings.SetEffort,
		"help":            c.Keybindings.Help,
		"quit":            c.Keybindings.Quit,
		"settings":        c.Keybindings.Settings,
		"tag_item":        c.Keybindings.TagItem,
	}
}

// GetDefaultNewTaskState returns the default state for new tasks
// Returns empty string if configured as "none" or if the configured state doesn't exist
func (c *Config) GetDefaultNewTaskState() string {
	// Empty string means no state
	if c.States.DefaultNewTaskState == "" {
		return ""
	}

	// Validate that the configured state exists
	for _, state := range c.States.States {
		if state.Name == c.States.DefaultNewTaskState {
			return c.States.DefaultNewTaskState
		}
	}

	// If configured state doesn't exist, fall back to first state or empty
	if len(c.States.States) > 0 {
		return c.States.States[0].Name
	}

	return ""
}
