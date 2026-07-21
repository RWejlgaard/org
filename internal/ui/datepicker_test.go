package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	datepicker "github.com/ethanefung/bubble-datepicker"
	"github.com/rwejlgaard/org/internal/config"
	"github.com/rwejlgaard/org/internal/model"
)

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func TestNewDatePickerDefaultsToToday(t *testing.T) {
	dp := newDatePicker(nil)
	if !sameDay(dp.Time, time.Now()) {
		t.Fatalf("expected today, got %v", dp.Time)
	}
	if !dp.Selected {
		t.Fatal("expected date to be selected")
	}
	if dp.Focused != datepicker.FocusCalendar {
		t.Fatalf("expected FocusCalendar, got %v", dp.Focused)
	}
	// Quit binding must be disabled so 'q' does not quit the app from the calendar.
	if dp.KeyMap.Quit.Enabled() && len(dp.KeyMap.Quit.Keys()) > 0 {
		t.Fatalf("expected Quit binding disabled, got keys %v", dp.KeyMap.Quit.Keys())
	}
	// Forwarding a 'q' key must not produce tea.Quit.
	_, cmd := dp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		t.Fatal("expected no command from 'q' after disabling Quit")
	}
}

func TestNewDatePickerUsesExisting(t *testing.T) {
	existing := time.Date(2027, time.March, 9, 0, 0, 0, 0, time.Local)
	dp := newDatePicker(&existing)
	if !sameDay(dp.Time, existing) {
		t.Fatalf("expected %v, got %v", existing, dp.Time)
	}
}

func newDateModeModel(dateType string, existing *time.Time) (uiModel, *model.Item) {
	cfg := config.DefaultConfig()
	item := &model.Item{Title: "Task", Level: 1, Deadline: existing, Scheduled: existing}
	of := &model.OrgFile{Items: []*model.Item{item}}
	m := InitialModel(of, cfg, false, "")
	m.editingItem = item
	m.dateTextFocused = false
	m.datepicker = newDatePicker(existing)
	if dateType == "DEADLINE" {
		m.mode = modeSetDeadline
	} else {
		m.mode = modeSetScheduled
	}
	return m, item
}

func TestCalendarEnterAppliesDate(t *testing.T) {
	m, item := newDateModeModel("DEADLINE", nil)
	m.datepicker.SetTime(time.Date(2027, time.March, 9, 0, 0, 0, 0, time.Local))
	res, _ := m.updateSetDate(tea.KeyMsg{Type: tea.KeyEnter}, "DEADLINE")
	rm := res.(uiModel)
	if item.Deadline == nil || !sameDay(*item.Deadline, time.Date(2027, time.March, 9, 0, 0, 0, 0, time.Local)) {
		t.Fatalf("expected deadline 2027-03-09, got %v", item.Deadline)
	}
	if rm.mode != modeList {
		t.Fatalf("expected mode to return to list, got %v", rm.mode)
	}
}

func TestCalendarXClearsDate(t *testing.T) {
	existing := time.Date(2027, time.March, 9, 0, 0, 0, 0, time.Local)
	m, item := newDateModeModel("SCHEDULED", &existing)
	res, _ := m.updateSetDate(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, "SCHEDULED")
	rm := res.(uiModel)
	if item.Scheduled != nil {
		t.Fatalf("expected scheduled cleared, got %v", item.Scheduled)
	}
	if rm.mode != modeList {
		t.Fatalf("expected mode to return to list, got %v", rm.mode)
	}
}

func TestDigitSwitchesToTextFocus(t *testing.T) {
	m, _ := newDateModeModel("DEADLINE", nil)
	res, _ := m.updateSetDate(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}, "DEADLINE")
	rm := res.(uiModel)
	if !rm.dateTextFocused {
		t.Fatal("expected text focus after typing a digit")
	}
	if rm.textinput.Value() != "2" {
		t.Fatalf("expected text seeded with '2', got %q", rm.textinput.Value())
	}
	if rm.mode != modeSetDeadline {
		t.Fatalf("expected to stay in deadline mode, got %v", rm.mode)
	}
}

func TestTextEnterParsesRelativeDate(t *testing.T) {
	m, item := newDateModeModel("DEADLINE", nil)
	m.dateTextFocused = true
	m.textinput.SetValue("+7")
	res, _ := m.updateSetDate(tea.KeyMsg{Type: tea.KeyEnter}, "DEADLINE")
	rm := res.(uiModel)
	want := time.Now().AddDate(0, 0, 7)
	if item.Deadline == nil || !sameDay(*item.Deadline, want) {
		t.Fatalf("expected deadline %v, got %v", want, item.Deadline)
	}
	if rm.mode != modeList {
		t.Fatalf("expected mode to return to list, got %v", rm.mode)
	}
}

func TestTextEnterInvalidStaysInTextFocus(t *testing.T) {
	m, item := newDateModeModel("DEADLINE", nil)
	m.dateTextFocused = true
	m.textinput.SetValue("not-a-date")
	res, _ := m.updateSetDate(tea.KeyMsg{Type: tea.KeyEnter}, "DEADLINE")
	rm := res.(uiModel)
	if item.Deadline != nil {
		t.Fatalf("expected no deadline set, got %v", item.Deadline)
	}
	if rm.mode != modeSetDeadline || !rm.dateTextFocused {
		t.Fatalf("expected to stay in text focus, got mode=%v textFocused=%v", rm.mode, rm.dateTextFocused)
	}
}

func TestTextEscReturnsToCalendar(t *testing.T) {
	m, _ := newDateModeModel("DEADLINE", nil)
	m.dateTextFocused = true
	m.textinput.SetValue("+3")
	res, _ := m.updateSetDate(tea.KeyMsg{Type: tea.KeyEsc}, "DEADLINE")
	rm := res.(uiModel)
	if rm.dateTextFocused {
		t.Fatal("expected to return to calendar focus")
	}
	if rm.mode != modeSetDeadline {
		t.Fatalf("expected to stay in deadline mode, got %v", rm.mode)
	}
	if rm.textinput.Value() != "" {
		t.Fatalf("expected text cleared, got %q", rm.textinput.Value())
	}
}

func TestCalendarEscCancels(t *testing.T) {
	existing := time.Date(2027, time.March, 9, 0, 0, 0, 0, time.Local)
	m, item := newDateModeModel("DEADLINE", &existing)
	res, _ := m.updateSetDate(tea.KeyMsg{Type: tea.KeyEsc}, "DEADLINE")
	rm := res.(uiModel)
	if item.Deadline == nil {
		t.Fatal("expected existing deadline untouched on cancel")
	}
	if rm.mode != modeList {
		t.Fatalf("expected mode to return to list, got %v", rm.mode)
	}
}

func TestOpeningDeadlineModeInitializesCalendar(t *testing.T) {
	cfg := config.DefaultConfig()
	item := &model.Item{Title: "Task", Level: 1}
	of := &model.OrgFile{Items: []*model.Item{item}}
	m := InitialModel(of, cfg, false, "")
	// default deadline key is "d"
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	rm := res.(uiModel)
	if rm.mode != modeSetDeadline {
		t.Fatalf("expected modeSetDeadline, got %v", rm.mode)
	}
	if !rm.datepicker.Selected {
		t.Fatal("expected datepicker to have a selected date")
	}
	if rm.dateTextFocused {
		t.Fatal("expected calendar focus (dateTextFocused false)")
	}
	if rm.editingItem != item {
		t.Fatal("expected editingItem to be the cursor item")
	}
}
