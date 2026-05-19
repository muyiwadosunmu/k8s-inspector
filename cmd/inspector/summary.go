package main

import (
	"context"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"muyiwadosunmu/k8s-inspector/internal/pkg/web"
)

// summaryHandler returns a cluster summary: number of namespaces, running pods, and nodes.
func (app *application) summaryHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Use a short timeout for K8s API calls
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client, ok := app.k8sClient.(*kubernetes.Clientset)
	if !ok || client == nil {
		return web.NewRequestError(http.StatusInternalServerError, "kubernetes client not initialized")
	}

	nsList, err := client.CoreV1().Namespaces().List(cctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	nodesList, err := client.CoreV1().Nodes().List(cctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	podsList, err := client.CoreV1().Pods("").List(cctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	running := 0
	for _, p := range podsList.Items {
		if p.Status.Phase == corev1.PodRunning {
			running++
		}
	}

	data := map[string]int{
		"namespaces":   len(nsList.Items),
		"running_pods": running,
		"nodes":        len(nodesList.Items),
	}

	return web.Respond(ctx, w, data, http.StatusOK)
}
