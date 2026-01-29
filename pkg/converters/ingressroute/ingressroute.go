package ingressroute

import (
	"fmt"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/converters/tls"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

// BuildIngressRoute builds ingress routes to nginx ingresses if required.
func BuildIngressRoute(ctx configs.Context) error {
	ing := ctx.Ingress

	// 1️⃣ Resolve backend protocol ONCE (Ingress-wide)
	scheme, err := resolveScheme(ctx.Annotations)
	if err != nil {
		return err
	}

	// 2️⃣ Deduplicate services
	type svcKey struct {
		name string
		port int32
	}

	services := map[svcKey]traefik.Service{}

	for _, rule := range ing.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}

		for _, path := range rule.HTTP.Paths {
			svc := path.Backend.Service
			if svc == nil {
				continue
			}

			key := svcKey{
				name: svc.Name,
				port: svc.Port.Number,
			}

			if _, exists := services[key]; exists {
				continue
			}

			services[key] = traefik.Service{
				LoadBalancerSpec: traefik.LoadBalancerSpec{
					Name: svc.Name,
					Port: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: svc.Port.Number,
					},
					Scheme: scheme,
				},
			}
		}
	}

	if len(services) == 0 {
		return nil
	}

	// 3️⃣ Build ONE Route per service
	routes := make([]traefik.Route, 0, len(services))

	for _, svc := range services {
		routes = append(routes, traefik.Route{
			Kind:  "Rule",
			Match: buildHostOnlyMatch(ing),
			Services: []traefik.Service{
				svc,
			},
			Middlewares: middlewareRefs(ctx),
		})
	}

	// 4️⃣ Build ONE IngressRoute
	ir := &traefik.IngressRoute{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "IngressRoute",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ing.Name,
			Namespace: ing.Namespace,
		},
		Spec: traefik.IngressRouteSpec{
			EntryPoints: entryPointsForScheme(scheme),
			Routes:      routes,
		},
	}

	tls.ApplyTLSOption(ir, ctx)

	ctx.Result.IngressRoutes = append(ctx.Result.IngressRoutes, ir)
	return nil
}

func middlewareRefs(ctx configs.Context) []traefik.MiddlewareRef {
	var refs []traefik.MiddlewareRef
	for _, mw := range ctx.Result.Middlewares {
		refs = append(refs, traefik.MiddlewareRef{
			Name: mw.GetName(),
		})
	}
	return refs
}

func buildHostOnlyMatch(ing *netv1.Ingress) string {
	hosts := []string{}

	for _, rule := range ing.Spec.Rules {
		if rule.Host != "" {
			hosts = append(hosts, fmt.Sprintf("Host(`%s`)", rule.Host))
		}
	}

	if len(hosts) == 0 {
		return "PathPrefix(`/`)"
	}

	return strings.Join(hosts, " || ")
}
