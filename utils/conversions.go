package utils

import "strconv"

func Str2Int(str string) int64 {
	i, _ := strconv.Atoi(str)
	return int64(i)
}
