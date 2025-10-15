package wrkb

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// ✅ RANDI64: базовий тест
func TestSubRandI64_Range(t *testing.T) {
	input := "__RANDI64_10_20__"

	for i := 0; i < 100; i++ {
		out := subRandI64(input)
		n, err := strconv.Atoi(out)
		if err != nil {
			t.Fatalf("expected numeric output, got %q", out)
		}
		if n < 10 || n > 20 {
			t.Errorf("value %d out of range 10–20", n)
		}
	}
}

// ✅ RANDHEX: різна довжина і hex формат
func TestSubRandHex_LengthAndFormat(t *testing.T) {
	for _, n := range []int{1, 8, 15, 32, 63} {
		in := fmt.Sprintf("__RANDHEX_%d__", n)
		out := subRandHex(in)

		if len(out) != n {
			t.Errorf("expected length %d, got %d (%q)", n, len(out), out)
		}
		if ok, _ := regexp.MatchString("^[0-9a-f]+$", out); !ok {
			t.Errorf("output %q not hex", out)
		}
	}
}

// ✅ RANDSTR: перевірка різних charset
func TestSubRandStr_Charsets(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		regex    string
		expected int
	}{
		{"letters", "__RANDSTR_letters_10__", "^[a-zA-Z]+$", 10},
		{"digits", "__RANDSTR_digits_6__", "^[0-9]+$", 6},
		{"lettersdigits", "__RANDSTR_lettersdigits_12__", "^[a-zA-Z0-9]+$", 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := subRandStr(tt.input)
			if len(out) != tt.expected {
				t.Errorf("expected length %d, got %d (%q)", tt.expected, len(out), out)
			}
			if ok, _ := regexp.MatchString(tt.regex, out); !ok {
				t.Errorf("output %q does not match %s", out, tt.regex)
			}
		})
	}
}

// ✅ substitute(): інтеграційний тест для кількох функцій
func TestSubstitute_MultiplePatterns(t *testing.T) {
	input := "http://localhost:8080/messages?from=__RANDI64_700_777__&text=__RANDSTR_lettersdigits_8__&token=__RANDHEX_12__"
	out := substitute(input)

	// перевіряємо, що всі шаблони зникли
	for _, pattern := range []string{"__RANDI64", "__RANDSTR", "__RANDHEX"} {
		if strings.Contains(out, pattern) {
			t.Errorf("%s not substituted in %q", pattern, out)
		}
	}

	// базова валідація параметрів
	if !strings.Contains(out, "from=") || !strings.Contains(out, "text=") || !strings.Contains(out, "token=") {
		t.Fatalf("some parameters missing in output: %q", out)
	}
}

// ✅ substitute(): якщо немає патернів — рядок не змінюється
func TestSubstitute_NoPattern(t *testing.T) {
	input := "https://example.com/static/path"
	out := substitute(input)
	if out != input {
		t.Errorf("expected unchanged output, got %q", out)
	}
}

//
// 🧪 Benchmarks
//

func BenchmarkSubRandI64(b *testing.B) {
	input := "__RANDI64_1000_9999__"
	for i := 0; i < b.N; i++ {
		_ = subRandI64(input)
	}
}

func BenchmarkSubRandHex(b *testing.B) {
	input := "__RANDHEX_64__"
	for i := 0; i < b.N; i++ {
		_ = subRandHex(input)
	}
}

func BenchmarkSubRandStr(b *testing.B) {
	input := "__RANDSTR_lettersdigits_32__"
	for i := 0; i < b.N; i++ {
		_ = subRandStr(input)
	}
}

func BenchmarkSubstitute_Combined(b *testing.B) {
	input := "http://localhost:8080/messages?from=__RANDI64_700_777__&to=__RANDI64_380670000001_380670099999__&text=__RANDSTR_lettersdigits_16__&token=__RANDHEX_8__"
	for i := 0; i < b.N; i++ {
		_ = substitute(input)
	}
}
