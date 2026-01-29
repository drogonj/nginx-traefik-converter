package convert

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/converters/ingressroute"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/converters/middleware"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/converters/tls"
)

func Run(ctx configs.Context, opts configs.Options) error {
	middleware.RewriteTarget(ctx)
	middleware.SSLRedirect(ctx)
	middleware.BasicAuth(ctx)

	if err := middleware.CORS(ctx); err != nil {
		return err
	}
	middleware.RateLimit(ctx)

	if err := middleware.BodySize(ctx); err != nil {
		return err
	}

	middleware.ExtraAnnotations(ctx)
	tls.HandleAuthTLSVerifyClient(ctx)
	middleware.ConfigurationSnippet(ctx)
	middleware.HandleProxyBufferSize(ctx, opts) // 👈 heuristic-aware

	if ingressroute.NeedsIngressRoute(ctx.Annotations) {
		if err := ingressroute.BuildIngressRoute(ctx); err != nil {
			ctx.Result.Warnings = append(ctx.Result.Warnings, err.Error())
		}
	}

	middleware.Warnings(ctx)

	return nil
}
