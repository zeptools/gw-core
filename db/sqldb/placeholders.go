package sqldb

import (
	"fmt"
	"strconv"
	"strings"
)

var PlaceholderPrefixForDBType = map[string]byte{
	"mysql":  '?',
	"pgsql":  '$',
	"mssql":  '@',
	"oracle": ':',
	"sqlite": 0, // NOTE: sqlite supports all of them
}

func PlaceholderGF(baseChar byte) func(...int) string { // vararg for optional
	if baseChar == '?' || baseChar == 0 {
		return func(_ ...int) string {
			return "?"
		}
	}
	return func(index ...int) string {
		var i int
		if len(index) == 0 {
			i = 1
		} else {
			i = index[0]
		}
		return fmt.Sprintf("%c%d", baseChar, i)
	}
}

func PlaceholdersGF(baseChar byte) func(int, ...int) []string { // length, start
	if baseChar == '?' || baseChar == 0 {
		return func(length int, _ ...int) []string {
			placeholders := make([]string, length)
			for i := range placeholders {
				placeholders[i] = "?"
			}
			return placeholders
		}
	}
	return func(length int, startIndex ...int) []string {
		placeholders := make([]string, length)
		var startI int
		if len(startIndex) == 0 {
			startI = 1
		} else {
			startI = startIndex[0]
		}
		cnt := startI
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("%c%d", baseChar, cnt)
			cnt++
		}
		return placeholders
	}
}

func ReplaceStaticPlaceholders(sql string, prefix byte) string {
	if prefix == '?' || prefix == 0 {
		return sql
	}
	var builder strings.Builder
	builder.Grow(len(sql) + 8)
	cnt := 1
	i := 0
	for i < len(sql) {
		if sql[i] == '?' {
			// Do Not Touch Dynamic Placeholders '??'
			if i+1 < len(sql) && sql[i+1] == '?' {
				builder.WriteByte('?')
				builder.WriteByte('?')
				i += 2
				continue
			}
			builder.WriteByte(prefix)
			builder.WriteString(strconv.Itoa(cnt))
			cnt++
		} else {
			builder.WriteByte(sql[i])
		}
		i++
	}
	return builder.String()
}

func ExpandDynamicPlaceholders(sql string, prefix byte, counts []int, start int) (string, error) {
	if prefix == '?' || prefix == 0 {
		return expandAnonymousPlaceholders(sql, counts)
	}
	return expandOrdinalPlaceholders(sql, prefix, counts, start)
}

// special case: '?', no numbering
func expandAnonymousPlaceholders(sql string, counts []int) (string, error) {
	const symbol = "??"
	var b strings.Builder
	b.Grow(len(sql) + 16*len(counts))

	i := 0
	countIndex := 0

	for {
		j := strings.Index(sql[i:], symbol)
		if j == -1 {
			b.WriteString(sql[i:])
			break
		}

		b.WriteString(sql[i : i+j])
		i += j + len(symbol) // const len -> compile-time optimized

		if countIndex >= len(counts) {
			return "", fmt.Errorf("expandAnonymousPlaceholders: not enough counts for %q", symbol)
		}

		n := counts[countIndex]
		countIndex++

		for k := 0; k < n; k++ {
			if k > 0 {
				b.WriteString(", ")
			}
			b.WriteByte('?')
		}
	}

	if countIndex < len(counts) {
		return "", fmt.Errorf("expandAnonymousPlaceholders: too many counts for %q", symbol)
	}

	return b.String(), nil
}

func expandOrdinalPlaceholders(sql string, prefix byte, counts []int, start int) (string, error) {
	const symbol = "??"
	var b strings.Builder
	b.Grow(len(sql) + 16*len(counts))

	i := 0
	countIndex := 0
	ord := start

	for {
		j := strings.Index(sql[i:], symbol)
		if j == -1 {
			b.WriteString(sql[i:])
			break
		}

		b.WriteString(sql[i : i+j])
		i += j + len(symbol)

		if countIndex >= len(counts) {
			return "", fmt.Errorf("expandOrdinalPlaceholders: not enough counts for %q", symbol)
		}

		n := counts[countIndex]
		countIndex++

		for k := 0; k < n; k++ {
			if k > 0 {
				b.WriteString(", ")
			}
			b.WriteByte(prefix)
			b.WriteString(strconv.Itoa(ord))
			ord++
		}
	}

	if countIndex < len(counts) {
		return "", fmt.Errorf("expandOrdinalPlaceholders: too many counts for %q", symbol)
	}

	return b.String(), nil
}
