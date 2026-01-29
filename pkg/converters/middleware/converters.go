package middleware

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"strings"
)

func mwName(ctx configs.Context, suffix string) string {
	return ctx.IngressName + "-" + suffix
}

/* ---------------- WARNINGS ---------------- */

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
