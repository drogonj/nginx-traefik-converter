//nolint:mnd
package middleware

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/models"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/errors"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* ---------------- Unsupported directives ---------------- */

type unsupportedDirective struct {
	Enterprise bool
	Message    string
}

type corsConfig struct {
	OriginRegex  string
	AllowHeaders []string
	AllowMethods []string
	AllowCreds   *bool
	MaxAge       int64
}

type conditionalReturnConfig struct {
	Method     string
	StatusCode int
	Headers    map[string]any
}

var unsupported = map[string]unsupportedDirective{
	"gzip": {
		Message: "gzip is only configurable via middleware in Traefik and was ignored",
	},
	"gzip_comp_level": {
		Message: "gzip_comp_level is not configurable in Traefik",
	},
	"gzip_types": {
		Message: "gzip_types is not configurable in Traefik",
	},
	"proxy_buffer_size": {
		Message: "proxy_buffer_size is not supported in Traefik",
	},
	"proxy_cache": {
		Enterprise: true,
		Message:    "proxy_cache is not supported in Traefik OSS",
	},
}

/* ---------------- CONFIGURATION SNIPPET ---------------- */

// ConfigurationSnippets handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/configuration-snippet"
func ConfigurationSnippets(ctx configs.Context) error {
	ctx.Log.Debug("running converter ConfigurationSnippet")

	ann := string(models.ConfigurationSnippet)

	snippet, ok := ctx.Annotations[ann]
	if !ok {
		return nil
	}

	lines := splitLines(snippet)
	if len(lines) == 0 {
		return nil
	}

	// 🔒 Conditional CORS handling
	if isConditionalCORSSnippet(lines) {
		cfg, err := parseConditionalCORSSnippet(lines)
		if err != nil {
			ctx.Result.Warnings = append(ctx.Result.Warnings,
				"failed to parse conditional CORS snippet; skipped",
			)

			return err
		}

		emitCORSMiddleware(ctx, cfg)

		if cr := parseConditionalReturn(lines); cr != nil {
			if err = emitConditionalReturnPlugin(ctx, cr); err != nil {
				return err
			}
		}

		return nil
	}

	convertGenericSnippet(ctx, lines)

	ctx.ReportConverted(ann)

	return nil
}

/* ---------------- Generic snippet handling ---------------- */

func convertGenericSnippet(ctx configs.Context, lines []string) {
	const (
		reqHeadersCount  = 4
		respHeadersCount = 8
		warningsCount    = 4
	)

	reqHeaders := make(map[string]string, reqHeadersCount)
	respHeaders := make(map[string]string, respHeadersCount)
	warnings := make([]string, 0, warningsCount)

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		lower := strings.ToLower(line)

		switch directive(lower) {
		case "add_header", "more_set_headers":
			if k, v, ok := parseResponseHeader(line); ok {
				respHeaders[k] = v
			} else {
				warnings = append(warnings,
					"failed to parse header directive: "+line,
				)
			}

		case "proxy_set_header":
			key, val := parseProxySetHeader(line)
			if key != "" {
				reqHeaders[key] = val
			}

			if strings.Contains(val, "$") {
				warnings = append(warnings,
					"proxy_set_header uses NGINX variables which are not evaluated by Traefik",
				)
			}

		case "gzip", "gzip_comp_level", "gzip_types", "proxy_buffer_size", "proxy_cache":
			if u, ok := unsupported[directive(lower)]; ok {
				warnUnsupported(&warnings, u)
			}

		default:
			warnings = append(warnings,
				"unsupported directive in configuration-snippet was ignored: "+line,
			)
		}
	}

	ctx.Result.Warnings = append(ctx.Result.Warnings, warnings...)

	if len(reqHeaders) == 0 && len(respHeaders) == 0 {
		return
	}

	ctx.Result.Middlewares = append(
		ctx.Result.Middlewares,
		newHeadersMiddleware(ctx, "configuration-snippet", &dynamic.Headers{
			CustomRequestHeaders:  reqHeaders,
			CustomResponseHeaders: respHeaders,
		}),
	)
}

/* ---------------- CORS handling ---------------- */

// NOTE:
// NGINX `if` directives are never converted,
// except when they implement pure CORS logic.
// In that case, Traefik's CORS middleware provides equivalent behavior.
func isConditionalCORSSnippet(lines []string) bool {
	var hasOriginIf, hasMethods bool

	for _, raw := range lines {
		line := strings.ToLower(raw)

		if strings.Contains(line, "if ($http_origin") {
			hasOriginIf = true
		}

		if strings.Contains(line, "access-control-allow-methods") {
			hasMethods = true
		}

		if strings.Contains(line, "rewrite") ||
			strings.Contains(line, "proxy_pass") ||
			strings.Contains(line, "fastcgi") ||
			strings.Contains(line, "lua_") ||
			strings.Contains(line, "set ") {
			return false
		}
	}

	return hasOriginIf && hasMethods
}

