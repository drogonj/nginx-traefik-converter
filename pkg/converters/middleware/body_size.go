package middleware

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

/* ---------------- BODY SIZE ---------------- */

// BodySize handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/proxy-body-size"
func BodySize(ctx configs.Context) error {
	val, ok := ctx.Annotations["nginx.ingress.kubernetes.io/proxy-body-size"]
	if !ok {
		return nil
	}

	intValue, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return err
	}

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, &traefik.Middleware{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "Middleware",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mwName(ctx, "bodysize"),
			Namespace: ctx.Namespace,
		},
		Spec: traefik.MiddlewareSpec{
			Buffering: &dynamic.Buffering{
				MaxRequestBodyBytes: intValue,
			},
		},
	})

	return nil
}
