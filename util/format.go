package util

import (
	"fmt"
	"strconv"
	"strings"
)

func FormatInt(n int) string {
	s := strconv.Itoa(n)
	i := len(s) - 3
	for i >= 1 {
		s = s[:i] + "," + s[i:]
		i -= 3
	}
	return s
}

func FormatFloat(f float64) string {
	s := fmt.Sprintf("%.2f", f)
	i := strings.Index(s, ".") - 3
	for i >= 1 {
		s = s[:i] + "," + s[i:]
		i -= 3
	}
	return s
}
