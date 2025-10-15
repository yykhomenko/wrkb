package wrkb

import (
	"crypto/rand"
	"encoding/hex"
	mathrand "math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Ініціалізуємо генератор псевдорандому
func init() {
	mathrand.Seed(time.Now().UnixNano())
}

// Основна точка входу — підставляє всі патерни в рядок
func substitute(s string) string {
	if len(s) == 0 || !strings.Contains(s, "__RAND") {
		return s // швидкий вихід для статичних URL
	}

	// Використовуємо predeclared slice (без виділень)
	subFns := [...]subFn{subRandI64, subRandHex, subRandStr}

	for i := range subFns {
		s = subFns[i](s)
	}

	return s
}

// Тип функції-підстановки
type subFn func(string) string

// 🧮 RANDI64 — __RANDI64_<low>_<high>__
var reRandI64 = regexp.MustCompile(`__RANDI64_([+-]?\d{1,19})_([+-]?\d{1,19})__`)

func subRandI64(s string) string {
	return reRandI64.ReplaceAllStringFunc(s, func(match string) string {
		m := reRandI64.FindStringSubmatch(match)
		if len(m) != 3 {
			return match
		}

		low, _ := strconv.ParseInt(m[1], 10, 64)
		high, _ := strconv.ParseInt(m[2], 10, 64)
		if high < low {
			low, high = high, low
		}
		val := low + mathrand.Int63n(high-low+1)
		return strconv.FormatInt(val, 10)
	})
}

// 🧩 RANDHEX — __RANDHEX_<len>__
var reRandHex = regexp.MustCompile(`__RANDHEX_(\d{1,3})__`)

func subRandHex(s string) string {
	return reRandHex.ReplaceAllStringFunc(s, func(match string) string {
		m := reRandHex.FindStringSubmatch(match)
		if len(m) != 2 {
			return match
		}
		length, _ := strconv.Atoi(m[1])
		if length <= 0 {
			return match
		}

		// генеруємо байти криптостійким способом
		buf := make([]byte, (length+1)/2)
		if _, err := rand.Read(buf); err != nil {
			return match
		}
		return hex.EncodeToString(buf)[:length]
	})
}

// 🔡 RANDSTR — __RANDSTR_<charset>_<len>__
// charset: letters, digits, lettersdigits
var reRandStr = regexp.MustCompile(`__RANDSTR_(letters|digits|lettersdigits)_(\d{1,3})__`)

func subRandStr(s string) string {
	return reRandStr.ReplaceAllStringFunc(s, func(match string) string {
		m := reRandStr.FindStringSubmatch(match)
		if len(m) != 3 {
			return match
		}

		charset := m[1]
		length, _ := strconv.Atoi(m[2])
		if length <= 0 {
			return match
		}

		var chars string
		switch charset {
		case "letters":
			chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		case "digits":
			chars = "0123456789"
		case "lettersdigits":
			chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		default:
			return match
		}

		b := make([]byte, length)
		for i := range b {
			b[i] = chars[mathrand.Intn(len(chars))]
		}
		return string(b)
	})
}
