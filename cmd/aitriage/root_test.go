package main

import "testing"

func TestResolvedBuildVersion(t *testing.T) {
	tests := []struct {
		name   string
		linked string
		module string
		want   string
	}{
		{name: "release ldflags win", linked: "1.4.0", module: "v1.3.0", want: "1.4.0"},
		{name: "go install module version", linked: "dev", module: "v1.3.0", want: "v1.3.0"},
		{name: "local pseudo version", linked: "dev", module: "v1.3.1-0.20260721160635-deadbeef+dirty", want: "dev"},
		{name: "local build", linked: "dev", module: "(devel)", want: "dev"},
		{name: "empty build info", linked: "dev", module: "", want: "dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolvedBuildVersion(tt.linked, tt.module); got != tt.want {
				t.Fatalf("resolvedBuildVersion(%q, %q) = %q, want %q", tt.linked, tt.module, got, tt.want)
			}
		})
	}
}
