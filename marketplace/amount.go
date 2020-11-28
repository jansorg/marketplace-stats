package marketplace

import (
	"github.com/jansorg/marketplace-stats/util"
)

type Amount float64

func (a Amount) Format() string {
	return util.FormatFloat(float64(a))
}
