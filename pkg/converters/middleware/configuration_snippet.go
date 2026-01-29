package middleware

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

/* ---------------- CONFIGURATION SNIPPET ---------------- */

// ConfigurationSnippet handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/configuration-snippet"
func ConfigurationSnippet(ctx configs.Context) {
	snippet, ok := ctx.Annotations["nginx.ingress.kubernetes.io/configuration-snippet"]
	if !ok || strings.TrimSpace(snippet) == "" {
		return
	}

	reqHeaders, respHeaders, warnings, unsupported :=
		parseConfigurationSnippet(snippet)

	// Emit Warnings (gzip, cache, etc.)
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

func parseConfigurationSnippet(snippet string) (
	reqHeaders map[string]string,
	respHeaders map[string]string,
	warnings []string,
	unsupported []string,
) {
	reqHeaders = map[string]string{}
	respHeaders = map[string]string{}

	lines := strings.Split(snippet, "\n")

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		switch {

		// ───── Headers (convertible) ─────

		case strings.HasPrefix(line, "more_set_headers"),
			strings.HasPrefix(line, "add_header"):
			if h := extractHeader(line); h != nil {
				respHeaders[h[0]] = h[1]
			}

		case strings.HasPrefix(line, "proxy_set_header"):
			// proxy_set_header X-Foo bar;
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				reqHeaders[parts[1]] = strings.TrimSuffix(parts[2], ";")
			}

		// ───── gzip (global-only in Traefik) ─────

		case strings.HasPrefix(line, "gzip "):
			warnings = append(warnings,
				"gzip must be enabled globally in Traefik static configuration",
			)

		case strings.HasPrefix(line, "gzip_comp_level"):
			warnings = append(warnings,
				"gzip_comp_level is not configurable in Traefik and was ignored, compression level is fixed",
			)

		case strings.HasPrefix(line, "gzip_types"):
			warnings = append(warnings,
				"gzip_types is not configurable in Traefik and was ignored. Compresses a fixed, internal set of MIME types",
			)

		// ───── proxy_cache (not supported) ─────

		case strings.HasPrefix(line, "proxy_cache"):
			warnings = append(warnings,
				"proxy_cache is not supported in Traefik OSS and was ignored",
			)

		// ───── Everything else is unsafe ─────

		default:
			unsupported = append(unsupported, line)
		}
	}

	return
}

func extractHeader(line string) []string {
	// expects: "X-Foo: bar"
	start := strings.Index(line, "\"")
	end := strings.LastIndex(line, "\"")
	if start == -1 || end <= start {
		return nil
	}

	kv := strings.SplitN(line[start+1:end], ":", 2)
	if len(kv) != 2 {
		return nil
	}

	return []string{
		strings.TrimSpace(kv[0]),
		strings.TrimSpace(kv[1]),
	}
}
