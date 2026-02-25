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

	// Resolve backend protocol ONCE (Ingress-wide).
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

			pathMatch, regexPromoted := buildPathMatch(path, useRegex)

			// Warn when use-regex was set but the regex is invalid.
			if useRegex && !regexPromoted && looksLikeRegex(path.Path) {
				msg := fmt.Sprintf("use-regex is set but path '%s' is not a valid Go regex for Traefik; fell back to PathPrefix", path.Path)

				ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
				ctx.ReportWarning(string(models.UseRegex), msg)
			}

			// Warn when path was heuristically promoted to PathRegexp.
			if !useRegex && regexPromoted {
				msg := fmt.Sprintf(
					"path '%s' contains regex patterns without use-regex annotation; "+
						"auto-promoted to PathRegexp — verify behavior",
					path.Path,
				)

				ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
				ctx.ReportWarning("path-regex-heuristic", msg)
			}

			// Warn when path looks like regex but doesn't compile.
			if !useRegex && !regexPromoted && looksLikeRegex(path.Path) {
				msg := fmt.Sprintf(
					"path '%s' contains regex-like characters but is not a valid Go regex; "+
						"fell back to PathPrefix — manual conversion required",
					path.Path,
				)

				ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
				ctx.ReportSkipped("path-regex-heuristic", msg)
			}

			match := combineMatch(hostMatch, pathMatch)

			// Build the service port: prefer number, fall back to name.
			svcPort := buildServicePort(svc.Port)

			// Build a stable dedup key.
			pathTypeStr := "Prefix"
			if path.PathType != nil {
				pathTypeStr = string(*path.PathType)
			}

			key := fmt.Sprintf(
				"host=%s|path=%s|pathtype=%s|useregex=%t|svc=%s|port=%s|scheme=%s",
				rule.Host,
				path.Path,
				pathTypeStr,
				useRegex,
				svc.Name,
				svcPort.String(),
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
							Name:   svc.Name,
							Port:   svcPort,
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

	// EntryPoints are always "web" by default.
	// Frontend TLS (spec.tls) promotes to "websecure".
	entryPoints := []string{"web"}
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

	// Apply TLS from Ingress spec.tls (standard TLS termination).
	applyIngressTLS(ingressRoute, ing)

	// Apply mTLS TLS Option if present (may extend TLS section).
	tls.ApplyTLSOption(ingressRoute, ctx)

	ctx.Result.IngressRoutes = append(ctx.Result.IngressRoutes, ingressRoute)

	if useRegex {
		ctx.ReportConverted(string(models.UseRegex))
	}

	return nil
}

// buildServicePort returns an IntOrString for the Traefik service port.
// It prefers the numeric port; when Number is 0 it falls back to Name.
func buildServicePort(port netv1.ServiceBackendPort) intstr.IntOrString {
	if port.Number != 0 {
		return intstr.FromInt32(port.Number)
	}

	if port.Name != "" {
		return intstr.FromString(port.Name)
	}

	// Both zero and empty — should not happen with valid Ingress.
	return intstr.FromInt32(0)
}

func applyIngressTLS(ingressRoute *traefik.IngressRoute, ing *netv1.Ingress) {
	if len(ing.Spec.TLS) == 0 {
		return
	}

	if ingressRoute.Spec.TLS == nil {
		ingressRoute.Spec.TLS = &traefik.TLS{}
	}

	// Use the first TLS entry's secretName.
	for _, t := range ing.Spec.TLS {
		if t.SecretName != "" {
			ingressRoute.Spec.TLS.SecretName = t.SecretName

			break
		}
	}
}

// middlewareRefs builds MiddlewareRef entries from the already-sorted
// Result.Middlewares slice (sorted by classify_middleware.go).
func middlewareRefs(ctx configs.Context) []traefik.MiddlewareRef {
	refs := make([]traefik.MiddlewareRef, 0, len(ctx.Result.Middlewares))

	for _, mw := range ctx.Result.Middlewares {
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

// regexMetacharPattern detects regex metacharacters that never appear in
// legitimate URL paths. It is intentionally conservative to avoid false
// positives on normal paths that happen to contain characters like '.'.
var regexMetacharPattern = regexp.MustCompile(
	`\.\*` + // .*
		`|\.\+` + // .+
		`|\(\?[:]` + // (?:  non-capturing group
		`|\[[^\]]+\]` + // [abc] character class
		`|\\[dDwWsS.]` + // \d, \w, \. etc.
		`|\{\d+,?\d*\}` + // {2}, {1,3} quantifiers
		`|[^/]\|[^/]`, // alternation (but not double-slash)
)

// looksLikeRegex returns true when a path contains regex metacharacters
// that would never appear in a normal URL path.
func looksLikeRegex(path string) bool {
	return regexMetacharPattern.MatchString(path)
}

// buildPathMatch produces the Traefik match expression for a single path.
// The second return value (regexPromoted) is true when the path was
// heuristically promoted from PathPrefix to PathRegexp.
func buildPathMatch(path netv1.HTTPIngressPath, useRegex bool) (match string, regexPromoted bool) {
	pth := path.Path
	if pth == "" {
		pth = "/"
	}

	// Explicit use-regex annotation: try to compile as Go regex.
	if useRegex {
		return buildRegexpMatch(pth)
	}

	// Heuristic: detect regex metacharacters even when use-regex is absent.
	if looksLikeRegex(pth) {
		if expr, ok := buildRegexpMatch(pth); ok {
			// Auto-promoted — caller will emit a warning.
			return expr, true
		}

		// Contains regex-like chars but does not compile as Go regex.
		// Fall through to PathPrefix; caller will emit a different warning.
		return fmt.Sprintf("PathPrefix(`%s`)", pth), false
	}

	// Normal path — use the declared pathType.
	// Guard against nil PathType (possible on pre-v1.22 clusters).
	pathType := netv1.PathTypePrefix
	if path.PathType != nil {
		pathType = *path.PathType
	}

	switch pathType {
	case netv1.PathTypeExact:
		return fmt.Sprintf("Path(`%s`)", pth), false
	case netv1.PathTypePrefix:
		return fmt.Sprintf("PathPrefix(`%s`)", pth), false
	case netv1.PathTypeImplementationSpecific:
		return fmt.Sprintf("PathPrefix(`%s`)", pth), false
	default:
		return fmt.Sprintf("PathPrefix(`%s`)", pth), false
	}
}

// buildRegexpMatch anchors the path and compiles it as a Go regex.
func buildRegexpMatch(pth string) (string, bool) {
	regex := pth
	if !strings.HasPrefix(regex, "^") {
		regex = "^" + regex
	}

	if _, err := regexp.Compile(regex); err == nil {
		return fmt.Sprintf("PathRegexp(`%s`)", regex), true
	}

	return fmt.Sprintf("PathPrefix(`%s`)", pth), false
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
