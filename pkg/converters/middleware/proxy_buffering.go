package middleware

import (
	"fmt"
	"strings"

	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/models"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProxyBuffering handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/proxy-buffering"
//
// When set to "on", NGINX buffers responses from the proxied server.
// Traefik's Buffering middleware provides equivalent functionality.
func ProxyBuffering(ctx configs.Context) {
	ctx.Log.Debug("running converter ProxyBuffering")

	ann := string(models.ProxyBuffering)

	val, ok := ctx.Annotations[ann]
	if !ok {
		return
	}

	v := strings.ToLower(strings.TrimSpace(val))

	switch v {
	case "on":
		ctx.Result.Middlewares = append(ctx.Result.Middlewares, &traefik.Middleware{
			TypeMeta: metav1.TypeMeta{
				APIVersion: traefik.SchemeGroupVersion.String(),
				Kind:       "Middleware",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      mwName(ctx, "buffering"),
				Namespace: ctx.Namespace,
			},
			Spec: traefik.MiddlewareSpec{
				Buffering: &dynamic.Buffering{},
			},
		})

		ctx.ReportConverted(ann)

	case "off":
		ctx.ReportIgnored(ann, "proxy-buffering=off is default behavior in Traefik")

	default:
		warningMessage := fmt.Sprintf(
			"proxy-buffering has unknown value %q and was ignored", val)

		ctx.Result.Warnings = append(ctx.Result.Warnings, warningMessage)

		ctx.ReportIgnored(ann, warningMessage)
	}
}
