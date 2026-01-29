package middleware

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* ---------------- PROXY BUFFER SIZE ---------------- */

// HandleProxyBufferSize handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/proxy-buffer-size"
func HandleProxyBufferSize(ctx configs.Context, opts configs.Options) {
	val, ok := ctx.Annotations["nginx.ingress.kubernetes.io/proxy-buffer-size"]
	if !ok {
		return
	}

	// Default: warn + ignore
	if !opts.ProxyBufferHeuristic {
		ctx.Result.Warnings = append(ctx.Result.Warnings,
			"proxy-buffer-size has no equivalent in Traefik and was ignored",
		)
		return
	}

	size, err := parseSizeBytes(val)
	if err != nil {
		ctx.Result.Warnings = append(ctx.Result.Warnings,
			"proxy-buffer-size value could not be parsed and was ignored",
		)
		return
	}

	mw := &traefik.Middleware{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "Middleware",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mwName(ctx, "buffering-heuristic"),
			Namespace: ctx.Namespace,
		},
		Spec: traefik.MiddlewareSpec{
			Buffering: &dynamic.Buffering{
				MaxResponseBodyBytes: size,
				// Intentionally NOT setting MaxRequestBodyBytes
			},
		},
	}

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, mw)

	ctx.Result.Warnings = append(ctx.Result.Warnings,
		"proxy-buffer-size was heuristically mapped to Traefik buffering; this is NOT equivalent to NGINX behavior",
		"Traefik buffering affects response bodies, not headers; verify application behavior",
	)
}
