package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	"github.com/vmware/govmomi/object"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/openshift/vcf-migration-operator/internal/openshift"
	"github.com/openshift/vcf-migration-operator/internal/vsphere"
)

const (
	contentTypeJSON = "application/json"
)

// VSphereConnectRequest is the body for POST /vsphere/connect.
type VSphereConnectRequest struct {
	Server       string               `json:"server"`
	Username     string               `json:"username,omitempty"`
	Password     string               `json:"password,omitempty"`
	SecretRef    *SecretRef           `json:"secretRef,omitempty"`
	CreateSecret *CreateSecretOptions `json:"createSecret,omitempty"`
}

// SecretRef points to a Kubernetes secret with vCenter credentials.
type SecretRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// CreateSecretOptions instructs the backend to create a secret from inline credentials.
type CreateSecretOptions struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// VSphereConnectResponse is the response for connect.
type VSphereConnectResponse struct {
	Datacenters   []string   `json:"datacenters"`
	SecretCreated *SecretRef `json:"secretCreated,omitempty"`
	Error         string     `json:"error,omitempty"`
}

func (h *Handler) ServeVSphereConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req VSphereConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "decoding request: %v", err)
		return
	}

	ctx := r.Context()
	username := req.Username
	password := req.Password

	if req.SecretRef != nil && req.SecretRef.Name != "" {
		ns := req.SecretRef.Namespace
		if ns == "" {
			ns = "openshift-vcf-migration"
		}
		sm := openshift.NewSecretManager(h.KubeClient)
		var err error
		username, password, err = sm.GetVCenterCredsFromSecret(ctx, ns, req.SecretRef.Name, req.Server)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "getting credentials from secret: %v", err)
			return
		}
	}

	if username == "" || password == "" {
		writeJSONError(w, http.StatusBadRequest, "username and password required (inline or via secretRef)")
		return
	}

	p := vsphere.Params{
		Server:   req.Server,
		Username: username,
		Password: password,
		Insecure: true,
	}

	datacenters, err := vsphere.ListDatacenters(ctx, p)
	if err != nil {
		klog.FromContext(ctx).V(1).Info("vsphere list datacenters failed", "server", req.Server, "err", err)
		writeJSON(w, http.StatusOK, VSphereConnectResponse{Error: err.Error()})
		return
	}

	resp := VSphereConnectResponse{Datacenters: datacenters}

	if req.CreateSecret != nil && req.CreateSecret.Name != "" && req.Username != "" && req.Password != "" {
		ns := req.CreateSecret.Namespace
		if ns == "" {
			ns = "openshift-vcf-migration"
		}
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.CreateSecret.Name,
				Namespace: ns,
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				req.Server + ".username": []byte(req.Username),
				req.Server + ".password": []byte(req.Password),
			},
		}
		_, err := h.KubeClient.CoreV1().Secrets(ns).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "creating secret: %v", err)
			return
		}
		resp.SecretCreated = &SecretRef{Name: req.CreateSecret.Name, Namespace: ns}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) getSession(r *http.Request) (*vsphere.Session, string, string, int) {
	ctx := r.Context()
	server := r.URL.Query().Get("server")
	datacenter := r.URL.Query().Get("datacenter")
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")
	secretName := r.URL.Query().Get("secretName")
	secretNamespace := r.URL.Query().Get("secretNamespace")

	if server == "" || datacenter == "" {
		return nil, "", "", http.StatusBadRequest
	}

	if secretName != "" {
		ns := secretNamespace
		if ns == "" {
			ns = "openshift-vcf-migration"
		}
		sm := openshift.NewSecretManager(h.KubeClient)
		var err error
		username, password, err = sm.GetVCenterCredsFromSecret(ctx, ns, secretName, server)
		if err != nil {
			klog.FromContext(ctx).V(1).Info("get creds from secret failed", "err", err)
			return nil, "", "", http.StatusBadRequest
		}
	}

	if username == "" || password == "" {
		return nil, "", "", http.StatusBadRequest
	}

	p := vsphere.Params{
		Server:     server,
		Datacenter: datacenter,
		Username:   username,
		Password:   password,
		Insecure:   true,
	}
	session, err := vsphere.GetOrCreate(ctx, p)
	if err != nil {
		klog.FromContext(ctx).V(1).Info("vsphere session failed", "err", err)
		return nil, "", "", http.StatusInternalServerError
	}
	return session, server, datacenter, 0
}

