package marketplace

import (
	"fmt"
	"strings"
)

type Amount float64

func (a Amount) Format() string {
	s := fmt.Sprintf("%.2f", a)
	i := strings.Index(s, ".") - 3
	for i >= 1 {
		s = s[:i] + "," + s[i:]
		i -= 3
	}
	return s
}
