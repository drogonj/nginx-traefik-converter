package middleware

import (
	"strings"

	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/models"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* ---------------- REWRITE ---------------- */

// RewriteTargets handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/rewrite-target"
func RewriteTargets(ctx configs.Context) {
	ctx.Log.Debug("running converter RewriteTarget")

	annRewriteTarget := string(models.RewriteTarget)

	val, ok := ctx.Annotations[annRewriteTarget]
	if !ok {
		return
	}

	if strings.Contains(val, "$") {
		ctx.Result.Warnings = append(ctx.Result.Warnings,
			"rewrite-target uses capture groups which cannot be safely converted without path context",
		)

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

	ctx.ReportConverted(annRewriteTarget)
}
