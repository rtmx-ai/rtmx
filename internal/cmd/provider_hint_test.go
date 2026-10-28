package cmd

import "testing"

func TestInferProviderHint(t *testing.T) {
	tests := []struct {
		name    string
		remotes []GitRemote
		want    string
	}{
		{
			name:    "empty remotes",
			remotes: nil,
			want:    "",
		},
		{
			name:    "empty url",
			remotes: []GitRemote{{Name: "origin", URL: ""}},
			want:    "",
		},
		{
			name:    "https github.com",
			remotes: []GitRemote{{Name: "origin", URL: "https://github.com/rtmx-ai/rtmx.git"}},
			want:    "github",
		},
		{
			name:    "ssh git@github.com",
			remotes: []GitRemote{{Name: "origin", URL: "git@github.com:rtmx-ai/rtmx.git"}},
			want:    "github",
		},
		{
			name:    "ssh://git@github.com",
			remotes: []GitRemote{{Name: "origin", URL: "ssh://git@github.com/rtmx-ai/rtmx"}},
			want:    "github",
		},
		{
			name:    "github enterprise host",
			remotes: []GitRemote{{Name: "origin", URL: "https://github.example.com/acme/app.git"}},
			want:    "github",
		},
		{
			name:    "www.github.com",
			remotes: []GitRemote{{Name: "origin", URL: "https://www.github.com/acme/app"}},
			want:    "github",
		},
		{
			name:    "https gitlab.com",
			remotes: []GitRemote{{Name: "origin", URL: "https://gitlab.com/acme/x.git"}},
			want:    "gitlab",
		},
		{
			name:    "ssh git@gitlab.com",
			remotes: []GitRemote{{Name: "origin", URL: "git@gitlab.com:acme/x.git"}},
			want:    "gitlab",
		},
		{
			name:    "self-hosted gitlab host",
			remotes: []GitRemote{{Name: "origin", URL: "https://gitlab.company.io/team/repo.git"}},
			want:    "gitlab",
		},
		{
			name:    "unknown host",
			remotes: []GitRemote{{Name: "origin", URL: "https://bitbucket.org/acme/app.git"}},
			want:    "",
		},
		{
			name: "prefer origin over other remotes",
			remotes: []GitRemote{
				{Name: "upstream", URL: "https://gitlab.com/acme/x.git"},
				{Name: "origin", URL: "https://github.com/acme/app.git"},
			},
			want: "github",
		},
		{
			name: "no origin uses first match",
			remotes: []GitRemote{
				{Name: "fork", URL: "https://bitbucket.org/acme/app.git"},
				{Name: "upstream", URL: "https://gitlab.com/acme/x.git"},
			},
			want: "gitlab",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferProviderHint(tt.remotes)
			if got != tt.want {
				t.Fatalf("InferProviderHint() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProviderHintFromRemoteURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"", ""},
		{"https://github.com/o/r", "github"},
		{"git@github.com:o/r.git", "github"},
		{"https://gitlab.com/o/r", "gitlab"},
		{"https://code.example.com/o/r", ""},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := ProviderHintFromRemoteURL(tt.url)
			if got != tt.want {
				t.Fatalf("ProviderHintFromRemoteURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}
