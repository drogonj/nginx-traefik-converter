package middleware

import (
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/models"
)

/* ---------------- PROXY REDIRECT ---------------- */

// ProxyRedirect handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/proxy-redirect-from"
//   - "nginx.ingress.kubernetes.io/proxy-redirect-to"
func ProxyRedirect(ctx configs.Context) error {
	ctx.Log.Debug("running converter ProxyRedirect")

	annRedirectFrom := string(models.ProxyRedirectFrom)
	annRedirectTo := string(models.ProxyRedirectTo)

	redirectFrom, hasFrom := ctx.Annotations[annRedirectFrom]
	redirectTo, hasTo := ctx.Annotations[annRedirectTo]

	if !hasFrom && !hasTo {
		return nil
	}

	mw, err := newRewriteResponseHeadersMiddleware(ctx, "Set-Cookie", redirectFrom, redirectTo, "proxy-redirect")
	if err != nil {
		return err
	}

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, mw)

	ctx.ReportConverted(annRedirectFrom)

	ctx.ReportConverted(annRedirectTo)

	return nil
}
