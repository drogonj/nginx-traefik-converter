package middleware

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/converters/models"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* ---------------- BASIC AUTH ---------------- */

// BasicAuth handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/auth-type"
//   - "nginx.ingress.kubernetes.io/auth-secret"
//   - "nginx.ingress.kubernetes.io/auth-realm"
func BasicAuth(ctx configs.Context) {
	ctx.Log.Debug("running converter BasicAuth")

	val, ok := ctx.Annotations[string(models.AuthType)]
	if !ok {
		return
	}

	if val != "basic" {
		ctx.ReportSkipped(string(models.AuthType), "not of type basic")

		return
	}

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, &traefik.Middleware{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "Middleware",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mwName(ctx, "basicauth"),
			Namespace: ctx.Namespace,
		},
		Spec: traefik.MiddlewareSpec{
			BasicAuth: &traefik.BasicAuth{
				Secret: ctx.Annotations[string(models.AuthSecret)],
				Realm:  ctx.Annotations[string(models.AuthRealm)],
			},
		},
	})

	ctx.ReportConverted(string(models.AuthSecret))

	ctx.ReportConverted(string(models.AuthRealm))
}
