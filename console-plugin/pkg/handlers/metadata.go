package handlers

import (
	"fmt"
	"net/http"

	"github.com/openshift/vcf-migration-operator/internal/metadata"
)

// ServeMetadataDownload serves the installer metadata JSON from the Secret
// associated with a migration. Query parameters: namespace, name (migration name).
func (h *Handler) ServeMetadataDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeJSONError(w, http.StatusBadRequest, "namespace and name query parameters are required")
		return
	}

	secretName := metadata.GetMetadataSecretName(name)
	mgr := metadata.NewMetadataManager(h.KubeClient)

	data, err := mgr.GetMetadataFromSecret(r.Context(), namespace, secretName)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "metadata not found: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-metadata.json"`, name))
	w.WriteHeader(http.StatusOK)
	w.Write(data) //nolint:errcheck
}
