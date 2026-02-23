package ingressroute

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/models"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/tls"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// BuildIngressRoute handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/backend-protocol"
//   - "nginx.ingress.kubernetes.io/grpc-backend"
//   - "nginx.ingress.kubernetes.io/use-regex"
func BuildIngressRoute(ctx configs.Context) error {
	ing := ctx.Ingress

	// 1️⃣ Resolve backend protocol ONCE (Ingress-wide)
	scheme, err := resolveScheme(ctx.Annotations)
	if err != nil {
		return err
	}

	useRegex := strings.ToLower(ctx.Annotations[string(models.UseRegex)]) == "true"

	routes := make([]traefik.Route, 0)
	seen := make(map[string]struct{}) // dedup key set

	for _, rule := range ing.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}

		hostMatch := buildHostMatch(rule.Host)

		for _, path := range rule.HTTP.Paths {
			svc := path.Backend.Service
			if svc == nil {
				continue
			}

			pathMatch, ok := buildPathMatch(path, useRegex)
			if useRegex && !ok {
				msg := fmt.Sprintf("use-regex is set but path '%s' is not a valid Go regex for Traefik; fell back to PathPrefix", path.Path)

				ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
				ctx.ReportWarning(string(models.UseRegex), msg)
			}

			match := combineMatch(hostMatch, pathMatch)

			// Build a stable dedup key
			key := fmt.Sprintf(
				"host=%s|path=%s|pathtype=%s|useregex=%t|svc=%s|port=%d|scheme=%s",
				rule.Host,
				path.Path,
				*path.PathType,
				useRegex,
				svc.Name,
				svc.Port.Number,
				scheme,
			)

			if _, exists := seen[key]; exists {
				continue // skip duplicate route
			}

			seen[key] = struct{}{}

			route := traefik.Route{
				Kind:  "Rule",
				Match: match,
				Services: []traefik.Service{
					{
						LoadBalancerSpec: traefik.LoadBalancerSpec{
							Name: svc.Name,
							Port: intstr.IntOrString{
								Type:   intstr.Int,
								IntVal: svc.Port.Number,
							},
							Scheme: scheme,
						},
					},
				},
				Middlewares: middlewareRefs(ctx),
			}

			routes = append(routes, route)
		}
	}

	if len(routes) == 0 {
		return nil
	}

	// Determine entryPoints: frontend TLS (spec.tls) takes priority over backend scheme
	entryPoints := entryPointsForScheme(scheme)
	if len(ing.Spec.TLS) > 0 {
		entryPoints = []string{"websecure"}
	}

	ingressRoute := &traefik.IngressRoute{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "IngressRoute",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ing.Name,
			Namespace: ing.Namespace,
		},
		Spec: traefik.IngressRouteSpec{
			EntryPoints: entryPoints,
			Routes:      routes,
		},
	}

	// Apply TLS from Ingress spec.tls (standard TLS termination)
	applyIngressTLS(ingressRoute, ing)

	// Apply mTLS TLS Option if present (may extend TLS section)
	tls.ApplyTLSOption(ingressRoute, ctx)

	ctx.Result.IngressRoutes = append(ctx.Result.IngressRoutes, ingressRoute)

	if useRegex {
		ctx.ReportConverted(string(models.UseRegex))
	}

	return nil
}

func applyIngressTLS(ingressRoute *traefik.IngressRoute, ing *netv1.Ingress) {
	if len(ing.Spec.TLS) == 0 {
		return
	}

	if ingressRoute.Spec.TLS == nil {
		ingressRoute.Spec.TLS = &traefik.TLS{}
	}

	// Use the first TLS entry's secretName
	for _, t := range ing.Spec.TLS {
		if t.SecretName != "" {
			ingressRoute.Spec.TLS.SecretName = t.SecretName

			break
		}
	}
}

func middlewareRefs(ctx configs.Context) []traefik.MiddlewareRef {
	return orderMiddlewares(ctx.Result.Middlewares)
}

//nolint:varnamelen
func orderMiddlewares(mws []*traefik.Middleware) []traefik.MiddlewareRef {
	var (
		conditional *traefik.Middleware
		cors        *traefik.Middleware
		rest        []*traefik.Middleware
	)

	for _, mw := range mws {
		name := mw.GetName()

		switch {
		case strings.Contains(name, "conditional-return"):
			conditional = mw

		case strings.Contains(name, "cors") || strings.Contains(name, "headers"):
			// your CORS/snippet headers middleware
			cors = mw

		default:
			rest = append(rest, mw)
		}
	}

	refs := make([]traefik.MiddlewareRef, 0, len(mws))

	if conditional != nil {
		refs = append(refs, traefik.MiddlewareRef{Name: conditional.GetName()})
	}

	if cors != nil {
		refs = append(refs, traefik.MiddlewareRef{Name: cors.GetName()})
	}

	for _, mw := range rest {
		refs = append(refs, traefik.MiddlewareRef{Name: mw.GetName()})
	}

	return refs
}

func buildHostMatch(host string) string {
	if host == "" {
		return ""
	}

	return fmt.Sprintf("Host(`%s`)", host)
}

func buildPathMatch(path netv1.HTTPIngressPath, useRegex bool) (string, bool) {
	pth := path.Path
	if pth == "" {
		pth = "/"
	}

	if useRegex {
		regex := pth
		if !strings.HasPrefix(regex, "^") {
			regex = "^" + regex
		}

		if _, err := regexp.Compile(regex); err == nil {
			return fmt.Sprintf("PathRegexp(`%s`)", regex), true
		}

		// invalid regex
		return fmt.Sprintf("PathPrefix(`%s`)", pth), false
	}

	switch *path.PathType {
	case netv1.PathTypeExact:
		return fmt.Sprintf("Path(`%s`)", pth), true
	case netv1.PathTypePrefix:
		return fmt.Sprintf("PathPrefix(`%s`)", pth), true
	case netv1.PathTypeImplementationSpecific:
		return fmt.Sprintf("PathPrefix(`%s`)", pth), true
	default:
		return fmt.Sprintf("PathPrefix(`%s`)", pth), true
	}
}

func combineMatch(hostMatch, pathMatch string) string {
	switch {
	case hostMatch != "" && pathMatch != "":
		return hostMatch + " && " + pathMatch
	case hostMatch != "":
		return hostMatch
	case pathMatch != "":
		return pathMatch
	default:
		return "PathPrefix(`/`)"
	}
}
