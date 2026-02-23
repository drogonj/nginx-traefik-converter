package middleware

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	responseHeaders "github.com/jamesmcroft/traefik-plugin-rewrite-response-headers"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/models"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* ---------------- PROXY COOKIE PATH ---------------- */

// ProxyCookiePath handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/proxy-cookie-path"
func ProxyCookiePath(ctx configs.Context) error {
	ctx.Log.Debug("running converter ProxyCookiePath")

	const ann = string(models.ProxyCookiePath)

	val, ok := ctx.Annotations[ann] //nolint:varnamelen
	if !ok || strings.TrimSpace(val) == "" {
		return nil
	}

	val = normalizeWhitespace(val)

	// NGINX format: "<from> <to>"
	fromPath, toPath, ok := parseTwoArgs(val)
	if !ok {
		msg := "proxy-cookie-path has invalid format, expected: '<from> <to>' (quotes required if values contain spaces)"

		ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
		ctx.ReportSkipped(ann, msg)

		return nil
	}

	// If plugins are not enabled, we cannot safely convert this
	if ctx.Options.DisablePlugins {
		msg := "proxy-cookie-path has no native Traefik equivalent; requires a response header rewrite plugin or backend change"

		ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
		ctx.ReportSkipped(ann, msg)

		return nil
	}

	// Build regex to rewrite Set-Cookie Path attribute
	// Example: (.*)Path=/backend(.*)  ->  $1Path=/$2
	regex := fmt.Sprintf(`(.*?)(Path=%s)(.*)`, regexp.QuoteMeta(fromPath))
	replacement := fmt.Sprintf(`$1Path=%s$3`, toPath)

	mw, err := newRewriteResponseHeadersMiddleware(ctx, "Set-Cookie", regex, replacement, "proxy-cookie-path")
	if err != nil {
		return err
	}

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, mw)

	ctx.ReportConverted(ann)

	return nil
}

func normalizeWhitespace(s string) string {
	// Replace all whitespace (including newlines, tabs) with single spaces
	fields := strings.Fields(s)

	return strings.Join(fields, " ")
}

//nolint:mnd
func parseTwoArgs(val string) (string, string, bool) {
	val = strings.TrimSpace(val)

	// If the whole value is wrapped in quotes, unwrap once.
	// Example: "\"/backend /\""  ->  "/backend /"
	if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
		val = strings.TrimSpace(val[1 : len(val)-1])
	}

	// Case A: both args quoted: "from" "to with spaces"
	// Example: "/" "/platform/oauth/; HTTPOnly; Secure; Domain=my.test.com"
	if strings.HasPrefix(val, `"`) {
		rest := val[1:]

		index := strings.Index(rest, `"`)
		if index == -1 {
			return "", "", false
		}

		from := rest[:index]
		rest = strings.TrimSpace(rest[index+1:])

		if !strings.HasPrefix(rest, `"`) {
			return "", "", false
		}

		rest = rest[1:]

		j := strings.LastIndex(rest, `"`)
		if j == -1 {
			return "", "", false
		}

		to := rest[:j]

		return from, to, true
	}

	// Case B: unquoted first, quoted second
	// Example: / "/platform/oauth/; HTTPOnly; Secure; Domain=my.test.com"
	if strings.Contains(val, `"`) {
		index := strings.IndexAny(val, " \t")
		if index == -1 {
			return "", "", false
		}

		from := strings.TrimSpace(val[:index])
		rest := strings.TrimSpace(val[index:])

		if !strings.HasPrefix(rest, `"`) || !strings.HasSuffix(rest, `"`) {
			return "", "", false
		}

		to := strings.TrimSuffix(strings.TrimPrefix(rest, `"`), `"`)

		return from, to, true
	}

	// Case C: simple space separated: from to
	parts := strings.Fields(val)
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

func newRewriteResponseHeadersMiddleware(ctx configs.Context, header, regex, replacement, suffix string) (*traefik.Middleware, error) {
	pluginConfig := responseHeaders.Config{
		Rewrites: []responseHeaders.Rewrite{
			{
				Header:      header,
				Regex:       regex,
				Replacement: replacement,
			},
		},
	}

	raw, err := json.Marshal(pluginConfig)
	if err != nil {
		return nil, err
	}

	middleware := &traefik.Middleware{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "Middleware",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mwName(ctx, suffix),
			Namespace: ctx.Namespace,
		},
		Spec: traefik.MiddlewareSpec{
			Plugin: map[string]apiextv1.JSON{
				"rewriteResponseHeaders": {Raw: raw},
			},
		},
	}

	return middleware, nil
}