func parseConditionalCORSSnippet(lines []string) (*corsConfig, error) {
	cfg := &corsConfig{}

	origin, ok := extractOriginRegex(lines)
	if !ok {
		return nil, &errors.ConverterError{Message: "no origin regex found"}
	}

	cfg.OriginRegex = origin

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		lower := strings.ToLower(line)

		switch {
		case strings.Contains(lower, "access-control-allow-headers"):
			cfg.AllowHeaders = splitCSV(extractQuotedHeaderValue(line))

		case strings.Contains(lower, "access-control-allow-methods"):
			cfg.AllowMethods = splitCSV(extractQuotedHeaderValue(line))

		case strings.Contains(lower, "access-control-allow-credentials"):
			v := strings.ToLower(extractQuotedHeaderValue(line))
			if v == "true" || v == "false" {
				b := v == "true"
				cfg.AllowCreds = &b
			}

		case strings.Contains(lower, "access-control-max-age"):
			if age := extractInt(line); age > 0 {
				cfg.MaxAge = age
			}
		}
	}

	if len(cfg.AllowMethods) == 0 {
		cfg.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}

	return cfg, nil
}

func emitCORSMiddleware(ctx configs.Context, cfg *corsConfig) {
	headers := &dynamic.Headers{
		AccessControlAllowMethods: cfg.AllowMethods,
		AccessControlAllowHeaders: cfg.AllowHeaders,
		AccessControlAllowOriginListRegex: []string{
			cfg.OriginRegex,
		},
		AccessControlMaxAge: cfg.MaxAge,
	}

	if cfg.AllowCreds != nil {
		headers.AccessControlAllowCredentials = *cfg.AllowCreds
	}

	ctx.Result.Middlewares = append(
		ctx.Result.Middlewares,
		newHeadersMiddleware(ctx, "cors", headers),
	)

	if len(cfg.AllowHeaders) == 0 || len(cfg.AllowMethods) == 0 {
		ctx.Result.Warnings = append(ctx.Result.Warnings,
			"conditional CORS snippet was partially parsed; verify generated middleware",
		)
	}

	ctx.Result.Warnings = append(ctx.Result.Warnings,
		"conditional NGINX CORS logic was converted to Traefik CORS middleware",
	)
}

func parseConditionalReturn(lines []string) *conditionalReturnConfig {
	var (
		inIf   bool
		method string
		status int
	)

	headers := make(map[string]any)

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		lower := strings.ToLower(line)

		// Detect: if ($request_method = 'OPTIONS') {
		if strings.HasPrefix(lower, "if") && strings.Contains(lower, "$request_method") {
			if strings.Contains(lower, "options") {
				method = "OPTIONS"
				inIf = true

				continue
			}
		}

		if inIf {
			// Detect return 204;
			if strings.HasPrefix(lower, "return") {
				fields := strings.Fields(lower)
				if len(fields) >= 2 {
					if code, err := strconv.Atoi(strings.TrimSuffix(fields[1], ";")); err == nil {
						status = code
					}
				}

				continue
			}

			//nolint:varnamelen
			// Parse headers properly
			if k, v, ok := parseAddHeaderNormalized(line); ok {
				// Special-case list headers
				switch strings.ToLower(k) {
				case "access-control-allow-headers", "access-control-allow-methods":
					list := splitCSV(v)
					if len(list) > 0 {
						headers[k] = list
					} else {
						headers[k] = v
					}
				default:
					headers[k] = v
				}

				continue
			}

			// End of block
			if strings.HasPrefix(line, "}") {
				inIf = false
			}
		}
	}

	if method != "" && status > 0 {
		return &conditionalReturnConfig{
			Method:     method,
			StatusCode: status,
			Headers:    headers,
		}
	}

	return nil
}

func emitConditionalReturnPlugin(ctx configs.Context, cfg *conditionalReturnConfig) error {
	pluginCfg := map[string]any{
		"rules": []map[string]any{
			{
				"method":     cfg.Method,
				"statusCode": cfg.StatusCode,
				"headers":    cfg.Headers,
			},
		},
	}

	raw, err := json.Marshal(pluginCfg)
	if err != nil {
		return err
	}

	ctx.Result.Middlewares = append(ctx.Result.Middlewares, &traefik.Middleware{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "Middleware",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mwName(ctx, "conditional-return"),
			Namespace: ctx.Namespace,
		},
		Spec: traefik.MiddlewareSpec{
			Plugin: map[string]apiextv1.JSON{
				"conditionalReturn": {Raw: raw},
			},
		},
	})

	return nil
}

