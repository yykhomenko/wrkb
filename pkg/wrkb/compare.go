package wrkb

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type compareRow struct {
	Field   string
	Base    string
	Next    string
	AbsDiff string
	PctDiff string
	Cmp     int
}

func readBestResultJSON(path string) (bestResultJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return bestResultJSON{}, err
	}

	var payload bestResultJSON
	if err := json.Unmarshal(data, &payload); err != nil {
		return bestResultJSON{}, err
	}
	return payload, nil
}

func writeBestResultCompareCSV(path string, rows []compareRow) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"field", "base", "next", "abs_diff", "pct_diff"}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{row.Field, row.Base, row.Next, row.AbsDiff, row.PctDiff}); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

func buildCompareRows(base bestResultJSON, next bestResultJSON) []compareRow {
	baseVal := reflect.ValueOf(base)
	nextVal := reflect.ValueOf(next)
	structType := baseVal.Type()

	rows := make([]compareRow, 0, structType.NumField())
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		label := field.Tag.Get("csv")
		if label == "" {
			continue
		}

		baseField := baseVal.Field(i)
		nextField := nextVal.Field(i)

		baseStr := formatValue(field, baseField)
		nextStr := formatValue(field, nextField)

		absDiff := ""
		pctDiff := ""
		cmp := 0
		if baseNum, ok := numericValue(field, baseField); ok {
			if nextNum, ok := numericValue(field, nextField); ok {
				diff := nextNum - baseNum
				absDiff = formatAbsDiff(field, math.Abs(diff))
				if baseNum != 0 {
					pctDiff = fmt.Sprintf("%+.2f%%", (diff/baseNum)*100)
				} else {
					pctDiff = fmt.Sprintf("%+.2f%%", 0.00)
				}
				if diff != 0 {
					if dir, ok := compareDirection(field); ok {
						if (dir == 1 && diff > 0) || (dir == -1 && diff < 0) {
							cmp = 1
						} else {
							cmp = -1
						}
					}
				}
			}
		}

		rows = append(rows, compareRow{
			Field:   label,
			Base:    baseStr,
			Next:    nextStr,
			AbsDiff: absDiff,
			PctDiff: pctDiff,
			Cmp:     cmp,
		})
	}

	return rows
}

func printCompareTable(rows []compareRow) {
	if len(rows) == 0 {
		return
	}

	headers := []string{"field", "base", "next", "abs_diff", "pct_diff"}
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = visibleWidth(h)
	}

	for _, row := range rows {
		values := []string{row.Field, row.Base, row.Next, row.AbsDiff, row.PctDiff}
		for i, v := range values {
			if w := visibleWidth(v); w > widths[i] {
				widths[i] = w
			}
		}
	}

	line := func(left, mid, right string) string {
		var b strings.Builder
		b.WriteString(left)
		for i, w := range widths {
			b.WriteString(strings.Repeat("─", w+2))
			if i == len(widths)-1 {
				b.WriteString(right)
			} else {
				b.WriteString(mid)
			}
		}
		return b.String()
	}

	fmt.Printf("%s\n", line("┌", "┬", "┐"))
	fmt.Printf("│ %s │ %s │ %s │ %s │ %s │\n",
		padRightANSI(headers[0], widths[0]),
		padRightANSI(headers[1], widths[1]),
		padRightANSI(headers[2], widths[2]),
		padRightANSI(headers[3], widths[3]),
		padRightANSI(headers[4], widths[4]),
	)
	fmt.Printf("%s\n", line("├", "┼", "┤"))

	for _, row := range rows {
		next := colorizeCompare(row.Next, row.Cmp)
		absDiff := colorizeCompare(row.AbsDiff, row.Cmp)
		pctDiff := colorizeCompare(row.PctDiff, row.Cmp)
		fmt.Printf("│ %s │ %s │ %s │ %s │ %s │\n",
			padRightANSI(row.Field, widths[0]),
			padRightANSI(row.Base, widths[1]),
			padRightANSI(next, widths[2]),
			padRightANSI(absDiff, widths[3]),
			padRightANSI(pctDiff, widths[4]),
		)
	}

	fmt.Printf("%s\n\n", line("└", "┴", "┘"))
}

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func padRightANSI(value string, width int) string {
	if width <= 0 {
		return value
	}
	visible := visibleWidth(value)
	if visible >= width {
		return value
	}
	return value + strings.Repeat(" ", width-visible)
}

func visibleWidth(value string) int {
	plain := ansiRegexp.ReplaceAllString(value, "")
	return utf8.RuneCountInString(plain)
}

func compareDirection(field reflect.StructField) (int, bool) {
	switch field.Tag.Get("cmpBetter") {
	case "higher":
		return 1, true
	case "lower":
		return -1, true
	default:
		return 0, false
	}
}

func colorizeCompare(value string, cmp int) string {
	if value == "" || cmp == 0 {
		return value
	}
	if cmp > 0 {
		return green + value + reset
	}
	return red + value + reset
}

func valueToString(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

func numericValue(field reflect.StructField, v reflect.Value) (float64, bool) {
	if field.Tag.Get("cmpKind") == "duration" {
		if us, ok := durationMicros(v); ok {
			return float64(us), true
		}
		return 0, false
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(v.Uint()), true
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	default:
		return 0, false
	}
}

func formatAbsDiff(field reflect.StructField, diff float64) string {
	if field.Tag.Get("cmpKind") == "duration" {
		return (time.Duration(int64(diff)) * time.Microsecond).String()
	}

	switch field.Type.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(int64(diff), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(uint64(diff), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(diff, 'f', -1, 64)
	default:
		return strconv.FormatFloat(diff, 'f', -1, 64)
	}
}

func formatValue(field reflect.StructField, v reflect.Value) string {
	if field.Tag.Get("cmpKind") == "duration" {
		if us, ok := durationMicros(v); ok {
			return (time.Duration(us) * time.Microsecond).String()
		}
	}
	return valueToString(v)
}

func durationMicros(v reflect.Value) (int64, bool) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int64(v.Uint()), true
	case reflect.Float32, reflect.Float64:
		return int64(v.Float()), true
	case reflect.String:
		if v.String() == "" {
			return 0, false
		}
		d, err := time.ParseDuration(v.String())
		if err != nil {
			return 0, false
		}
		return d.Microseconds(), true
	default:
		return 0, false
	}
}
