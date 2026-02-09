package middleware

import (
	"strings"

	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/converters/models"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

/* ---------------- UPSTREAM VHOST ---------------- */

// UpstreamVHost handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/upstream-vhost"
func UpstreamVHost(ctx configs.Context) {
	ctx.Log.Debug("running converter UpstreamVHost")

	annUpstreamVhost := string(models.UpstreamVhost)

	val, ok := ctx.Annotations[annUpstreamVhost]
	if !ok || strings.TrimSpace(val) == "" {
		return
	}

	ctx.Result.Middlewares = append(ctx.Result.Middlewares,
		newHeadersMiddleware(ctx, "upstream-vhost", &dynamic.Headers{
			CustomRequestHeaders: map[string]string{
				"Host": val,
			},
		}),
	)

	ctx.ReportConverted(annUpstreamVhost)
}
