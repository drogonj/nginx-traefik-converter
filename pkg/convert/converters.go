package convert

import (
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
)

func mwName(ctx Context, suffix string) string {
	return ctx.IngressName + "-" + suffix
}

/* ---------------- REWRITE ---------------- */

func rewriteTarget(ctx Context) {
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

/* ---------------- REDIRECT ---------------- */

func sslRedirect(ctx Context) {
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

/* ---------------- BASIC AUTH ---------------- */

func basicAuth(ctx Context) {
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

/* ---------------- CORS ---------------- */

func cors(ctx Context) error {
	if ctx.Annotations["nginx.ingress.kubernetes.io/enable-cors"] != "true" {
		return nil
	}

	h := &dynamic.Headers{}

	if v := ctx.Annotations["nginx.ingress.kubernetes.io/cors-allow-origin"]; v != "" {
		h.AccessControlAllowOriginList = strings.Split(v, ",")
	}
	if v := ctx.Annotations["nginx.ingress.kubernetes.io/cors-allow-methods"]; v != "" {
		h.AccessControlAllowMethods = strings.Split(v, ",")
	}
	if v := ctx.Annotations["nginx.ingress.kubernetes.io/cors-allow-headers"]; v != "" {
		h.AccessControlAllowHeaders = strings.Split(v, ",")
	}
	if v := ctx.Annotations["nginx.ingress.kubernetes.io/cors-allow-credentials"]; v == "true" {
		h.AccessControlAllowCredentials = true
	}
	if v := ctx.Annotations["nginx.ingress.kubernetes.io/cors-max-age"]; v != "" {
		secs, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}

		h.AccessControlMaxAge = secs
	}
	if v := ctx.Annotations["nginx.ingress.kubernetes.io/cors-expose-headers"]; v != "" {
		h.AccessControlExposeHeaders = strings.Split(v, ",")
	}
	if v := ctx.Annotations["nginx.ingress.kubernetes.io/cors-expose-headers"]; v != "" {
		h.AccessControlExposeHeaders = strings.Split(v, ",")
	}

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, &traefik.Middleware{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "Middleware",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mwName(ctx, "cors"),
			Namespace: ctx.Namespace,
		},
		Spec: traefik.MiddlewareSpec{Headers: h},
	})

	return nil
}

/* ---------------- RATE LIMIT ---------------- */

func rateLimit(ctx Context) {
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

/* ---------------- BODY SIZE ---------------- */

func bodySize(ctx Context) error {
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

/* ---------------- CONFIGURATION SNIPPET ---------------- */

func configurationSnippet(ctx Context) {
	snippet, ok := ctx.Annotations["nginx.ingress.kubernetes.io/configuration-snippet"]
	if !ok || strings.TrimSpace(snippet) == "" {
		return
	}

	reqHeaders, respHeaders, warnings, unsupported :=
		parseConfigurationSnippet(snippet)

	// Emit warnings (gzip, cache, etc.)
	ctx.Result.Warnings = append(ctx.Result.Warnings, warnings...)

	// If there are unsupported directives (rewrite, lua, etc), do NOT convert
	if len(unsupported) > 0 {
		ctx.Result.Warnings = append(ctx.Result.Warnings,
			"configuration-snippet contains unsupported NGINX directives and was skipped",
		)
		return
	}

	// Nothing convertible
	if len(reqHeaders) == 0 && len(respHeaders) == 0 {
		return
	}

	// Create Headers middleware
	mw := &traefik.Middleware{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "Middleware",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mwName(ctx, "snippet-headers"),
			Namespace: ctx.Namespace,
		},
		Spec: traefik.MiddlewareSpec{
			Headers: &dynamic.Headers{
				CustomRequestHeaders:  reqHeaders,
				CustomResponseHeaders: respHeaders,
			},
		},
	}

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, mw)
}

/* ---------------- UNSUPPORTED/REDUNDANT ANNOTATIONS ---------------- */

func extraAnnotations(ctx Context) {
	if _, ok := ctx.Annotations["nginx.ingress.kubernetes.io/proxy-buffer-size"]; ok {
		ctx.Result.Warnings = append(ctx.Result.Warnings,
			"proxy-buffer-size has no Traefik equivalent",
		)
	}

	if ctx.Annotations["nginx.ingress.kubernetes.io/proxy-buffering"] == "off" {
		ctx.Result.Warnings = append(ctx.Result.Warnings,
			"proxy-buffering=off is default behavior in Traefik",
		)
	}

	if ctx.Annotations["nginx.ingress.kubernetes.io/service-upstream"] == "true" {
		ctx.Result.Warnings = append(ctx.Result.Warnings,
			"service-upstream=true is default behavior in Traefik",
		)
	}

	if ctx.Annotations["nginx.ingress.kubernetes.io/enable-opentracing"] == "true" {
		ctx.Result.Warnings = append(
			ctx.Result.Warnings,
			"enable-opentracing is global in Traefik and cannot be enabled per Ingress",
		)
	}

	if ctx.Annotations["nginx.ingress.kubernetes.io/enable-opentelemetry"] == "true" {
		ctx.Result.Warnings = append(
			ctx.Result.Warnings,
			"enable-opentelemetry must be configured globally in Traefik static config"+`tracing:
  otlp:
    grpc:
      endpoint: otel-collector:4317`,
		)
	}

	if v := ctx.Annotations["nginx.ingress.kubernetes.io/backend-protocol"]; v != "" {
		ctx.Result.Warnings = append(ctx.Result.Warnings,
			"backend-protocol must be applied to IngressRoute service scheme, check for generated ingressroutes.yaml",
		)
	}

	if ctx.Annotations["nginx.ingress.kubernetes.io/grpc-backend"] == "true" {
		ctx.Result.Warnings = append(ctx.Result.Warnings,
			"grpc-backend requires IngressRoute service scheme h2c or https+h2, check for generated ingressroutes.yaml",
		)
	}
}

/* ---------------- WARNINGS ---------------- */

func warnings(ctx Context) {
	for k := range ctx.Annotations {
		if strings.Contains(k, "auth-tls") ||
			strings.Contains(k, "snippet") ||
			strings.Contains(k, "proxy-read") ||
			strings.Contains(k, "proxy-send") {
			ctx.Result.Warnings = append(ctx.Result.Warnings, k+" is not safely convertible")
		}
	}
}
