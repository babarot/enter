package config

import (
	"github.com/goccy/go-yaml"
)

// Field[T] represents an optional configuration field.
// Presence is determined by whether the YAML key exists.
type Field[T any] struct {
	value   *T
	present bool
}

// NewField creates a Field that is present with the given value.
func NewField[T any](v T) Field[T] {
	return Field[T]{value: &v, present: true}
}

// Present reports whether the field was specified in the configuration.
func (f Field[T]) Present() bool {
	return f.present
}

// Get returns the field value. If not present, returns the zero value of T.
func (f Field[T]) Get() T {
	if f.value != nil {
		return *f.value
	}
	var zero T
	return zero
}

// Set updates the field value and marks it as present.
func (f *Field[T]) Set(v T) {
	f.value = &v
	f.present = true
}

// MarkPresent marks the field as present without changing its value.
// This is used for YAML keys that exist but have null/empty values,
// where the YAML unmarshaler is not called.
func (f *Field[T]) MarkPresent() {
	f.present = true
	if f.value == nil {
		var zero T
		f.value = &zero
	}
}

// UnmarshalYAML implements goccy/go-yaml BytesUnmarshaler.
func (f *Field[T]) UnmarshalYAML(data []byte) error {
	f.present = true
	var v T
	if err := yaml.Unmarshal(data, &v); err != nil {
		return err
	}
	f.value = &v
	return nil
}
