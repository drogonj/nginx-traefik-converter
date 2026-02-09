package middleware

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/converters/models"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* ---------------- PROXY BUFFER SIZE ---------------- */

// ProxyBufferSizes handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/proxy-buffer-size"
func ProxyBufferSizes(ctx configs.Context, opts configs.Options) {
	ctx.Log.Debug("running converter ProxyBufferSize")

	val, ok := ctx.Annotations[string(models.ProxyBufferSize)]
	if !ok {
		return
	}

	// Default: warn + ignore
	if !opts.ProxyBufferHeuristic {
		warningMessage := "proxy-buffer-size has no equivalent in Traefik and was ignored\n  " +
			"Traefik does not expose upstream buffer sizing, it does not buffer responses the same way and uses Go’s HTTP stack"

		ctx.Result.Warnings = append(ctx.Result.Warnings, warningMessage)

		ctx.ReportWarning(string(models.ProxyBufferSize), warningMessage)

		return
	}

	size, err := parseSizeBytes(val)
	if err != nil {
		warningMessage := "proxy-buffer-size value could not be parsed and was ignored"

		ctx.Result.Warnings = append(ctx.Result.Warnings, warningMessage)

		ctx.ReportWarning(string(models.ProxyBufferSize), warningMessage)

		return
	}

	middleware := &traefik.Middleware{
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

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, middleware)

	warningMessage := "proxy-buffer-size was heuristically mapped to Traefik buffering; this is NOT equivalent to NGINX behavior" +
		" Traefik buffering affects response bodies, not headers; verify application behavior"

	ctx.Result.Warnings = append(ctx.Result.Warnings, warningMessage)

	ctx.ReportWarning(string(models.ProxyBufferSize), warningMessage)
}