func parseAddHeaderNormalized(line string) (string, string, bool) {
	// Trim trailing ;
	line = strings.TrimSpace(strings.TrimSuffix(line, ";"))

	lower := strings.ToLower(line)
	if !strings.HasPrefix(lower, "add_header") && !strings.HasPrefix(lower, "more_set_headers") {
		return "", "", false
	}

	// Remove directive name
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return "", "", false
	}

	rest := strings.TrimSpace(line[len(fields[0]):])

	// Remove "always" if present
	rest = strings.TrimSpace(strings.TrimSuffix(rest, "always"))

	// Now rest should look like:
	// "Header-Name" "Header Value"
	// 'Header-Name' 600
	// Header-Name value

	var key, val string

	// Try quoted key
	if strings.HasPrefix(rest, `"`) || strings.HasPrefix(rest, `'`) {
		quote := rest[0:1]

		end := strings.Index(rest[1:], quote)
		if end == -1 {
			return "", "", false
		}

		end++

		key = rest[1:end]
		after := strings.TrimSpace(rest[end+1:])

		// Value may be quoted or unquoted
		if strings.HasPrefix(after, `"`) || strings.HasPrefix(after, `'`) {
			q := after[0:1]

			e := strings.Index(after[1:], q)
			if e == -1 {
				return "", "", false
			}

			val = after[1 : 1+e]
		} else {
			// Unquoted value (e.g. 600)
			val = strings.TrimSpace(after)
		}
	} else {
		// No quoted key, fallback to fields
		parts := strings.Fields(rest)
		if len(parts) < 2 {
			return "", "", false
		}

		key = strings.Trim(parts[0], `"'`)
		val = strings.Trim(strings.Join(parts[1:], " "), `"'`)
	}

	key = strings.TrimSpace(key)
	val = strings.TrimSpace(val)

	if key == "" || val == "" {
		return "", "", false
	}

	// Handle $http_origin (cannot be evaluated by Traefik)
	if strings.Contains(val, "$http_origin") {
		// Best-effort: let CORS middleware handle dynamic origin
		val = "*"
	}

	return key, val, true
}

/* ---------------- Helpers ---------------- */

func splitLines(s string) []string {
	out := make([]string, 0)

	for _, l := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(l); t != "" {
			out = append(out, t)
		}
	}

	return out
}

func directive(line string) string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return ""
	}

	return fields[0]
}

func warnUnsupported(warnings *[]string, d unsupportedDirective) {
	msg := d.Message
	if d.Enterprise {
		msg += ". Traefik Enterprise provides an alternative, but it cannot be auto-converted."
	}

	*warnings = append(*warnings, msg)
}

func newHeadersMiddleware(
	ctx configs.Context,
	name string,
	headers *dynamic.Headers,
) *traefik.Middleware {
	return &traefik.Middleware{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "Middleware",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mwName(ctx, name),
			Namespace: ctx.Namespace,
		},
		Spec: traefik.MiddlewareSpec{
			Headers: headers,
		},
	}
}

/* ---------------- Parsing helpers ---------------- */

func parseProxySetHeader(line string) (string, string) {
	line = strings.TrimSuffix(line, ";")
	parts := strings.Fields(line)

	const proxySetHeaderCount = 3

	if len(parts) < proxySetHeaderCount {
		return "", ""
	}

	return strings.Trim(parts[1], `"`), strings.Join(parts[2:], " ")
}

func parseResponseHeader(line string) (string, string, bool) {
	line = strings.TrimSuffix(strings.TrimSpace(line), ";")

	if strings.HasPrefix(line, "more_set_headers") {
		start := strings.Index(line, `"`)
		end := strings.LastIndex(line, `"`)

		if start == -1 || end <= start {
			return "", "", false
		}

		const moreSetHeadersCount = 2

		kv := strings.SplitN(line[start+1:end], ":", moreSetHeadersCount)
		if len(kv) != moreSetHeadersCount {
			return "", "", false
		}

		return strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1]), true
	}

	const addHeaderCount = 3

	if strings.HasPrefix(line, "add_header") {
		fields := strings.Fields(line)
		if len(fields) < addHeaderCount {
			return "", "", false
		}

		// Only strip when there are 4+ fields (directive + name + value + always);
		const alwaysMinFields = 4

		if len(fields) >= alwaysMinFields && strings.EqualFold(fields[len(fields)-1], "always") {
			fields = fields[:len(fields)-1]
		}

		if len(fields) < addHeaderCount {
			return "", "", false
		}

		return strings.Trim(fields[1], `"`),
			strings.Trim(strings.Join(fields[2:], " "), `"`),
			true
	}

	return "", "", false
}

var originIfRe = regexp.MustCompile(
	`\$http_origin\s+~\*\s+\((.+?)\)\s*\)`,
)

func extractOriginRegex(lines []string) (string, bool) {
	const originRegexCount = 2

	for _, l := range lines {
		if m := originIfRe.FindStringSubmatch(l); len(m) == originRegexCount {
			return m[1], true
		}
	}

	return "", false
}

func extractQuotedHeaderValue(line string) string {
	values := make([]string, 0)

	for _, quote := range []string{`"`, `'`} {
		tmp := line

		for {
			start := strings.Index(tmp, quote)
			if start == -1 {
				break
			}

			end := strings.Index(tmp[start+1:], quote)
			if end == -1 {
				break
			}

			end = start + 1 + end
			values = append(values, tmp[start+1:end])
			tmp = tmp[end+1:]
		}
	}

	if len(values) == 0 {
		return ""
	}

	return values[len(values)-1]
}

func splitCSV(v string) []string {
	out := make([]string, 0)

	for _, p := range strings.Split(v, ",") {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}

	return out
}

func extractInt(line string) int64 {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return 0
	}

	newValue, _ := strconv.ParseInt(
		strings.TrimSuffix(fields[len(fields)-1], ";"),
		10,
		64,
	)

	return newValue
}
