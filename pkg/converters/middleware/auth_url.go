package middleware

import (
	"strings"

	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/models"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* ---------------- AUTH URL ---------------- */

// HandleAuthURL handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/auth-url"
func HandleAuthURL(ctx configs.Context) {
	ctx.Log.Debug("running converter HandleAuthURL")

	const ann = string(models.AuthURL)

	val, ok := ctx.Annotations[ann]
	if !ok || strings.TrimSpace(val) == "" {
		return
	}

	address := strings.TrimSpace(val)

	// Basic sanity check
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		msg := "auth-url must be an absolute URL (http:// or https://)"
		ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
		ctx.ReportSkipped(ann, msg)

		return
	}

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, &traefik.Middleware{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "Middleware",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mwName(ctx, "auth-url"),
			Namespace: ctx.Namespace,
		},
		Spec: traefik.MiddlewareSpec{
			ForwardAuth: &traefik.ForwardAuth{
				Address:            address,
				TrustForwardHeader: true,
			},
		},
	})

	// Warn about partial compatibility / related annotations
	ctx.Result.Warnings = append(ctx.Result.Warnings,
		"auth-url converted to Traefik ForwardAuth middleware; verify headers and auth behavior",
	)

	ctx.ReportConverted(ann)
}
