package ingressroute

import (
	"strings"

	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/models"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/errors"
)

func resolveScheme(annotations map[string]string) (string, error) {
	backendProto := strings.ToUpper(annotations[string(models.BackendProtocol)])
	grpcBackend := annotations[string(models.GrpcBackend)] == "true"

	if grpcBackend && backendProto == "HTTP" {
		return "h2c", nil // but you could also emit a warning via ctx
	}

	// If backend-protocol is explicitly set, it should take precedence
	switch backendProto {
	case "", "HTTP":
		if grpcBackend {
			// gRPC without TLS
			return "h2c", nil
		}

		return "http", nil

	case "HTTPS":
		return "https", nil

	case "GRPC":
		// gRPC without TLS
		return "h2c", nil

	case "GRPCS":
		// gRPC over TLS
		return "https", nil

	default:
		return "", &errors.ConverterError{Message: "unsupported backend-protocol"}
	}
}
