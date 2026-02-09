package models

type Annotation string

const (
	AuthType             Annotation = "nginx.ingress.kubernetes.io/auth-type"
	AuthSecret           Annotation = "nginx.ingress.kubernetes.io/auth-secret" //nolint:gosec
	AuthRealm            Annotation = "nginx.ingress.kubernetes.io/auth-realm"
	ProxyBodySize        Annotation = "nginx.ingress.kubernetes.io/proxy-body-size"
	ConfigurationSnippet Annotation = "nginx.ingress.kubernetes.io/configuration-snippet"
	EnableCORS           Annotation = "nginx.ingress.kubernetes.io/enable-cors"
	CorsAllowOrigin      Annotation = "nginx.ingress.kubernetes.io/cors-allow-origin"
	CorsAllowMethods     Annotation = "nginx.ingress.kubernetes.io/cors-allow-methods"
	CorsAllowHeaders     Annotation = "nginx.ingress.kubernetes.io/cors-allow-headers"
	CorsAllowCredentials Annotation = "nginx.ingress.kubernetes.io/cors-allow-credentials" //nolint:gosec
	CorsMaxAge           Annotation = "nginx.ingress.kubernetes.io/cors-max-age"
	CorsExposeHeaders    Annotation = "nginx.ingress.kubernetes.io/cors-expose-headers"
	ProxyBuffering       Annotation = "nginx.ingress.kubernetes.io/proxy-buffering"
	ServiceUpstream      Annotation = "nginx.ingress.kubernetes.io/service-upstream"
	EnableOpentracing    Annotation = "nginx.ingress.kubernetes.io/enable-opentracing"
	EnableOpentelemetry  Annotation = "nginx.ingress.kubernetes.io/enable-opentelemetry"
	BackendProtocol      Annotation = "nginx.ingress.kubernetes.io/backend-protocol"
	GrpcBackend          Annotation = "nginx.ingress.kubernetes.io/grpc-backend"
	ProxyBufferSize      Annotation = "nginx.ingress.kubernetes.io/proxy-buffer-size"
	LimitRPS             Annotation = "nginx.ingress.kubernetes.io/limit-rps"
	LimitBurstMultiplier Annotation = "nginx.ingress.kubernetes.io/limit-burst-multiplier"
	RewriteTarget        Annotation = "nginx.ingress.kubernetes.io/rewrite-target"
	SSLRedirect          Annotation = "nginx.ingress.kubernetes.io/ssl-redirect"
	ForceSSLRedirect     Annotation = "nginx.ingress.kubernetes.io/force-ssl-redirect"
	UpstreamVhost        Annotation = "nginx.ingress.kubernetes.io/upstream-vhost"
	ProxyRedirectFrom    Annotation = "nginx.ingress.kubernetes.io/proxy-redirect-from"
	ProxyRedirectTo      Annotation = "nginx.ingress.kubernetes.io/proxy-redirect-to"
	ServerSnippet        Annotation = "nginx.ingress.kubernetes.io/server-snippet"
)

var AllAnnotations = []Annotation{
	AuthType,
	AuthSecret,
	AuthRealm,
	ProxyBodySize,
	ConfigurationSnippet,
	EnableCORS,
	CorsAllowOrigin,
	CorsAllowMethods,
	CorsAllowHeaders,
	CorsAllowCredentials,
	CorsMaxAge,
	CorsExposeHeaders,
	ProxyBuffering,
	ServiceUpstream,
	EnableOpentracing,
	EnableOpentelemetry,
	BackendProtocol,
	GrpcBackend,
	ProxyBufferSize,
	LimitRPS,
	LimitBurstMultiplier,
	RewriteTarget,
	SSLRedirect,
	ForceSSLRedirect,
	UpstreamVhost,
	ProxyRedirectFrom,
	ProxyRedirectTo,
	ServerSnippet,
}

func (a Annotation) String() string {
	return string(a)
}

func GetAnnotations() []string {
	annotations := make([]string, 0)

	for _, annotation := range AllAnnotations {
		annotations = append(annotations, string(annotation))
	}

	return annotations
}
