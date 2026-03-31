package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
)

// ServeEventsSSE streams Kubernetes Events for a VmwareCloudFoundationMigration
// as Server-Sent Events. Query params: namespace, name.
func (h *Handler) ServeEventsSSE(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		http.Error(w, "namespace and name query params required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	log := klog.FromContext(ctx)
	fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=VmwareCloudFoundationMigration", name)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	sendEvent := func(ev *corev1.Event) {
		payload := map[string]interface{}{
			"type":          ev.Type,
			"reason":        ev.Reason,
			"message":       ev.Message,
			"count":         ev.Count,
			"lastTimestamp": ev.LastTimestamp.Format(time.RFC3339),
			"involvedObject": map[string]string{
				"kind":      ev.InvolvedObject.Kind,
				"name":      ev.InvolvedObject.Name,
				"namespace": ev.InvolvedObject.Namespace,
			},
		}
		data, _ := json.Marshal(payload)
		_, _ = w.Write([]byte("data: " + string(data) + "\n\n"))
		flusher.Flush()
	}

	sendKeepAlive := func() {
		_, _ = w.Write([]byte(": keepalive\n\n"))
		flusher.Flush()
	}

	list, err := h.KubeClient.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{FieldSelector: fieldSelector})
	if err != nil {
		log.Error(err, "listing events")
		_, _ = w.Write([]byte("data: {}\n\n"))
		flusher.Flush()
	} else {
		for i := range list.Items {
			sendEvent(&list.Items[i])
		}
	}

	for {
		if ctx.Err() != nil {
			return
		}

		watcher, err := h.KubeClient.CoreV1().Events(namespace).Watch(ctx, metav1.ListOptions{FieldSelector: fieldSelector})
		if err != nil {
			log.Error(err, "watching events, retrying in 5s")
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}

		watchLoop(ctx, watcher, sendEvent, sendKeepAlive)
		watcher.Stop()

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}

func watchLoop(ctx interface{ Done() <-chan struct{} }, watcher watch.Interface, sendEvent func(*corev1.Event), sendKeepAlive func()) {
	keepAlive := time.NewTicker(30 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-keepAlive.C:
			sendKeepAlive()
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return
			}
			if event.Type == watch.Deleted {
				continue
			}
			ev, ok := event.Object.(*corev1.Event)
			if !ok {
				continue
			}
			sendEvent(ev)
		}
	}
}
