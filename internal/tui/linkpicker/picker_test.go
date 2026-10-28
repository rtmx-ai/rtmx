package linkpicker

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewPreselectsHint(t *testing.T) {
	m := New("gitlab")
	if m.Cursor() != 4 { // github,jira,asana,monday,gitlab
		t.Fatalf("cursor = %d, want 4 (GitLab)", m.Cursor())
	}
	m = New("github")
	if m.Cursor() != 0 {
		t.Fatalf("cursor = %d, want 0 (GitHub)", m.Cursor())
	}
	m = New("")
	if m.Cursor() != 0 {
		t.Fatalf("empty hint cursor = %d, want 0", m.Cursor())
	}
	m = New("unknown")
	if m.Cursor() != 0 {
		t.Fatalf("unknown hint cursor = %d, want 0", m.Cursor())
	}
}

func TestPickerConfirmSelection(t *testing.T) {
	m := New("github")
	// move to jira
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("expected Quit cmd")
	}
	if !m.Done() || m.Chosen() != "jira" {
		t.Fatalf("chosen=%q done=%v, want jira", m.Chosen(), m.Done())
	}
}

func TestPickerSkip(t *testing.T) {
	m := New("")
	for i := 0; i < len(DefaultChoices)-1; i++ {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updated.(Model)
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if m.Chosen() != "skip" {
		t.Fatalf("chosen=%q, want skip", m.Chosen())
	}
}

func TestPickerCancel(t *testing.T) {
	m := New("github")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if !m.Done() || m.Chosen() != "" {
		t.Fatalf("cancel: chosen=%q done=%v", m.Chosen(), m.Done())
	}
}

func TestDefaultChoicesIDs(t *testing.T) {
	want := []string{"github", "jira", "asana", "monday", "gitlab", "tests", "skip"}
	if len(DefaultChoices) != len(want) {
		t.Fatalf("len=%d want %d", len(DefaultChoices), len(want))
	}
	for i, id := range want {
		if DefaultChoices[i].ID != id {
			t.Fatalf("[%d]=%q want %q", i, DefaultChoices[i].ID, id)
		}
	}
}

func TestViewContainsLabels(t *testing.T) {
	view := New("github").View()
	for _, label := range []string{"GitHub", "Jira", "Skip", "Link an issue store"} {
		if !strings.Contains(view, label) {
			t.Fatalf("view missing %q:\n%s", label, view)
		}
	}
}
