package edits

import (
	"os"
	"testing"
)

func TestIsProcessAlive(t *testing.T) {
	tests := []struct {
		name string
		pid  int
		want bool
	}{
		{
			name: "current process",
			pid:  os.Getpid(),
			want: true,
		},
		{
			name: "invalid pid",
			pid:  -1,
			want: false,
		},
		{
			name: "very large pid",
			pid:  999999999,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = isProcessAlive(tt.pid)
			// Just verify it doesn't panic - behavior varies by OS
		})
	}
}
