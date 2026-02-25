package middleware

import (
	"strings"

	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/models"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* ------------- WHITELIST SOURCE RANGE ------------- */

// WhitelistSourceRange converts the nginx whitelist-source-range annotation
// into a Traefik IPAllowList middleware.
//
// Annotations:
//   - "nginx.ingress.kubernetes.io/whitelist-source-range"
//
// The annotation value is a comma-separated list of CIDRs or IPs.
// Example: "10.0.0.0/8,172.16.0.0/12,192.168.0.1"
func WhitelistSourceRange(ctx configs.Context) {
	ctx.Log.Debug("running converter WhitelistSourceRange")

	ann := string(models.WhitelistSourceRange)

	val, ok := ctx.Annotations[ann]
	if !ok {
		return
	}

	ranges := splitAndTrim(val)
	if len(ranges) == 0 {
		return
	}

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, &traefik.Middleware{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "Middleware",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mwName(ctx, "ipallowlist"),
			Namespace: ctx.Namespace,
		},
		Spec: traefik.MiddlewareSpec{
			IPAllowList: &dynamic.IPAllowList{
				SourceRange: ranges,
			},
		},
	})

	ctx.ReportConverted(ann)
}

// splitAndTrim splits a comma-separated string and trims whitespace from
// each element. Empty entries are discarded.
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}

	return out
}
