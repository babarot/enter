package modules

import "testing"

func TestFormatPath(t *testing.T) {
	home := "/Users/test"

	tests := []struct {
		name, path, style, want string
	}{
		// parent style
		{"parent basic", "/Users/test/src/project", "parent", "src/project"},
		{"parent home", "/Users/test/project", "parent", "~/project"},
		{"parent root", "/", "parent", "/"},
		{"parent single", "/Users/test/dir", "parent", "~/dir"},

		// full style
		{"full basic", "/Users/test/src/project", "full", "~/src/project"},
		{"full no home", "/opt/data", "full", "/opt/data"},

		// short style
		{"short basic", "/Users/test/src/github/project", "short", "~/s/g/project"},
		{"short shallow", "/Users/test/dir", "short", "~/dir"},

		// basename style
		{"basename basic", "/Users/test/src/project", "basename", "project"},
		{"basename root", "/", "basename", ""},
	}

	for _, tt := range tests {
		got := formatPath(tt.path, home, tt.style)
		if got != tt.want {
			t.Errorf("%s: formatPath(%q, %q, %q) = %q, want %q",
				tt.name, tt.path, home, tt.style, got, tt.want)
		}
	}
}

func TestShortenPath(t *testing.T) {
	tests := []struct {
		path, want string
	}{
		{"~/src/github/project", "~/s/g/project"},
		{"~/dir", "~/dir"},
		{"/a/b/c", "/a/b/c"},
		{"/usr/local/bin", "/u/l/bin"},
	}

	for _, tt := range tests {
		got := shortenPath(tt.path)
		if got != tt.want {
			t.Errorf("shortenPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}
