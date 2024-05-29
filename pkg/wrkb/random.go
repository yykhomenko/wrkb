package wrkb

import (
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
			log.Printf("error: unable to parse 'low' parameter: %s\n", err.Error())
			return s
		}

		high, err := strconv.ParseInt(highStr, 10, 64)
		if err != nil {
			log.Printf("error: unable to parse 'high' parameter: %s\n", err.Error())
			return s
		}

		value := strconv.FormatInt(low+rand.Int63n(high-low+1), 10)
		return strings.Replace(s, toReplace, value, 1)
	} else {
		return s
	}
}
