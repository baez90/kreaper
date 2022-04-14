package reaper_test

import (
	"flag"
	"testing"

	"github.com/maxatome/go-testdeep/td"

	"github.com/baez90/kreaper/reaper"
)

func TestParseTarget(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{
			name:    "Empty target",
			in:      "",
			wantErr: true,
		},
		{
			name:    "Missing value",
			in:      "app",
			wantErr: true,
		},
		{
			name:    "Invalid separator",
			in:      "app: prometheus",
			wantErr: true,
		},
		{
			name: "Simple target",
			in:   "app=prometheus",
		},
		{
			name: "full qualified target",
			in:   "app.kubernetes.io/name=prometheus",
		},
		{
			name:    "Invalid domain - too many path segments",
			in:      "app.kubernetes.io/name/detailed=prometheus",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := reaper.ParseTarget(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestTargetFlag(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		want    reaper.Target
		wantErr bool
	}{
		{
			name: "Empty args",
		},
		{
			name:    "Valid target",
			args:    []string{"-target", "app.kubernetes.io/name=prometheus"},
			want:    "app.kubernetes.io/name=prometheus",
			wantErr: false,
		},
		{
			name:    "Invalid target",
			args:    []string{"-target", "app.kubernetes.io/name/detailed=prometheus"},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var target reaper.Target
			flagSet := flag.NewFlagSet(t.Name(), flag.ContinueOnError)
			flagSet.Var(&target, "target", "")
			if err := flagSet.Parse(tt.args); err != nil {
				if !tt.wantErr {
					t.Errorf("Failed to parse arguments: %v", err)
				}
				return
			}

			td.Cmp(t, target, tt.want)
		})
	}
}
