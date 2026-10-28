// SPDX-License-Identifier: GPL-2.0-or-later
package utils

import (
	"strconv"
	"time"
)

func Str2Int(str string) int64 {
	i, _ := strconv.Atoi(str)
	return int64(i)
}

func Str2Float(str string) float64 {
	f, _ := strconv.ParseFloat(str, 64)
	return f
}

func Str2Time(str string) time.Time {
	if t, err := time.Parse(time.DateOnly, str); err != nil {
		return time.Time{}
	} else {
		return t
	}
}
