package middleware

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/errors"
)

func parseSizeBytes(val string) (int64, error) {
	value := strings.TrimSpace(strings.ToLower(val))

	const byteValue = 1024

	multiplier := int64(1)

	switch {
	case strings.HasSuffix(value, "k"):
		multiplier = byteValue
		value = strings.TrimSuffix(value, "k")
	case strings.HasSuffix(value, "m"):
		multiplier = byteValue * byteValue
		value = strings.TrimSuffix(value, "m")
	case strings.HasSuffix(value, "g"):
		multiplier = byteValue * byteValue * byteValue
		value = strings.TrimSuffix(value, "g")
	}

	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, &errors.ConverterError{Message: fmt.Sprintf("invalid size value: %s", val)}
	}

	return n * multiplier, nil
}
