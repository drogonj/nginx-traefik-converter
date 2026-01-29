package middleware

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"strings"
)

/* ---------------- WARNINGS ---------------- */

// Warnings adds warnings to the parsed annotations if any.
func Warnings(ctx configs.Context) {
	for k := range ctx.Annotations {
		if strings.Contains(k, "auth-tls") ||
			strings.Contains(k, "snippet") ||
			strings.Contains(k, "proxy-read") ||
			strings.Contains(k, "proxy-send") {
			ctx.Result.Warnings = append(ctx.Result.Warnings, k+" is not safely convertible")
		}
	}
}

func mwName(ctx configs.Context, suffix string) string {
	return ctx.IngressName + "-" + suffix
}
