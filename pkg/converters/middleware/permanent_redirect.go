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

/* ---------------- PERMANENT REDIRECT ---------------- */

// PermanentRedirect handles the below annotation.
// Annotations:
//   - "nginx.ingress.kubernetes.io/permanent-redirect"
func PermanentRedirect(ctx configs.Context) {
	ctx.Log.Debug("running converter PermanentRedirect")

	ann := string(models.PermanentRedirect)
	rawTarget, ok := ctx.Annotations[ann]
	if !ok {
		return
	}

	target := strings.TrimSpace(rawTarget)
	if target == "" {
		msg := fmt.Sprintf("%s is set but empty", ann)
		ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
		ctx.ReportSkipped(ann, msg)

		return
	}

	parsed, err := url.Parse(target)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		msg := fmt.Sprintf("%s must be a valid absolute URL, got %q", ann, rawTarget)
		ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
		ctx.ReportSkipped(ann, msg)

		return
	}

	replacement := target
	if !strings.Contains(target, "${1}") && !strings.Contains(target, "$1") {
		// Preserve request path when the redirect target does not already use capture groups.
		replacement += "${1}"
	}

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, &traefik.Middleware{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "Middleware",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mwName(ctx, "permanent-redirect"),
			Namespace: ctx.Namespace,
		},
		Spec: traefik.MiddlewareSpec{
			RedirectRegex: &dynamic.RedirectRegex{
				Regex:       "^https?://[^/]+(/.*)?$",
				Replacement: replacement,
				Permanent:   true,
			},
		},
	})

	ctx.ReportConverted(ann)
}
