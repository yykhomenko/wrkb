package wrkb

import (
	"encoding/hex"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

func substitute(s string) string {
	switch getFuncName(s) {
	case "RANDI64":
		return substitute(substituteRandI64(s))
	case "RANDHEX":
		return substitute(substituteRandHex(s))
	case "RANDSTR":
		return substitute(substituteRandStr(s))
	default:
		return s
	}
}

var nameFuncRegexp = regexp.MustCompile(`__(\w+?)_.*__`)

func getFuncName(s string) string {
	matches := nameFuncRegexp.FindStringSubmatch(s)
	if matches != nil {
		return matches[1]
	} else {
		return "NONE"
	}
}

var funcRandDRegexp = regexp.MustCompile(`__(\w+?)_([+-]?\d{1,19})_([+-]?\d{1,19})__`)

func substituteRandI64(s string) string {
	matches := funcRandDRegexp.FindStringSubmatch(s)
	if matches != nil {
		toReplace := matches[0]
		lowStr := matches[2]
		highStr := matches[3]

		low, err := strconv.ParseInt(lowStr, 10, 64)
		if err != nil {
			log.Printf("error: RANDI64: unable to parse 'low' parameter: %s\n", err.Error())
			return s
		}

		high, err := strconv.ParseInt(highStr, 10, 64)
		if err != nil {
			log.Printf("error: RANDI64: unable to parse 'high' parameter: %s\n", err.Error())
			return s
		}

		value := strconv.FormatInt(low+rand.Int63n(high-low+1), 10)
		return strings.Replace(s, toReplace, value, 1)
	} else {
		return s
	}
}

var funcRandHexRegexp = regexp.MustCompile(`__(\w+?)_(\d{1,3})__`)

func substituteRandHex(s string) string {
	matches := funcRandHexRegexp.FindStringSubmatch(s)
	if matches != nil {
		toReplace := matches[0]
		lengthStr := matches[2]

		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			log.Printf("error: RANDHEX: unable to parse 'length' parameter: %s\n", err.Error())
			return s
		}

		buf := make([]byte, (length+1)/2)
		_, err = rand.Read(buf)
		if err != nil {
			log.Printf("error: RANDHEX: %s\n", err.Error())
			return s
		}

		hexStr := hex.EncodeToString(buf)[:length]
		return strings.Replace(s, toReplace, hexStr, 1)
	}
	return s
}

var funcRandStrRegexp = regexp.MustCompile(`__RANDSTR_(\w+)_(\d{1,3})__`)

func substituteRandStr(s string) string {
	matches := funcRandStrRegexp.FindStringSubmatch(s)
	if matches != nil {
		toReplace := matches[0]
		charset := matches[1]
		lengthStr := matches[2]

		length, err := strconv.Atoi(lengthStr)
		if err != nil || length <= 0 {
			return s
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
			return s
		}

		b := make([]byte, length)
		for i := range b {
			b[i] = chars[rand.Intn(len(chars))]
		}
		return strings.Replace(s, toReplace, string(b), 1)
	}
	return s
}
