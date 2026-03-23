package kubernetes

import (
	"log/slog"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Config holds values required to initialise the kubernetes client.
type Config struct {
	NameSpace     string `json:"name_space,omitempty" yaml:"name_space,omitempty"`
	Context       string `json:"context,omitempty"    yaml:"context,omitempty"`
	All           bool   `json:"all,omitempty"        yaml:"all,omitempty"`
	clientSet     *kubernetes.Clientset
	dynamicClient dynamic.Interface
	logger        *slog.Logger

	// certCache avoids redundant LIST calls when multiple TLS entries
	// reference Certificates in the same namespace.
	certCache map[string]*certCacheEntry
}

// certCacheEntry stores the result of a single LIST Certificates call.
type certCacheEntry struct {
	items []unstructured.Unstructured
	err   error
}

// SetKubeClient sets kube client to Config with specified configurations.
func (cfg *Config) SetKubeClient() error {
	kubeConfig := os.Getenv("KUBECONFIG")

	cfg.logger.Debug("found kubeconfig", slog.Any("kubeConfig", kubeConfig))
	cfg.logger.Debug("using context", slog.Any("context", cfg.Context))

	config, err := buildKubeClientConfig(cfg.Context, kubeConfig, cfg.logger)
	if err != nil {
		cfg.logger.Error("failed to load Kubernetes config", slog.Any("error", err))

		return err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		cfg.logger.Error("failed to create Kubernetes client", slog.Any("error", err))

		return err
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		cfg.logger.Error("failed to create dynamic Kubernetes client", slog.Any("error", err))

		return err
	}

	// Assign both clients only after both succeed to avoid partial init.
	cfg.clientSet = clientSet
	cfg.dynamicClient = dynClient

	return nil
}

// SetKubeNameSpace sets namespace to the initialised client.
func (cfg *Config) SetKubeNameSpace() {
	if cfg.All {
		cfg.NameSpace = ""
	}

	cfg.logger.Debug("using namespace", slog.Any("namespace", cfg.NameSpace))
}

// GetKubeClient returns the configured kube client.
func (cfg *Config) GetKubeClient() *kubernetes.Clientset {
	return cfg.clientSet
}

func buildConfigWithContextFromFlags(kubeContext, kubeConfigPath string) (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if strings.TrimSpace(kubeConfigPath) != "" {
		loadingRules.Precedence = strings.Split(kubeConfigPath, string(os.PathListSeparator))
	}

	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: kubeContext,
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	).ClientConfig()
}

func buildKubeClientConfig(kubeContext, kubeConfigPath string, logger *slog.Logger) (*rest.Config, error) {
	// When context or KUBECONFIG is explicitly provided, keep the current CLI semantics.
	if strings.TrimSpace(kubeContext) != "" || strings.TrimSpace(kubeConfigPath) != "" {
		return buildConfigWithContextFromFlags(kubeContext, kubeConfigPath)
	}

	// Prefer in-cluster auth when running inside Kubernetes with a ServiceAccount.
	config, err := rest.InClusterConfig()
	if err == nil {
		logger.Debug("using in-cluster Kubernetes configuration")

		return config, nil
	}

	logger.Debug("in-cluster configuration unavailable, falling back to default kubeconfig", slog.Any("error", err))

	return buildConfigWithContextFromFlags("", "")
}

// SetLogger sets logger to the Config.
func (cfg *Config) SetLogger(logger *slog.Logger) {
	cfg.logger = logger
}

// New returns new instance of Config when invoked.
func New() *Config {
	return &Config{}
}
