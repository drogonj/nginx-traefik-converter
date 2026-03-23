package middleware

import (
	"fmt"
	"net/url"
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

	regex := "^((?:https?)://[^/?#]+)(?:/)?(\\?.*)?$"
	replacement := ""

	switch {
	case strings.HasPrefix(value, "/"):
		replacement = "${1}" + value + "${2}"
	case strings.Contains(value, "://"):
		parsed, err := url.Parse(value)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			msg := fmt.Sprintf("%s must be a valid absolute URL or path, got %q", ann, raw)
			ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
			ctx.ReportSkipped(ann, msg)

			return
		}

		replacement = value + "${2}"
	default:
		normalizedHost := strings.TrimRight(value, "/")
		if strings.Contains(normalizedHost, "/") {
			msg := fmt.Sprintf("%s host format must not contain '/', got %q", ann, raw)
			ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
			ctx.ReportSkipped(ann, msg)

			return
		}

		parsed, err := url.Parse("https://" + normalizedHost)
		if err != nil || parsed.Host == "" {
			msg := fmt.Sprintf("%s host format is invalid, got %q", ann, raw)
			ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
			ctx.ReportSkipped(ann, msg)

			return
		}

		// Preserve incoming request scheme for host-only values.
		regex = "^((?:https?)://)[^/?#]+(?:/)?(\\?.*)?$"
		replacement = "${1}" + normalizedHost + "${2}"
	}

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
