package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	datepicker "github.com/ethanefung/bubble-datepicker"
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
