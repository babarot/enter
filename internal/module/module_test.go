package module

import "testing"

func TestNewSegment(t *testing.T) {
	seg := NewSegment("hello", Primary)
	if seg.Text != "hello" {
		t.Errorf("got %q, want %q", seg.Text, "hello")
	}
	if seg.Color != Primary {
		t.Errorf("got %v, want %v", seg.Color, Primary)
	}
}

func TestPlain(t *testing.T) {
	seg := Plain("text")
	if seg.Text != "text" {
		t.Errorf("got %q, want %q", seg.Text, "text")
	}
	if seg.Color != Default {
		t.Errorf("got %v, want %v", seg.Color, Default)
	}
}
