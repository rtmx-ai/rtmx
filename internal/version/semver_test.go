package version

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input   string
		want    Version
		wantErr bool
	}{
		{"v1.2.3", Version{1, 2, 3, ""}, false},
		{"1.2.3", Version{1, 2, 3, ""}, false},
		{"v0.3.0", Version{0, 3, 0, ""}, false},
		{"v1.0.0-rc1", Version{1, 0, 0, "rc1"}, false},
		{"v2.0.0-beta.1", Version{2, 0, 0, "beta.1"}, false},
		{"v0.0.0", Version{0, 0, 0, ""}, false},
		{"invalid", Version{}, true},
		{"v1.2", Version{}, true},
		{"v1.2.3.4", Version{}, true},
		{"v1.a.3", Version{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Parse(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		v    Version
		want string
	}{
		{Version{1, 2, 3, ""}, "v1.2.3"},
		{Version{0, 3, 0, ""}, "v0.3.0"},
		{Version{1, 0, 0, "rc1"}, "v1.0.0-rc1"},
	}

	for _, tt := range tests {
		if got := tt.v.String(); got != tt.want {
			t.Errorf("%v.String() = %q, want %q", tt.v, got, tt.want)
		}
	}
}

func TestBump(t *testing.T) {
	v := Version{1, 2, 3, ""}
	if got := v.BumpMajor(); got != (Version{2, 0, 0, ""}) {
		t.Errorf("BumpMajor = %v, want v2.0.0", got)
	}
	if got := v.BumpMinor(); got != (Version{1, 3, 0, ""}) {
		t.Errorf("BumpMinor = %v, want v1.3.0", got)
	}
	if got := v.BumpPatch(); got != (Version{1, 2, 4, ""}) {
		t.Errorf("BumpPatch = %v, want v1.2.4", got)
	}
}

func TestActualBump(t *testing.T) {
	tests := []struct {
		from, to Version
		want     BumpLevel
	}{
		{Version{0, 2, 7, ""}, Version{0, 3, 0, ""}, BumpMinor},
		{Version{0, 3, 0, ""}, Version{0, 3, 1, ""}, BumpPatch},
		{Version{0, 3, 0, ""}, Version{1, 0, 0, ""}, BumpMajor},
		{Version{1, 0, 0, ""}, Version{1, 0, 0, ""}, BumpNone},
		{Version{1, 2, 3, ""}, Version{2, 0, 0, ""}, BumpMajor},
	}

	for _, tt := range tests {
		name := tt.from.String() + "->" + tt.to.String()
		t.Run(name, func(t *testing.T) {
			if got := ActualBump(tt.from, tt.to); got != tt.want {
				t.Errorf("ActualBump(%v, %v) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestParseBumpLevel(t *testing.T) {
	tests := []struct {
		input string
		want  BumpLevel
	}{
		{"major", BumpMajor},
		{"Minor", BumpMinor},
		{"PATCH", BumpPatch},
		{"none", BumpNone},
		{"", BumpNone},
		{"invalid", BumpNone},
	}

	for _, tt := range tests {
		if got := ParseBumpLevel(tt.input); got != tt.want {
			t.Errorf("ParseBumpLevel(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
