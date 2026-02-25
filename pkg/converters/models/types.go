package models

type Annotation string

const (
	AuthType                 Annotation = "nginx.ingress.kubernetes.io/auth-type"
	AuthSecret               Annotation = "nginx.ingress.kubernetes.io/auth-secret" //nolint:gosec
	AuthRealm                Annotation = "nginx.ingress.kubernetes.io/auth-realm"
	AuthTLSVerifyClient      Annotation = "nginx.ingress.kubernetes.io/auth-tls-verify-client"
	AuthTLSSecret            Annotation = "nginx.ingress.kubernetes.io/auth-tls-secret" //nolint:gosec
	AuthURL                  Annotation = "nginx.ingress.kubernetes.io/auth-url"
	ProxyBodySize            Annotation = "nginx.ingress.kubernetes.io/proxy-body-size"
	ConfigurationSnippet     Annotation = "nginx.ingress.kubernetes.io/configuration-snippet"
	EnableCORS               Annotation = "nginx.ingress.kubernetes.io/enable-cors"
	CorsAllowOrigin          Annotation = "nginx.ingress.kubernetes.io/cors-allow-origin"
	CorsAllowMethods         Annotation = "nginx.ingress.kubernetes.io/cors-allow-methods"
	CorsAllowHeaders         Annotation = "nginx.ingress.kubernetes.io/cors-allow-headers"
	CorsAllowCredentials     Annotation = "nginx.ingress.kubernetes.io/cors-allow-credentials" //nolint:gosec
	CorsMaxAge               Annotation = "nginx.ingress.kubernetes.io/cors-max-age"
	CorsExposeHeaders        Annotation = "nginx.ingress.kubernetes.io/cors-expose-headers"
	ProxyBuffering           Annotation = "nginx.ingress.kubernetes.io/proxy-buffering"
	ServiceUpstream          Annotation = "nginx.ingress.kubernetes.io/service-upstream"
	EnableOpentracing        Annotation = "nginx.ingress.kubernetes.io/enable-opentracing"
	EnableOpentelemetry      Annotation = "nginx.ingress.kubernetes.io/enable-opentelemetry"
	BackendProtocol          Annotation = "nginx.ingress.kubernetes.io/backend-protocol"
	GrpcBackend              Annotation = "nginx.ingress.kubernetes.io/grpc-backend"
	ProxyBufferSize          Annotation = "nginx.ingress.kubernetes.io/proxy-buffer-size"
	LimitConnections         Annotation = "nginx.ingress.kubernetes.io/limit-connections"
	LimitRPS                 Annotation = "nginx.ingress.kubernetes.io/limit-rps"
	LimitRPM                 Annotation = "nginx.ingress.kubernetes.io/limit-rpm"
	LimitBurstMultiplier     Annotation = "nginx.ingress.kubernetes.io/limit-burst-multiplier"
	ProxyReadTimeout         Annotation = "nginx.ingress.kubernetes.io/proxy-read-timeout"
	ProxySendTimeout         Annotation = "nginx.ingress.kubernetes.io/proxy-send-timeout"
	RewriteTarget            Annotation = "nginx.ingress.kubernetes.io/rewrite-target"
	SSLRedirect              Annotation = "nginx.ingress.kubernetes.io/ssl-redirect"
	ForceSSLRedirect         Annotation = "nginx.ingress.kubernetes.io/force-ssl-redirect"
	UpstreamVhost            Annotation = "nginx.ingress.kubernetes.io/upstream-vhost"
	ProxyRedirectFrom        Annotation = "nginx.ingress.kubernetes.io/proxy-redirect-from"
	ProxyRedirectTo          Annotation = "nginx.ingress.kubernetes.io/proxy-redirect-to"
	ProxyCookiePath          Annotation = "nginx.ingress.kubernetes.io/proxy-cookie-path"
	ServerSnippet            Annotation = "nginx.ingress.kubernetes.io/server-snippet"
	UnderscoresInHeaders     Annotation = "nginx.ingress.kubernetes.io/enable-underscores-in-headers"
	UseRegex                 Annotation = "nginx.ingress.kubernetes.io/use-regex"
	ClientHeaderBufferSize   Annotation = "nginx.ingress.kubernetes.io/client-header-buffer-size"
	LargeClientHeaderBuffers Annotation = "nginx.ingress.kubernetes.io/large-client-header-buffers"

	// cert-manager annotations (used by ingress-shim to auto-create Certificate resources).
	CertManagerClusterIssuer Annotation = "cert-manager.io/cluster-issuer"
	CertManagerIssuer        Annotation = "cert-manager.io/issuer"
	CertManagerIssuerKind    Annotation = "cert-manager.io/issuer-kind"
	CertManagerIssuerGroup   Annotation = "cert-manager.io/issuer-group"
	CertManagerCommonName    Annotation = "cert-manager.io/common-name"
	CertManagerDuration      Annotation = "cert-manager.io/duration"
	CertManagerRenewBefore   Annotation = "cert-manager.io/renew-before"
)

// NginxAnnotations contains all supported nginx ingress controller annotations.
var NginxAnnotations = []Annotation{
	AuthType,
	AuthSecret,
	AuthRealm,
	AuthTLSVerifyClient,
	AuthTLSSecret,
	AuthURL,
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
	LimitConnections,
	LimitRPS,
	LimitRPM,
	LimitBurstMultiplier,
	ProxyReadTimeout,
	ProxySendTimeout,
	RewriteTarget,
	SSLRedirect,
	ForceSSLRedirect,
	UpstreamVhost,
	ProxyRedirectFrom,
	ProxyRedirectTo,
	ProxyCookiePath,
	ServerSnippet,
	UnderscoresInHeaders,
	UseRegex,
	ClientHeaderBufferSize,
	LargeClientHeaderBuffers,
}

// CertManagerAnnotations contains cert-manager annotations that the converter
// uses to extract or generate Certificate resources.
var CertManagerAnnotations = []Annotation{
	CertManagerClusterIssuer,
	CertManagerIssuer,
	CertManagerIssuerKind,
	CertManagerIssuerGroup,
	CertManagerCommonName,
	CertManagerDuration,
	CertManagerRenewBefore,
}

// AllAnnotations is the union of all annotation families recognised by the converter.
var AllAnnotations = append(NginxAnnotations, CertManagerAnnotations...)

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
