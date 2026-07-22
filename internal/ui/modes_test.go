package ui

import (
	"testing"
	"time"

	"github.com/rwejlgaard/org/internal/model"
)

func TestApplyClosedTransitionEntersDone(t *testing.T) {
	now := time.Date(2026, 7, 21, 14, 30, 0, 0, time.UTC)
	item := &model.Item{Notes: []string{"a note", "CLOSED: [old stamp]"}}

	applyClosedTransition(item, false, true, now)

	if item.Closed == nil || !item.Closed.Equal(now) {
		t.Fatalf("expected Closed set to now, got %v", item.Closed)
	}
	if len(item.Notes) != 1 || item.Notes[0] != "a note" {
		t.Fatalf("expected stale CLOSED note stripped, got %v", item.Notes)
	}
}

func TestApplyClosedTransitionLeavesDone(t *testing.T) {
	now := time.Now()
	stamp := now
	item := &model.Item{Closed: &stamp, Notes: []string{"  CLOSED: [x]", "keep"}}

	applyClosedTransition(item, true, false, now)

	if item.Closed != nil {
		t.Fatalf("expected Closed cleared, got %v", item.Closed)
	}
	if len(item.Notes) != 1 || item.Notes[0] != "keep" {
		t.Fatalf("expected CLOSED note stripped, got %v", item.Notes)
	}
}

func TestApplyClosedTransitionDoneToDoneUnchanged(t *testing.T) {
	orig := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	item := &model.Item{Closed: &orig, Notes: []string{"keep"}}

	applyClosedTransition(item, true, true, time.Now())

	if item.Closed == nil || !item.Closed.Equal(orig) {
		t.Fatalf("done->done must keep original Closed, got %v", item.Closed)
	}
	if len(item.Notes) != 1 || item.Notes[0] != "keep" {
		t.Fatalf("done->done must not touch notes, got %v", item.Notes)
	}
}
