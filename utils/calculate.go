package utils

import "errors"

// 计算盈利百分比
func CalculateProfitPercentage(beginValue, nowValue float64) (float64, error) {
	if beginValue == 0 {
		return 0, errors.New("begin value cannot be zero")
	}
	profitPercentage := ((nowValue - beginValue) / beginValue) * 100
	return profitPercentage, nil
}