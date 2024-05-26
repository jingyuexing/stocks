package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func TimeDuration(duration string) (time.Time, error) {
	const (
		SECOND        uint64 = 1
		MINUTE_SECOND uint64 = 60 * SECOND
		HOUR_SECOND   uint64 = 60 * MINUTE_SECOND
		DAY_SECOND    uint64 = 24 * HOUR_SECOND
		WEEK_SECOND   uint64 = 7 * DAY_SECOND
		MONTH_SECOND  uint64 = 30 * DAY_SECOND
		YEAR_SECOND   uint64 = 12 * MONTH_SECOND
	)
	uints := map[string]uint64{
		"y": YEAR_SECOND,
		"m": MONTH_SECOND,
		"w": WEEK_SECOND,
		"d": DAY_SECOND,
		"h": HOUR_SECOND,
		"s": SECOND,
	}
	splitDuration := strings.Split(duration, " ")
	durationSecond := uint64(0)
	for _, val := range splitDuration {
		unit := val[len(val)-1:]
		num, _ := strconv.Atoi(val[:len(val)-1])
		if unitSeconds, ok := uints[unit]; ok {
			durationSecond += uint64(num) * unitSeconds
		} else {
			return time.Time{}, fmt.Errorf("haven't this unit")
		}
	}
	return time.Now().Add(time.Duration(durationSecond) * time.Second).UTC(), nil
}
