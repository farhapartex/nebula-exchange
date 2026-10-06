package handler

import "strconv"

func formatSeed(seed int64) string {
	return strconv.FormatInt(seed, 10)
}
