package middleware

import (
	"strconv"

	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/models"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NGINX default burst multiplier when limit-burst-multiplier is not set.
const defaultBurstMultiplier = 5

/* ---------------- RATE LIMIT ---------------- */

// RateLimit handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/limit-rps"
//   - "nginx.ingress.kubernetes.io/limit-rpm"
//   - "nginx.ingress.kubernetes.io/limit-burst-multiplier"
func RateLimit(ctx configs.Context) error {
	ctx.Log.Debug("running converter RateLimit")

	annLimitRPS := string(models.LimitRPS)
	annLimitRPM := string(models.LimitRPM)
	annLimitBurstMultiplier := string(models.LimitBurstMultiplier)

	rpsStr, hasRPS := ctx.Annotations[annLimitRPS]
	rpmStr, hasRPM := ctx.Annotations[annLimitRPM]

	if !hasRPS && !hasRPM {
		return nil
	}

	burstMultiplier := defaultBurstMultiplier
	if m := ctx.Annotations[annLimitBurstMultiplier]; m != "" {
		if v, err := strconv.Atoi(m); err == nil {
			burstMultiplier = v
		}
	}

	// In NGINX, limit-rps and limit-rpm apply simultaneously.
	// Generate separate middlewares for each.
	if hasRPS {
		avg, err := strconv.Atoi(rpsStr)
		if err != nil {
			ctx.ReportWarning(annLimitRPS, err.Error())

			return err
		}

		average := int64(avg)
		burst := int64(avg * burstMultiplier)

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
					Burst:   &burst,
				},
			},
		})

		ctx.ReportConverted(annLimitRPS)
	}

	// limit-rpm: requests per minute → Traefik period=1m
	if hasRPM {
		avg, err := strconv.Atoi(rpmStr)
		if err != nil {
			ctx.ReportWarning(annLimitRPM, err.Error())

			return err
		}

		average := int64(avg)
		burst := int64(avg * burstMultiplier)
		period := intstr.FromString("1m")

		// Use a distinct name when both rps and rpm are present
		name := "ratelimit"
		if hasRPS {
			name = "ratelimit-rpm"
		}

		ctx.Result.Middlewares = append(ctx.Result.Middlewares, &traefik.Middleware{
			TypeMeta: metav1.TypeMeta{
				APIVersion: traefik.SchemeGroupVersion.String(),
				Kind:       "Middleware",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      mwName(ctx, name),
				Namespace: ctx.Namespace,
			},
			Spec: traefik.MiddlewareSpec{
				RateLimit: &traefik.RateLimit{
					Average: &average,
					Period:  &period,
					Burst:   &burst,
				},
			},
		})

		ctx.ReportConverted(annLimitRPM)
	}

	if ctx.Annotations[annLimitBurstMultiplier] != "" {
		ctx.ReportConverted(annLimitBurstMultiplier)
	}

	return nil
}

/* ---------------- LIMIT CONNECTIONS (InFlightReq) ---------------- */

// LimitConnections handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/limit-connections"
//
// NGINX limit-connections limits concurrent connections from a single IP.
// Traefik InFlightReq middleware provides equivalent functionality.
func LimitConnections(ctx configs.Context) error {
	ctx.Log.Debug("running converter LimitConnections")

	ann := string(models.LimitConnections)

	val, ok := ctx.Annotations[ann]
	if !ok {
		return nil
	}

	amount, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		ctx.ReportWarning(ann, "invalid limit-connections value: "+val)

		return nil
	}

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, &traefik.Middleware{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "Middleware",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mwName(ctx, "inflightreq"),
			Namespace: ctx.Namespace,
		},
		Spec: traefik.MiddlewareSpec{
			InFlightReq: &dynamic.InFlightReq{
				Amount: amount,
				SourceCriterion: &dynamic.SourceCriterion{
					IPStrategy: &dynamic.IPStrategy{},
				},
			},
		},
	})

	ctx.ReportConverted(ann)

	return nil
}

/* ---------------- PROXY TIMEOUTS (warnings) ---------------- */

// ProxyTimeouts handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/proxy-read-timeout"
//   - "nginx.ingress.kubernetes.io/proxy-send-timeout"
//
// These cannot be set per-route in Traefik.
// They should be configured via ServersTransport or static config.
func ProxyTimeouts(ctx configs.Context) {
	ctx.Log.Debug("running converter ProxyTimeouts")

	annRead := string(models.ProxyReadTimeout)
	annSend := string(models.ProxySendTimeout)

	if _, ok := ctx.Annotations[annRead]; ok {
		msg := "proxy-read-timeout cannot be set per-route in Traefik; " +
			"configure via ServersTransport forwardingTimeouts.responseHeaderTimeout in dynamic config"

		ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
		ctx.ReportSkipped(annRead, msg)
	}

	if _, ok := ctx.Annotations[annSend]; ok {
		msg := "proxy-send-timeout cannot be set per-route in Traefik; " +
			"configure via ServersTransport forwardingTimeouts in dynamic config"

		ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
		ctx.ReportSkipped(annSend, msg)
	}
}
