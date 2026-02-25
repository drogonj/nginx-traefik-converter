package convert

import (
	"fmt"
	"strings"

	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/certificate"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/ingressroute"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/middleware"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/models"
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

	if err := middleware.LimitConnections(ctx); err != nil {
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
	middleware.ProxyTimeouts(ctx)

	sortMiddlewares(ctx.Result.Middlewares)

	if err := ingressroute.BuildIngressRoute(ctx); err != nil {
		ctx.Result.Warnings = append(ctx.Result.Warnings, err.Error())
	}

	tls.HandleAuthTLSVerifyClient(ctx)

	// Extract or generate cert-manager Certificate resources.
	certificate.ExtractOrGenerate(ctx)

	// Warn about any nginx.ingress.kubernetes.io/* annotations that are
	// present on the Ingress but not recognised by the converter.
	warnUnknownAnnotations(ctx)

	return nil
}

// knownAnnotations is the precomputed set of annotation keys the converter
// recognises. Built once from models.AllAnnotations.
var knownAnnotations = func() map[string]struct{} {
	m := make(map[string]struct{}, len(models.AllAnnotations))
	for _, a := range models.AllAnnotations {
		m[string(a)] = struct{}{}
	}

	return m
}()

// ignoredPrefixes lists annotation prefixes that are not nginx-specific and
// should not trigger an "unknown annotation" warning.
var ignoredPrefixes = []string{
	"kubernetes.io/",
	"kubectl.kubernetes.io/",
	"meta.helm.sh/",
	"helm.sh/",
	"app.kubernetes.io/",
	"argocd.argoproj.io/",
	"fluxcd.io/",
}

// warnUnknownAnnotations reports any nginx.ingress.kubernetes.io/* annotation
// that is not in the converter's known list.
func warnUnknownAnnotations(ctx configs.Context) {
	const nginxPrefix = "nginx.ingress.kubernetes.io/"

	for key := range ctx.Annotations {
		// Only flag nginx ingress annotations.
		if !strings.HasPrefix(key, nginxPrefix) {
			// Also skip well-known non-nginx prefixes silently.
			if hasAnyPrefix(key, ignoredPrefixes) {
				continue
			}

			// For any other unrecognised prefix (not nginx, not well-known),
			// skip silently — the converter only cares about nginx annotations.
			continue
		}

		if _, known := knownAnnotations[key]; !known {
			msg := fmt.Sprintf("annotation %q is not supported by the converter and was ignored", key)
			ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
			ctx.ReportSkipped(key, msg)
		}
	}
}

func hasAnyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}

	return false
}
