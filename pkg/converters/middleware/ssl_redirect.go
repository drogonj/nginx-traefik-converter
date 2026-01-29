package middleware

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* ---------------- REDIRECT ---------------- */

func SSLRedirect(ctx configs.Context) {
	ssl := ctx.Annotations["nginx.ingress.kubernetes.io/ssl-redirect"]
	force := ctx.Annotations["nginx.ingress.kubernetes.io/force-ssl-redirect"]

	if ssl != "true" && force != "true" {
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
}
