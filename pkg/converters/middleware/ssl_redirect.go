package middleware

import (
	"fmt"

	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/converters/models"
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
		ctx.ReportSkipped(annSSLRedirect, fmt.Sprintf("%s is not set to true", annSSLRedirect))

		ctx.ReportSkipped(annForceSslRedirect, fmt.Sprintf("%s is not set to true", annForceSslRedirect))

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

	ctx.ReportConverted(annSSLRedirect)

	ctx.ReportConverted(annForceSslRedirect)
}
