/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/openshift/vcf-migration-operator/console-plugin/pkg/handlers"
	"github.com/openshift/vcf-migration-operator/console-plugin/pkg/server"
)

func main() {
	var port, certFile, keyFile, staticDir string
	flag.StringVar(&port, "port", "9443", "HTTPS port for the plugin server")
	flag.StringVar(&certFile, "tls-cert", "tls.crt", "Path to TLS certificate")
	flag.StringVar(&keyFile, "tls-key", "tls.key", "Path to TLS private key")
	flag.StringVar(&staticDir, "static-dir", "dist", "Directory containing plugin static assets")
	flag.Parse()

	restConfig, err := rest.InClusterConfig()
	if err != nil {
		restConfig, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		if err != nil {
			klog.ErrorS(err, "building kubeconfig")
			os.Exit(1)
		}
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		klog.ErrorS(err, "creating kubernetes client")
		os.Exit(1)
	}

	h := handlers.New(kubeClient)
	srv, err := server.New(port, certFile, keyFile, staticDir, h)
	if err != nil {
		klog.ErrorS(err, "creating server")
		os.Exit(1)
	}

	klog.InfoS("starting console plugin server", "port", port)
	if err := srv.ListenAndServe(); err != nil {
		klog.ErrorS(err, "server exited")
		os.Exit(1)
	}
}
