package wrkb

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type compareRow struct {
	Field   string
	Base    string
	Next    string
	AbsDiff string
	PctDiff string
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
		label := field.Tag.Get("cmp")
		if label == "" {
			continue
		}

		baseField := baseVal.Field(i)
		nextField := nextVal.Field(i)

		baseStr := valueToString(baseField)
		nextStr := valueToString(nextField)

		absDiff := ""
		pctDiff := ""
		if baseNum, ok := numericValue(field, baseField); ok {
			if nextNum, ok := numericValue(field, nextField); ok {
				diff := nextNum - baseNum
				absDiff = formatAbsDiff(field, math.Abs(diff))
				if baseNum != 0 {
					pctDiff = fmt.Sprintf("%+.2f%%", (diff/baseNum)*100)
				}
			}
		}

		rows = append(rows, compareRow{
			Field:   label,
			Base:    baseStr,
			Next:    nextStr,
			AbsDiff: absDiff,
			PctDiff: pctDiff,
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
		widths[i] = len(h)
	}

	for _, row := range rows {
		values := []string{row.Field, row.Base, row.Next, row.AbsDiff, row.PctDiff}
		for i, v := range values {
			if len(v) > widths[i] {
				widths[i] = len(v)
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
	fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │\n",
		widths[0], headers[0],
		widths[1], headers[1],
		widths[2], headers[2],
		widths[3], headers[3],
		widths[4], headers[4],
	)
	fmt.Printf("%s\n", line("├", "┼", "┤"))

	for _, row := range rows {
		fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │\n",
			widths[0], row.Field,
			widths[1], row.Base,
			widths[2], row.Next,
			widths[3], row.AbsDiff,
			widths[4], row.PctDiff,
		)
	}

	fmt.Printf("%s\n\n", line("└", "┴", "┘"))
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
		if v.Kind() != reflect.String {
			return 0, false
		}
		if v.String() == "" {
			return 0, false
		}
		d, err := time.ParseDuration(v.String())
		if err != nil {
			return 0, false
		}
		return float64(d.Nanoseconds()), true
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
		return time.Duration(int64(diff)).String()
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
