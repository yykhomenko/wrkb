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
	case "RAND":
		return substitute(substituteRand(s))
	default:
		return s
	}
}

var nameFuncRegexp = regexp.MustCompile(`%(\w+?)%.*%`)

func getFuncName(s string) string {
	matches := nameFuncRegexp.FindStringSubmatch(s)
	if matches != nil {
		return matches[1]
	} else {
		return "NONE"
	}
}

var funcRandDRegexp = regexp.MustCompile(`%(\w+?)%([+-]?\d{1,19}),([+-]?\d{1,19})%`)

func substituteRand(s string) string {
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
