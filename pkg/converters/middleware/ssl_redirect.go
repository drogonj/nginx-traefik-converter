package middleware

import (
	"fmt"

	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/models"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* ---------------- REDIRECT ---------------- */

// SSLRedirect handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/ssl-redirect"
//   - "nginx.ingress.kubernetes.io/force-ssl-redirect"
func SSLRedirect(ctx configs.Context) {
	ctx.Log.Debug("running converter SSLRedirect")

	annSSLRedirect := string(models.SSLRedirect)
	annForceSslRedirect := string(models.ForceSSLRedirect)

	ssl, annSSLRedirectOk := ctx.Annotations[annSSLRedirect]

	force, annForceSslRedirectOk := ctx.Annotations[annForceSslRedirect]

	if !annSSLRedirectOk && !annForceSslRedirectOk {
		return
	}

	if ssl != "true" && force != "true" {
		if annSSLRedirectOk {
			ctx.ReportSkipped(annSSLRedirect, fmt.Sprintf("%s is not set to true", annSSLRedirect))
		}

		if annForceSslRedirectOk {
			ctx.ReportSkipped(annForceSslRedirect, fmt.Sprintf("%s is not set to true", annForceSslRedirect))
		}

		return
	}

	// When the Ingress has spec.tls, the IngressRoute will use entryPoints=[websecure],
	// so a redirectScheme middleware is useless (traffic is already HTTPS).
	// The HTTP→HTTPS redirect should be handled at the Traefik entrypoint level.
	if len(ctx.Ingress.Spec.TLS) > 0 {
		msg := "IngressRoute already uses entryPoints=[websecure] due to spec.tls; " +
			"configure HTTP→HTTPS redirect via Traefik static config " +
			"(entryPoints.web.http.redirections.entryPoint.to=websecure) " +
			"or create a separate IngressRoute on the web entryPoint with this middleware"

		ctx.Result.Warnings = append(ctx.Result.Warnings, msg)

		if annSSLRedirectOk {
			ctx.ReportSkipped(annSSLRedirect, msg)
		}

		if annForceSslRedirectOk {
			ctx.ReportSkipped(annForceSslRedirect, msg)
		}

		return
	}

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, &traefik.Middleware{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "Middleware",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mwName(ctx, "https-redirect"),
			Namespace: ctx.Namespace,
		},
		Spec: traefik.MiddlewareSpec{
			RedirectScheme: &dynamic.RedirectScheme{
				Scheme:    "https",
				Permanent: true,
			},
		},
	})

	if annSSLRedirectOk {
		ctx.ReportConverted(annSSLRedirect)
	}

	if annForceSslRedirectOk {
		ctx.ReportConverted(annForceSslRedirect)
	}
}
