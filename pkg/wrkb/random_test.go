package wrkb

import (
	"regexp"
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

func TestSubstituteRandHex_LengthAndFormat(t *testing.T) {
	input := "__RANDHEX_32__"
	out := substituteRandHex(input)

	if len(out) != 32 {
		t.Errorf("got length %d, want 32", len(out))
	}

	match, _ := regexp.MatchString("^[0-9a-f]+$", out)
	if !match {
		t.Errorf("output %q contains non-hex characters", out)
	}
}

func TestSubstituteRandStr_LettersDigits(t *testing.T) {
	input := "__RANDSTR_lettersdigits_16__"
	out := substituteRandStr(input)

	if len(out) != 16 {
		t.Errorf("got length %d, want 16", len(out))
	}

	match, _ := regexp.MatchString("^[a-zA-Z0-9]+$", out)
	if !match {
		t.Errorf("output %q contains invalid characters", out)
	}
}
