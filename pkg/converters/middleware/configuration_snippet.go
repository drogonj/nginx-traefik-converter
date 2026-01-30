package middleware

import (
	"strings"

	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type unsupportedDirective struct {
	Name       string
	Enterprise bool
	Message    string
}

var unsupported = []unsupportedDirective{
	{"gzip", false, "gzip is only configurable via middleware in Traefik and was ignored"},
	{"gzip_comp_level", false, "gzip_comp_level is not configurable in Traefik"},
	{"gzip_types", false, "gzip_types is not configurable in Traefik"},
	{"proxy_buffer_size", false, "proxy_buffer_size is not supported in Traefik"},
	{"proxy_cache", true, "proxy_cache is not supported in Traefik OSS"},
}

/* ---------------- CONFIGURATION SNIPPET ---------------- */

// ConfigurationSnippet handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/configuration-snippet"
func ConfigurationSnippet(ctx configs.Context) {
	ctx.Log.Debug("running converter ConfigurationSnippet")

	snippet, ok := ctx.Annotations["nginx.ingress.kubernetes.io/configuration-snippet"]
	if !ok || strings.TrimSpace(snippet) == "" {
		return
	}

	reqHeaders, respHeaders, warnings := parseConfigurationSnippet(snippet)

	// Emit Warnings (gzip, cache, etc.)
	ctx.Result.Warnings = append(ctx.Result.Warnings, warnings...)

	// Nothing convertible
	if len(reqHeaders) == 0 && len(respHeaders) == 0 {
		return
	}

	// Create Headers middleware
	middleware := &traefik.Middleware{
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

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, middleware)
}

func parseConfigurationSnippet(snippet string) (map[string]string, map[string]string, []string) {
	reqHeaders := map[string]string{}
	respHeaders := map[string]string{}
	warnings := make([]string, 0)

	lines := strings.Split(snippet, "\n")

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		// ─── OPTIONAL WARNING: auto-correct malformed quoting ───
		if strings.HasSuffix(line, `;"`) || strings.HasSuffix(line, `"`) && !strings.HasSuffix(line, `"`+";") {
			warnings = append(warnings,
				"configuration-snippet contained malformed quoting and was auto-corrected: "+line,
			)
		}

		switch {
		case strings.HasPrefix(line, "add_header"),
			strings.HasPrefix(line, "more_set_headers"):
			if h := extractHeader(line); h != nil {
				respHeaders[h[0]] = h[1]
			} else {
				warnings = append(warnings,
					"failed to parse header directive: "+line,
				)
			}

		case strings.HasPrefix(line, "proxy_set_header"):
			line = strings.TrimSuffix(line, ";")
			parts := strings.Fields(line)

			if len(parts) >= 3 {
				key := strings.Trim(parts[1], `"`)
				value := strings.Join(parts[2:], " ")

				reqHeaders[key] = value
			}

			if strings.Contains(line, "$") {
				warnings = append(warnings,
					"proxy_set_header uses NGINX variables which are not evaluated by Traefik",
				)
			}

		case strings.HasPrefix(line, "gzip "):
			warnUnsupported(&warnings, unsupported[0])
			warnings = append(warnings,
				"gzip must be enabled globally in Traefik static configuration",
			)

		case strings.HasPrefix(line, "gzip_comp_level"):
			warnUnsupported(&warnings, unsupported[1])

		case strings.HasPrefix(line, "gzip_types"):
			warnUnsupported(&warnings, unsupported[2])

		case strings.HasPrefix(line, "proxy_buffer_size"):
			warnUnsupported(&warnings, unsupported[3])

		case strings.HasPrefix(line, "proxy_cache"):
			warnUnsupported(&warnings, unsupported[4])

		default:
			warnings = append(warnings,
				"unsupported directive in configuration-snippet was ignored: "+line,
			)
		}
	}

	return reqHeaders, respHeaders, warnings
}

func extractHeader(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimSuffix(line, ";")
	line = strings.TrimSuffix(line, `"`) // fixes broken YAML quoting

	// ─── more_set_headers "X-Foo: bar" ───

	if strings.HasPrefix(line, "more_set_headers") {
		start := strings.Index(line, `"`)
		end := strings.LastIndex(line, `"`)

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

	// ─── add_header X-Foo bar;
	// ─── add_header "X-Foo" "bar baz";
	fields := strings.Fields(line)
	if len(fields) < 3 || fields[0] != "add_header" {
		return nil
	}

	key := strings.Trim(fields[1], `"`)
	value := strings.Join(fields[2:], " ")
	value = strings.Trim(value, `"`)

	if key == "" || value == "" {
		return nil
	}

	return []string{key, value}
}

func warnUnsupported(warnings *[]string, d unsupportedDirective) {
	msg := d.Message
	if d.Enterprise {
		msg += ". Traefik Enterprise provides an alternative, but it cannot be auto-converted."
	}

	*warnings = append(*warnings, msg)
}
