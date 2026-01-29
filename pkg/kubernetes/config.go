package kubernetes

import (
	"log/slog"
	"os"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Config holds values required to initialise the kubernetes client.
type Config struct {
	NameSpace string `json:"name_space,omitempty" yaml:"name_space,omitempty"`
	Context   string `json:"context,omitempty"    yaml:"context,omitempty"`
	All       bool   `json:"all,omitempty"        yaml:"all,omitempty"`
	clientSet *kubernetes.Clientset
	logger    *slog.Logger
}

// SetKubeClient sets kube client to Config with specified configurations.
func (cfg *Config) SetKubeClient() error {
	kubeConfig := os.Getenv("KUBECONFIG")

	cfg.logger.Debug("found kubeconfig", slog.Any("kubeConfig", kubeConfig))
	cfg.logger.Debug("using context", slog.Any("context", cfg.Context))

	config, err := buildConfigWithContextFromFlags(cfg.Context, kubeConfig)
	if err != nil {
		cfg.logger.Error("failed to load Kubernetes config", slog.Any("error", err))

		return err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		cfg.logger.Error("failed to create Kubernetes client", slog.Any("error", err))

		return err
	}

	cfg.clientSet = clientSet

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
	loadingRules := &clientcmd.ClientConfigLoadingRules{}
	loadingRules.Precedence = strings.Split(kubeConfigPath, string(os.PathListSeparator))

	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: kubeContext,
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	).ClientConfig()
}

// SetLogger sets logger to the Config.
func (cfg *Config) SetLogger(logger *slog.Logger) {
	cfg.logger = logger
}

// New returns new instance of Config when invoked.
func New() *Config {
	return &Config{}
}
