package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"k8s.io/klog/v2"

	"github.com/openshift/vcf-migration-operator/console-plugin/pkg/handlers"
)

// Server is the HTTPS server for the console plugin.
type Server struct {
	port     string
	certFile string
	keyFile  string
	mux      *http.ServeMux
	handler  *handlers.Handler
}

// New creates a new Server that serves static assets and API routes.
func New(port, certFile, keyFile, staticDir string, h *handlers.Handler) (*Server, error) {
	mux := http.NewServeMux()

	mux.HandleFunc("/vsphere/connect", h.ServeVSphereConnect)
	mux.HandleFunc("/vsphere/datacenters", h.ServeVSphereDatacenters)
	mux.HandleFunc("/vsphere/clusters", h.ServeVSphereClusters)
	mux.HandleFunc("/vsphere/datastores", h.ServeVSphereDatastores)
	mux.HandleFunc("/vsphere/networks", h.ServeVSphereNetworks)
	mux.HandleFunc("/vsphere/resourcepools", h.ServeVSphereResourcePools)
	mux.HandleFunc("/vsphere/templates", h.ServeVSphereTemplates)
	mux.HandleFunc("/vsphere/folders", h.ServeVSphereFolders)
	mux.HandleFunc("/events", h.ServeEventsSSE)
	mux.HandleFunc("/metadata", h.ServeMetadataDownload)

	// Static assets: serve everything from staticDir at /
	fs := http.FileServer(http.Dir(staticDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Don't serve directory listing; return 404 for directories.
		path := filepath.Join(staticDir, r.URL.Path)
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			index := filepath.Join(path, "index.html")
			if _, err := os.Stat(index); os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			}
		}
		// Let API routes take precedence (already registered above).
		if strings.HasPrefix(r.URL.Path, "/vsphere/") || r.URL.Path == "/events" || r.URL.Path == "/metadata" {
			http.NotFound(w, r)
			return
		}
		fs.ServeHTTP(w, r)
	})

	return &Server{
		port:     port,
		certFile: certFile,
		keyFile:  keyFile,
		mux:      mux,
		handler:  h,
	}, nil
}

// recoveryMiddleware wraps an http.Handler to catch panics from downstream
// handlers (e.g. govmomi panicking on malformed URLs) and return a 500
// response instead of crashing the server process.
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				klog.Errorf("recovered from panic serving %s %s: %v\n%s", r.Method, r.URL.Path, rec, debug.Stack())
				http.Error(w, fmt.Sprintf(`{"error":"internal server error"}`), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ListenAndServe starts the HTTPS server.
func (s *Server) ListenAndServe() error {
	return http.ListenAndServeTLS(":"+s.port, s.certFile, s.keyFile, recoveryMiddleware(s.mux))
}
