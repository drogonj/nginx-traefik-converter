package middleware

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* ---------------- REWRITE ---------------- */

// RewriteTarget handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/rewrite-target"
func RewriteTarget(ctx configs.Context) {
	val, ok := ctx.Annotations["nginx.ingress.kubernetes.io/rewrite-target"]
	if !ok {
		return
	}

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, &traefik.Middleware{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "Middleware",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mwName(ctx, "rewrite"),
			Namespace: ctx.Namespace,
		},
		Spec: traefik.MiddlewareSpec{
			ReplacePathRegex: &dynamic.ReplacePathRegex{
				Regex:       "^(.*)",
				Replacement: val,
			},
		},
	})
}
