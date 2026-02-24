package middleware

import (
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
)

func mwName(ctx configs.Context, suffix string) string {
	return ctx.IngressName + "-" + suffix
}
