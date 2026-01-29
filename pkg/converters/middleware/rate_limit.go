package middleware

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

/* ---------------- RATE LIMIT ---------------- */

func RateLimit(ctx configs.Context) {
	rps, ok := ctx.Annotations["nginx.ingress.kubernetes.io/limit-rps"]
	if !ok {
		return
	}

	avg, _ := strconv.Atoi(rps)
	burst := avg * 2

	if m := ctx.Annotations["nginx.ingress.kubernetes.io/limit-burst-multiplier"]; m != "" {
		if v, err := strconv.Atoi(m); err == nil {
			burst = avg * v
		}
	}

	average := int64(avg)
	averageBurst := int64(burst)

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, &traefik.Middleware{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "Middleware",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mwName(ctx, "ratelimit"),
			Namespace: ctx.Namespace,
		},
		Spec: traefik.MiddlewareSpec{
			RateLimit: &traefik.RateLimit{
				Average: &average,
				Burst:   &averageBurst,
			},
		},
	})
}
