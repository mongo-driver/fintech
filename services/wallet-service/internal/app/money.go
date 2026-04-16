package app

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidAmount = errors.New("invalid amount")

func ParseAmountToCents(amount string) (int64, error) {
	amount = strings.TrimSpace(amount)
	if amount == "" {
		return 0, ErrInvalidAmount
	}
	parts := strings.Split(amount, ".")
	if len(parts) > 2 {
		return 0, ErrInvalidAmount
	}
	intPart, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || intPart < 0 {
		return 0, ErrInvalidAmount
	}
	cents := intPart * 100

	if len(parts) == 2 {
		frac := parts[1]
		if len(frac) == 1 {
			frac += "0"
		}
		if len(frac) != 2 {
			return 0, ErrInvalidAmount
		}
		v, fracErr := strconv.ParseInt(frac, 10, 64)
		if fracErr != nil {
			return 0, ErrInvalidAmount
		}
		cents += v
	}

	if cents <= 0 {
		return 0, ErrInvalidAmount
	}
	return cents, nil
}

func FormatCents(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}
