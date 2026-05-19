package main

import (
	"context"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"muyiwadosunmu/k8s-inspector/internal/pkg/web"
)

type PodInfo struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Node     string `json:"node,omitempty"`
	Requests struct {
		CPU    string `json:"cpu,omitempty"`
		Memory string `json:"memory,omitempty"`
	} `json:"requests,omitempty"`
}

// podsHandler returns pods for a given namespace specified by the 'namespace' query param.
func (app *application) podsHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	qs := r.URL.Query()
	ns := qs.Get("namespace")
	if ns == "" {
		return web.NewRequestError(http.StatusBadRequest, "missing 'namespace' query parameter")
	}

	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client, ok := app.k8sClient.(*kubernetes.Clientset)
	if !ok || client == nil {
		return web.NewRequestError(http.StatusInternalServerError, "kubernetes client not initialized")
	}

	// Verify namespace exists
	if _, err := client.CoreV1().Namespaces().Get(cctx, ns, metav1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			return web.NewRequestError(http.StatusNotFound, "namespace not found")
		}
		return err
	}

	pods, err := client.CoreV1().Pods(ns).List(cctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	var out []PodInfo
	for _, p := range pods.Items {
		pi := PodInfo{
			Name:   p.Name,
			Status: string(p.Status.Phase),
			Node:   p.Spec.NodeName,
		}

		// Sum requests across containers
		totalCPU := resource.Quantity{}
		totalMem := resource.Quantity{}
		hasCPU := false
		hasMem := false
		for _, c := range p.Spec.Containers {
			if q, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
				totalCPU.Add(q)
				hasCPU = true
			}
			if q, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
				totalMem.Add(q)
				hasMem = true
			}
		}
		if hasCPU {
			pi.Requests.CPU = totalCPU.String()
		}
		if hasMem {
			pi.Requests.Memory = totalMem.String()
		}

		out = append(out, pi)
	}

	if len(out) == 0 {
		// Return empty list with 200; client can interpret empty result
		return web.Respond(ctx, w, []PodInfo{}, http.StatusOK)
	}

	return web.Respond(ctx, w, out, http.StatusOK)
}
