package config

import (
	"testing"

	"github.com/goccy/go-yaml"
)

func TestNewFieldPresent(t *testing.T) {
	f := NewField(GitCwdConfig{Style: "tree"})
	if !f.Present() {
		t.Error("NewField should be present")
	}
	if f.Get().Style != "tree" {
		t.Errorf("Get().Style: got %q, want %q", f.Get().Style, "tree")
	}
}

func TestFieldZeroValueNotPresent(t *testing.T) {
	var f Field[GitCwdConfig]
	if f.Present() {
		t.Error("zero value Field should not be present")
	}
	if f.Get().Style != "" {
		t.Errorf("Get() on zero value should return zero, got %q", f.Get().Style)
	}
}

func TestFieldSet(t *testing.T) {
	var f Field[GitCwdConfig]
	f.Set(GitCwdConfig{Style: "breadcrumb"})
	if !f.Present() {
		t.Error("Set should make field present")
	}
	if f.Get().Style != "breadcrumb" {
		t.Errorf("Get().Style: got %q, want %q", f.Get().Style, "breadcrumb")
	}
}

func TestFieldUnmarshalYAMLPresent(t *testing.T) {
	type wrapper struct {
		Cwd Field[GitCwdConfig] `yaml:"cwd"`
	}
	var w wrapper
	if err := yaml.Unmarshal([]byte(`cwd:
  style: tree
`), &w); err != nil {
		t.Fatal(err)
	}
	if !w.Cwd.Present() {
		t.Error("cwd should be present after unmarshal")
	}
	if w.Cwd.Get().Style != "tree" {
		t.Errorf("Style: got %q, want %q", w.Cwd.Get().Style, "tree")
	}
}

func TestFieldUnmarshalYAMLAbsent(t *testing.T) {
	type wrapper struct {
		Cwd    Field[GitCwdConfig]    `yaml:"cwd"`
		Status Field[GitStatusConfig] `yaml:"status"`
	}
	var w wrapper
	if err := yaml.Unmarshal([]byte(`cwd:
  style: tree
`), &w); err != nil {
		t.Fatal(err)
	}
	if !w.Cwd.Present() {
		t.Error("cwd should be present")
	}
	if w.Status.Present() {
		t.Error("status should not be present (not in YAML)")
	}
}

func TestFieldUnmarshalYAMLEmptyKey(t *testing.T) {
	// goccy/go-yaml does NOT call UnmarshalYAML for null/empty values,
	// so an empty key like "url:" results in Present() == false after unmarshal.
	// The Load() function compensates by calling ensureFieldsPresent() using
	// the key list from extractOrder().
	type wrapper struct {
		Url Field[GitUrlConfig] `yaml:"url"`
	}
	var w wrapper
	if err := yaml.Unmarshal([]byte(`url:
`), &w); err != nil {
		t.Fatal(err)
	}
	// Raw unmarshal: not present
	if w.Url.Present() {
		t.Error("raw unmarshal of empty key should not be present (goccy/go-yaml behavior)")
	}
	// After MarkPresent: present
	w.Url.MarkPresent()
	if !w.Url.Present() {
		t.Error("after MarkPresent should be present")
	}
}

func TestFieldUnmarshalYAMLDeprecatedEnabled(t *testing.T) {
	// Ensure that an "enabled" key inside a field struct is silently ignored
	// (GitUrlConfig is now an empty struct, unknown keys are ignored by goccy/go-yaml)
	type wrapper struct {
		Url Field[GitUrlConfig] `yaml:"url"`
	}
	var w wrapper
	if err := yaml.Unmarshal([]byte(`url:
  enabled: true
`), &w); err != nil {
		t.Fatal(err)
	}
	if !w.Url.Present() {
		t.Error("url should be present even with deprecated enabled key")
	}
}
