package convert

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/converters/ingressroute"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/converters/middleware"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/converters/tls"
)

// Run processes ingress annotations using the available converters.
// It is the core function responsible for converting NGINX Ingress
// annotations into their Traefik equivalents.
func Run(ctx configs.Context, opts configs.Options) error {
	middleware.RewriteTargets(ctx)
	middleware.SSLRedirect(ctx)
	middleware.BasicAuth(ctx)

	if err := middleware.RateLimit(ctx); err != nil {
		return err
	}

	if err := middleware.CORS(ctx); err != nil {
		return err
	}

	if err := middleware.BodySize(ctx); err != nil {
		return err
	}

	if err := middleware.ProxyRedirect(ctx); err != nil {
		return err
	}

	middleware.ConfigurationSnippets(ctx)
	middleware.ProxyBufferSizes(ctx, opts) // 👈 heuristic-aware
	middleware.UpstreamVHost(ctx)
	middleware.ServerSnippet(ctx)

	if ingressroute.NeedsIngressRoute(ctx.Annotations) {
		if err := ingressroute.BuildIngressRoute(ctx); err != nil {
			ctx.Result.Warnings = append(ctx.Result.Warnings, err.Error())
		}
	}

	middleware.ExtraAnnotations(ctx)
	tls.HandleAuthTLSVerifyClient(ctx)

	// middleware.Warnings(ctx)

	return nil
}
