package middleware

import (
	"fmt"
	"strings"

	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/models"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* ---------------- APP ROOT ---------------- */

// AppRoot handles the below annotation.
// Annotations:
//   - "nginx.ingress.kubernetes.io/app-root"
func AppRoot(ctx configs.Context) {
	ctx.Log.Debug("running converter AppRoot")

	ann := string(models.AppRoot)
	raw, ok := ctx.Annotations[ann]
	if !ok {
		return
	}

	value := strings.TrimSpace(raw)
	if value == "" {
		msg := fmt.Sprintf("%s is set but empty", ann)
		ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
		ctx.ReportSkipped(ann, msg)

		return
	}

	if !strings.HasPrefix(value, "/") {
		msg := fmt.Sprintf("%s must be an absolute path starting with '/', got %q", ann, raw)
		ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
		ctx.ReportSkipped(ann, msg)

		return
	}

	regex := "^(https?://[^/]+)/?(\\?.*)?$"
	replacement := "${1}" + value + "${2}"

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, &traefik.Middleware{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "Middleware",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mwName(ctx, "app-root"),
			Namespace: ctx.Namespace,
		},
		Spec: traefik.MiddlewareSpec{
			RedirectRegex: &dynamic.RedirectRegex{
				Regex:       regex,
				Replacement: replacement,
				Permanent:   false,
			},
		},
	})

	ctx.ReportConverted(ann)
}
