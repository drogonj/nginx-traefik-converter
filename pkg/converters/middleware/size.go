package middleware

import (
	"fmt"
	"strconv"
	"strings"
)

func parseSizeBytes(val string) (int64, error) {
	v := strings.TrimSpace(strings.ToLower(val))

	multiplier := int64(1)

	switch {
	case strings.HasSuffix(v, "k"):
		multiplier = 1024
		v = strings.TrimSuffix(v, "k")
	case strings.HasSuffix(v, "m"):
		multiplier = 1024 * 1024
		v = strings.TrimSuffix(v, "m")
	case strings.HasSuffix(v, "g"):
		multiplier = 1024 * 1024 * 1024
		v = strings.TrimSuffix(v, "g")
	}

	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size value: %s", val)
	}

	return n * multiplier, nil
}
