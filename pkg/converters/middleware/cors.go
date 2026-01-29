package middleware

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
)

/* ---------------- CORS ---------------- */

func CORS(ctx configs.Context) error {
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
