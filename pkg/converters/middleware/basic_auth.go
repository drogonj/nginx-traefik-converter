package middleware

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* ---------------- BASIC AUTH ---------------- */

func BasicAuth(ctx configs.Context) {
	if ctx.Annotations["nginx.ingress.kubernetes.io/auth-type"] != "basic" {
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
				Secret: ctx.Annotations["nginx.ingress.kubernetes.io/auth-secret"],
				Realm:  ctx.Annotations["nginx.ingress.kubernetes.io/auth-realm"],
			},
		},
	})
}
