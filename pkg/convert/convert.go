package convert

import (
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/ingressroute"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/middleware"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/tls"
)

// Run processes ingress annotations using the available converters.
// It is the core function responsible for converting NGINX Ingress
// annotations into their Traefik equivalents.
func Run(ctx configs.Context) error {
	if err := middleware.CORS(ctx); err != nil {
		return err
	}

	if err := middleware.ProxyCookiePath(ctx); err != nil {
		return err
	}

	middleware.UpstreamVHost(ctx)
	middleware.BasicAuth(ctx)

	if err := middleware.BodySize(ctx); err != nil {
		return err
	}

	middleware.RewriteTargets(ctx)
	middleware.SSLRedirect(ctx)

	if err := middleware.RateLimit(ctx); err != nil {
		return err
	}

	if err := middleware.ProxyRedirect(ctx); err != nil {
		return err
	}

	if err := middleware.ConfigurationSnippets(ctx); err != nil {
		return err
	}

	middleware.ProxyBufferSizes(ctx) // 👈 heuristic-aware
	middleware.ServerSnippet(ctx)
	middleware.EnableUnderscoresInHeaders(ctx)
	middleware.ExtraAnnotations(ctx)
	middleware.ProxyBuffering(ctx)
	middleware.HandleAuthURL(ctx)

	sortMiddlewares(ctx.Result.Middlewares)

		if err := ingressroute.BuildIngressRoute(ctx); err != nil {
			ctx.Result.Warnings = append(ctx.Result.Warnings, err.Error())
	}

	tls.HandleAuthTLSVerifyClient(ctx)

	return nil
}
