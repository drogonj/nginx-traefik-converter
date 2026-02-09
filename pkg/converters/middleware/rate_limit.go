package middleware

import (
	"strconv"

	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/converters/models"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* ---------------- RATE LIMIT ---------------- */

// RateLimit handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/limit-rps"
//   - "nginx.ingress.kubernetes.io/limit-burst-multiplier"
func RateLimit(ctx configs.Context) error {
	ctx.Log.Debug("running converter RateLimit")

	annLimitRPS := string(models.LimitRPS)
	annLimitBurstMultiplier := string(models.LimitBurstMultiplier)

	rps, ok := ctx.Annotations[annLimitRPS]
	if !ok {
		return nil
	}

	const averageValue = 2

	avg, err := strconv.Atoi(rps)
	if err != nil {
		ctx.ReportWarning(annLimitRPS, err.Error())
		ctx.ReportWarning(annLimitBurstMultiplier, err.Error())

		return err
	}

	burst := avg * averageValue

	if m := ctx.Annotations[annLimitBurstMultiplier]; m != "" {
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

	ctx.ReportConverted(annLimitRPS)
	ctx.ReportConverted(annLimitBurstMultiplier)

	return nil
}
