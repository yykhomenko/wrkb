package wrkb

import (
	"testing"
)

func TestGetFuncName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"__RANDI64_1_10__", "RANDI64"},
		{"__FOO_123__", "FOO"},
		{"hello world", "NONE"},
	}

	for _, tt := range tests {
		got := getFuncName(tt.input)
		if got != tt.want {
			t.Errorf("getFuncName(%q) = %q; want %q", tt.input, got, tt.want)
		}
	}
}

func TestFuncRandDRegexp(t *testing.T) {
	tests := []struct {
		input    string
		wantLow  string
		wantHigh string
	}{
		{"__RANDI64_5_15__", "5", "15"},
		{"__RANDI64_-10_20__", "-10", "20"},
	}

	for _, tt := range tests {
		matches := funcRandDRegexp.FindStringSubmatch(tt.input)
		if matches == nil {
			t.Errorf("regexp failed to match %q", tt.input)
			continue
		}
		if matches[2] != tt.wantLow || matches[3] != tt.wantHigh {
			t.Errorf("for %q got low=%q high=%q; want low=%q high=%q",
				tt.input, matches[2], matches[3], tt.wantLow, tt.wantHigh)
		}
	}
}
