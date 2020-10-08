package marketplace

import (
	"time"
)

var ServerTimeZone, _ = time.LoadLocation("Europe/Berlin")
var feeChangeDate = time.Date(2020, time.July, 1, 0, 0, 0, 0, ServerTimeZone)
