package convert

import (
	"sort"
	"strings"

	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
)

type middlewareCategory int

const (
	catShortCircuit     middlewareCategory = iota // A: return-status plugin (future)
	catResponseHeaders                            // B: CORS, headers, cookie rewrites, upstream-vhost
	catAuth                                       // C: BasicAuth, ForwardAuth
	catRequestTransform                           // D: rewrite, redirect, bodysize, proxy-redirect
	catOther                                      // E: fallback
)

func classifyMiddleware(middleware *traefik.Middleware) middlewareCategory {
	name := strings.ToLower(middleware.GetName())

	switch {
	// A: short-circuit responders (future plugin)
	case strings.Contains(name, "conditional-return"):
		return catShortCircuit

	// B: response header injectors
	case middleware.Spec.Headers != nil,
		strings.Contains(name, "cors"),
		strings.Contains(name, "configuration-snippet"),
		strings.Contains(name, "proxy-cookie"),
		strings.Contains(name, "upstream-vhost"):
		return catResponseHeaders

	// C: auth
	case middleware.Spec.BasicAuth != nil,
		middleware.Spec.ForwardAuth != nil:
		return catAuth

	// D: request transformers
	case middleware.Spec.ReplacePath != nil,
		middleware.Spec.ReplacePathRegex != nil,
		middleware.Spec.StripPrefix != nil,
		middleware.Spec.RedirectScheme != nil,
		middleware.Spec.RedirectRegex != nil,
		strings.Contains(name, "rewrite"),
		strings.Contains(name, "redirect"),
		strings.Contains(name, "bodysize"),
		strings.Contains(name, "proxy-redirect"):
		return catRequestTransform

	default:
		return catOther
	}
}

func sortMiddlewares(mws []*traefik.Middleware) {
	sort.SliceStable(mws, func(i, j int) bool {
		return classifyMiddleware(mws[i]) < classifyMiddleware(mws[j])
	})
}
