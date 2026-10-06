package handler

import "strconv"

func formatCents(cents int64) string {
	return strconv.FormatInt(cents, 10)
}

func formatOptionalCents(cents *int64) *string {
	if cents == nil {
		return nil
	}
	formattedCents := formatCents(*cents)
	return &formattedCents
}