func (h *Handler) ServeVSphereDatacenters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Datacenters are returned from connect; this endpoint can mirror connect for GET.
	server := r.URL.Query().Get("server")
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")
	secretName := r.URL.Query().Get("secretName")
	secretNamespace := r.URL.Query().Get("secretNamespace")
	if server == "" {
		writeJSONError(w, http.StatusBadRequest, "server query required")
		return
	}
	ctx := r.Context()
	if secretName != "" {
		ns := secretNamespace
		if ns == "" {
			ns = "openshift-vcf-migration"
		}
		sm := openshift.NewSecretManager(h.KubeClient)
		var err error
		username, password, err = sm.GetVCenterCredsFromSecret(ctx, ns, secretName, server)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "secret: %v", err)
			return
		}
	}
	if username == "" || password == "" {
		writeJSONError(w, http.StatusBadRequest, "username and password required")
		return
	}
	p := vsphere.Params{Server: server, Username: username, Password: password, Insecure: true}
	datacenters, err := vsphere.ListDatacenters(ctx, p)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "listing datacenters: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"datacenters": datacenters})
}

func (h *Handler) ServeVSphereClusters(w http.ResponseWriter, r *http.Request) {
	session, _, _, status := h.getSession(r)
	if status != 0 {
		writeJSONError(w, status, "invalid session params")
		return
	}
	ctx := r.Context()
	list, err := session.Finder.ClusterComputeResourceList(ctx, "*")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "listing clusters: %v", err)
		return
	}
	items := make([]string, 0, len(list))
	for _, c := range list {
		items = append(items, c.InventoryPath)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

func (h *Handler) ServeVSphereDatastores(w http.ResponseWriter, r *http.Request) {
	session, _, _, status := h.getSession(r)
	if status != 0 {
		writeJSONError(w, status, "invalid session params")
		return
	}
	ctx := r.Context()
	list, err := session.Finder.DatastoreList(ctx, "*")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "listing datastores: %v", err)
		return
	}
	items := make([]string, 0, len(list))
	for _, d := range list {
		items = append(items, d.InventoryPath)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

func (h *Handler) ServeVSphereNetworks(w http.ResponseWriter, r *http.Request) {
	session, _, _, status := h.getSession(r)
	if status != 0 {
		writeJSONError(w, status, "invalid session params")
		return
	}
	ctx := r.Context()
	list, err := session.Finder.NetworkList(ctx, "*")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "listing networks: %v", err)
		return
	}
	items := make([]string, 0, len(list))
	for _, n := range list {
		items = append(items, path.Base(n.GetInventoryPath()))
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

func (h *Handler) ServeVSphereResourcePools(w http.ResponseWriter, r *http.Request) {
	session, _, _, status := h.getSession(r)
	if status != 0 {
		writeJSONError(w, status, "invalid session params")
		return
	}
	ctx := r.Context()
	list, err := session.Finder.ResourcePoolList(ctx, "*")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "listing resource pools: %v", err)
		return
	}
	items := make([]string, 0, len(list))
	for _, rp := range list {
		items = append(items, rp.InventoryPath)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

func (h *Handler) ServeVSphereTemplates(w http.ResponseWriter, r *http.Request) {
	session, _, _, status := h.getSession(r)
	if status != 0 {
		writeJSONError(w, status, "invalid session params")
		return
	}
	ctx := r.Context()
	list, err := session.Finder.VirtualMachineList(ctx, "*")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "listing VMs: %v", err)
		return
	}
	items := make([]string, 0, len(list))
	for _, vm := range list {
		items = append(items, vm.InventoryPath)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

func (h *Handler) ServeVSphereFolders(w http.ResponseWriter, r *http.Request) {
	session, _, _, status := h.getSession(r)
	if status != 0 {
		writeJSONError(w, status, "invalid session params")
		return
	}
	ctx := r.Context()
	dc, err := session.Finder.Datacenter(ctx, session.Datacenter)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "finding datacenter: %v", err)
		return
	}
	folders, err := dc.Folders(ctx)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "getting folders: %v", err)
		return
	}
	vmFolder := folders.VmFolder
	if vmFolder == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"items": []string{}})
		return
	}

	var items []string
	if err := collectFolders(ctx, vmFolder, &items); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "listing VM folders: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

// collectFolders recursively walks a VM folder tree and appends each folder's
// inventory path to items.
func collectFolders(ctx context.Context, folder *object.Folder, items *[]string) error {
	*items = append(*items, folder.InventoryPath)
	children, err := folder.Children(ctx)
	if err != nil {
		return fmt.Errorf("listing children of %s: %w", folder.InventoryPath, err)
	}
	for _, child := range children {
		if subfolder, ok := child.(*object.Folder); ok {
			if err := collectFolders(ctx, subfolder, items); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if status == 0 {
		status = http.StatusInternalServerError
	}
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
