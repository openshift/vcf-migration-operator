Task summary

create a new branch

Review https://raw.githubusercontent.com/openshift/installer/refs/heads/main/pkg/types/vsphere/metadata.go
The installer usually creates this file so that when the cluster is destroyed, 
it can connect to vcenter and delete objects. The problem being that
the cluster was built with the source configuration.

Create the replacement, include this file as a downloadable object in the 
details page. This should also be generated as a secret since it contains
authentication data. 

commit and push

Expectations:

- Build and run unit tests after EVERY code change
- Test driven development, use govmomi simulator for vSphere specific tests
- Expert level at golang
- Expert level at patternfly
- Will follow all best practices for golang, kubernetes and patternfly
- Document everything
- Re-use, clean, minimal, simple to understand code.
- Use openshift libraries and other repositories for code re-use as much as possible.
  https://github.com/openshift/library-go
  https://github.com/openshift/client-go
  https://github.com/openshift/installer
  https://github.com/openshift/machine-api-operator
  https://github.com/kubernetes-sigs/cluster-api-provider-vsphere
  https://github.com/openshift/cluster-control-plane-machine-set-operator
  https://github.com/openshift/vmware-vsphere-csi-driver
