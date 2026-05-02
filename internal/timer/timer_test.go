package timer

import (
	"testing"
	"time"
)

func TestParseTimer(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"5s", 5 * time.Second, false},
		{"70s", 70 * time.Second, false},
		{"1m5s", 65 * time.Second, false},
		{"1m 5s", 65 * time.Second, false},
		{"1h5m", 65 * time.Minute, false},
		{"1h 5m", 65 * time.Minute, false},
		{"", 0, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		got, err := ParseTimer(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseTimer(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.expected {
			t.Errorf("ParseTimer(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestTimerState(t *testing.T) {
	timer := Timer{
		Name:     "Test",
		Duration: 10 * time.Minute,
		Start:    time.Now().Add(-5 * time.Minute),
		Running:  true,
	}

	remaining := timer.GetRemainingTime()
	// Should be around 5 minutes
	if remaining < 4*time.Minute || remaining > 6*time.Minute {
		t.Errorf("Expected ~5m remaining, got %v", remaining)
	}
}
