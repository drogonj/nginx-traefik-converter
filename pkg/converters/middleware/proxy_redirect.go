package middleware

import (
	"encoding/json"

	responseHeaders "github.com/jamesmcroft/traefik-plugin-rewrite-response-headers"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/converters/models"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* ---------------- PROXY REDIRECT ---------------- */

// ProxyRedirect handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/proxy-redirect-from"
//   - "nginx.ingress.kubernetes.io/proxy-redirect-to"
func ProxyRedirect(ctx configs.Context) error {
	annRedirectFrom := string(models.ProxyRedirectFrom)
	annRedirectTo := string(models.ProxyRedirectTo)

	redirectFrom, hasFrom := ctx.Annotations[annRedirectFrom]
	redirectTo, hasTo := ctx.Annotations[annRedirectTo]

	if !hasFrom && !hasTo {
		return nil
	}

	pluginConfig := responseHeaders.Config{
		Rewrites: []responseHeaders.Rewrite{
			{
				Header:      "Location",
				Regex:       redirectFrom,
				Replacement: redirectTo,
			},
		},
	}

	raw, err := json.Marshal(pluginConfig)
	if err != nil {
		return err
	}

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, &traefik.Middleware{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "Middleware",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mwName(ctx, "proxy-redirect"),
			Namespace: ctx.Namespace,
		},
		Spec: traefik.MiddlewareSpec{
			Plugin: map[string]apiextv1.JSON{
				"rewriteResponseHeaders": {Raw: raw},
			},
		},
	})

	ctx.ReportConverted(annRedirectFrom)

	ctx.ReportConverted(annRedirectTo)

	return nil
}
