package main

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type config struct {
	port int
	env  string
}

type application struct {
	Config    config
	logger    *zap.Logger
	k8sClient *kubernetes.Clientset
	wg        sync.WaitGroup
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	k8sClient, err := openK8sClient()
	if err != nil {
		logger.Fatal("Failed to open K8s client", zap.Error(err))
	}

	app := &application{
		Config: config{
			port: 3000,
			env:  "development",
		},
		logger:    logger,
		k8sClient: k8sClient,
	}

	err = app.serve()
	if err != nil {
		logger.Fatal("server error", zap.Error(err))
	}
}

func openK8sClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		var kubeconfig string
		home := homedir.HomeDir()
		if home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		} else {
			kubeconfig = os.Getenv("KUBECONFIG")
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
