package middleware

import (
	"strconv"
	"strings"

	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/models"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

/* ---------------- CORS ---------------- */

// CORS handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/enable-cors"
//   - "nginx.ingress.kubernetes.io/cors-allow-origin"
//   - "nginx.ingress.kubernetes.io/cors-allow-methods"
//   - "nginx.ingress.kubernetes.io/cors-allow-headers"
//   - "nginx.ingress.kubernetes.io/cors-allow-credentials"
//   - "nginx.ingress.kubernetes.io/cors-max-age"
//   - "nginx.ingress.kubernetes.io/cors-expose-headers"
func CORS(ctx configs.Context) error {
	ctx.Log.Debug("running converter CORS")

	val, ok := ctx.Annotations[string(models.EnableCORS)]
	if !ok {
		return nil
	}

	if val != "true" {
		ctx.ReportIgnored(string(models.EnableCORS), "enable-cors was not set to true")

		return nil
	}

	ctx.ReportConverted(string(models.EnableCORS))

	headers := &dynamic.Headers{}

	if v := ctx.Annotations[string(models.CorsAllowOrigin)]; v != "" {
		headers.AccessControlAllowOriginList = headersNeat(v)

		ctx.ReportConverted(string(models.CorsAllowOrigin))
	}

	if v := ctx.Annotations[string(models.CorsAllowMethods)]; v != "" {
		headers.AccessControlAllowMethods = headersNeat(v)

		ctx.ReportConverted(string(models.CorsAllowMethods))
	}

	if v := ctx.Annotations[string(models.CorsAllowHeaders)]; v != "" {
		headers.AccessControlAllowHeaders = headersNeat(v)

		ctx.ReportConverted(string(models.CorsAllowHeaders))
	}

	if v := ctx.Annotations[string(models.CorsAllowCredentials)]; v == "true" {
		headers.AccessControlAllowCredentials = true

		ctx.ReportConverted(string(models.CorsAllowCredentials))
	}

	if v := ctx.Annotations[string(models.CorsMaxAge)]; v != "" {
		secs, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			ctx.ReportWarning(string(models.CorsMaxAge), err.Error())

			return err
		}

		headers.AccessControlMaxAge = secs

		ctx.ReportConverted(string(models.CorsMaxAge))
	}

	if v := ctx.Annotations[string(models.CorsExposeHeaders)]; v != "" {
		headers.AccessControlExposeHeaders = headersNeat(v)

		ctx.ReportConverted(string(models.CorsExposeHeaders))
	}

	ctx.Result.Middlewares = append(ctx.Result.Middlewares,
		newHeadersMiddleware(ctx, "cors", headers),
	)

	return nil
}

func headersNeat(value string) []string {
	headers := strings.Split(value, ",")

	for i, header := range headers {
		headers[i] = strings.TrimSpace(header)
	}

	return headers
}
