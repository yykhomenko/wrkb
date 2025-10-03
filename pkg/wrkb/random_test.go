package wrkb

import (
	"regexp"
	"strconv"
	"strings"
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

func TestSubstitute_MultipleFunctions(t *testing.T) {
	input := "http://localhost:8080/messages?from=__RANDI64_700_777__&text=__RANDSTR_lettersdigits_8__"
	out := substitute(input)

	if strings.Contains(out, "__RANDI64") {
		t.Errorf("RANDI64 not substituted in %q", out)
	}
	if strings.Contains(out, "__RANDSTR") {
		t.Errorf("RANDSTR not substituted in %q", out)
	}

	parts := strings.Split(out, "?")
	if len(parts) != 2 {
		t.Fatalf("invalid URL format: %q", out)
	}

	params := parts[1]
	vals := strings.Split(params, "&")
	if len(vals) != 2 {
		t.Fatalf("unexpected query params: %q", params)
	}

	from := strings.Split(vals[0], "=")[1]
	n, err := strconv.Atoi(from)
	if err != nil {
		t.Errorf("expected number for 'from', got %q", from)
	}
	if n < 700 || n > 777 {
		t.Errorf("from=%d out of range 700–777", n)
	}

	text := strings.Split(vals[1], "=")[1]
	if len(text) != 8 {
		t.Errorf("text=%q length=%d; want 8", text, len(text))
	}
	match, _ := regexp.MatchString("^[a-zA-Z0-9]+$", text)
	if !match {
		t.Errorf("text=%q contains invalid characters", text)
	}
}
