package handlers

import (
	"k8s.io/client-go/kubernetes"
)

// Handler holds dependencies for the console plugin HTTP handlers.
type Handler struct {
	KubeClient kubernetes.Interface
}

// New returns a new Handler.
func New(kubeClient kubernetes.Interface) *Handler {
	return &Handler{KubeClient: kubeClient}
}
