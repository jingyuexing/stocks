package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseTimePoint 将 "HH:MM" 解析为小时和分钟
func ParseTimePoint(val string) (hour, minute int, err error) {
	parts := strings.Split(val, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time point format: %s", val)
	}
	hour, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid hour: %s", parts[0])
	}
	minute, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minute: %s", parts[1])
	}
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("time out of range: %s", val)
	}
	return hour, minute, nil
}

// ParseLocation 解析时区字符串，支持常用缩写和 IANA 名称
func ParseLocation(name string) (*time.Location, error) {
	switch strings.ToUpper(name) {
	case "UTC":
		return time.UTC, nil
	case "EST":
		return time.FixedZone("EST", -5*60*60), nil
	case "EDT":
		return time.FixedZone("EDT", -4*60*60), nil
	case "CST":
		// 默认中国标准时间 UTC+8，如需美国中部可显式用 "America/Chicago"
		return time.FixedZone("CST", 8*60*60), nil
	case "CDT":
		return time.FixedZone("CDT", -5*60*60), nil
	case "PST":
		return time.FixedZone("PST", -8*60*60), nil
	case "PDT":
		return time.FixedZone("PDT", -7*60*60), nil
	default:
		loc, err := time.LoadLocation(name)
		if err != nil {
			return nil, fmt.Errorf("unknown timezone: %s", name)
		}
		return loc, nil
	}
}

// ParseDate 将 "YYYY-MM-DD" 解析为 time.Time
func ParseDate(val string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", val)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: %s", val)
	}
	return t, nil
}

// IsTimeInRanges 检查当前时间是否在任一配对时段内
// session: ["09:30", "11:30", "13:00", "15:00", "CST"]
// 时区为最后一个元素（如果存在），前面的时间点两两配对
func IsTimeInRanges(session []string, now time.Time) (bool, error) {
	if len(session) < 2 {
		return false, fmt.Errorf("session requires at least start and end time")
	}

	// 提取时区（最后一个元素，如果不是时间格式）
	var loc *time.Location
	var timePoints []string
	last := session[len(session)-1]
	if strings.Contains(last, ":") {
		// 最后一个也是时间，没有指定时区，使用本地时区
		loc = time.Local
		timePoints = session
	} else {
		// 最后一个是时区
		var err error
		loc, err = ParseLocation(last)
		if err != nil {
			return false, err
		}
		timePoints = session[:len(session)-1]
	}

	if len(timePoints)%2 != 0 {
		return false, fmt.Errorf("session time points must be paired (start, end)")
	}

	// 将当前时间转换为目标时区
	nowInLoc := now.In(loc)
	currentMinutes := nowInLoc.Hour()*60 + nowInLoc.Minute()

	// 两两配对检查
	for i := 0; i < len(timePoints); i += 2 {
		startH, startM, err := ParseTimePoint(timePoints[i])
		if err != nil {
			return false, err
		}
		endH, endM, err := ParseTimePoint(timePoints[i+1])
		if err != nil {
			return false, err
		}

		startMinutes := startH*60 + startM
		endMinutes := endH*60 + endM

		if startMinutes <= endMinutes {
			// 正常时段，如 09:30…16:00
			if currentMinutes >= startMinutes && currentMinutes <= endMinutes {
				return true, nil
			}
		} else {
			// 跨午夜时段，如 21:00…02:00
			if currentMinutes >= startMinutes || currentMinutes <= endMinutes {
				return true, nil
			}
		}
	}

	return false, nil
}

// IsDateInRange 检查当前日期是否在暂停范围内
// pauseRange: ["2024-02-01", "2024-02-07"] 或 ["2024-02-01"]
func IsDateInRange(pauseRange []string, now time.Time) (bool, error) {
	if len(pauseRange) == 0 {
		return false, nil
	}

	start, err := ParseDate(pauseRange[0])
	if err != nil {
		return false, err
	}

	var end time.Time
	if len(pauseRange) >= 2 {
		end, err = ParseDate(pauseRange[1])
		if err != nil {
			return false, err
		}
	} else {
		end = start
	}

	// 只比较日期部分
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	startDate := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	endDate := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC)

	return !nowDate.Before(startDate) && !nowDate.After(endDate), nil
}
