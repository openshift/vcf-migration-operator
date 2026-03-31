Task summary

On the machines tab separate control plane (masters) and compute (workers). 
In the control plane section make sure to include the control plane machine set

Below the machine tables create a patternfly topology.
There should be minimally two groups, one for each vcenter 
It should show the current location of the virtual machines
It should be color coded based on machine type (control plane or worker)
It should be a status for state (see https://github.com/patternfly/react-topology/blob/main/packages/module/patternfly-docs/content/examples/TopologyCustomNodesDemo.tsx)
- provisioning, not ready, ready




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
