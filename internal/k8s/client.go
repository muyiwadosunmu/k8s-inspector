package k8s

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// NewClient creates a new Kubernetes client using in-cluster config with kubeconfig fallback.
func NewClient(kubeconfig string) (*kubernetes.Clientset, error) {
	// Try to use in-cluster config first
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig file
		if kubeconfig == "" {
			// If no kubeconfig provided, look for default locations
			home := homedir.HomeDir()
			if home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
			} else {
				kubeconfig = os.Getenv("KUBECONFIG")
			}
		}

		// Build config from kubeconfig file
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
