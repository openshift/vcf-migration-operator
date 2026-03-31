When a user does not define a folder path, the folder needs to be created based on the infra id
and then used within the defintion of the new machine set. See this example for the problem.

```
apiVersion: machine.openshift.io/v1beta1
kind: MachineSet
metadata:
  annotations:
    machine.openshift.io/memoryMb: '16384'
    machine.openshift.io/vCPU: '4'
  resourceVersion: '36196'
  name: jcallen2-ztbvp-worker-fd-wldn-120-dc
  uid: 5e2bde8d-a966-410c-98d6-74d70af5da58
  creationTimestamp: '2026-03-12T14:02:14Z'
  generation: 1
  namespace: openshift-machine-api
  labels:
    machine.openshift.io/cluster-api-cluster: jcallen2-ztbvp
spec:
  replicas: 3
  selector:
    matchLabels:
      machine.openshift.io/cluster-api-cluster: jcallen2-ztbvp
      machine.openshift.io/cluster-api-machineset: jcallen2-ztbvp-worker-fd-wldn-120-dc
  template:
    metadata:
      labels:
        machine.openshift.io/cluster-api-cluster: jcallen2-ztbvp
        machine.openshift.io/cluster-api-machine-role: worker
        machine.openshift.io/cluster-api-machine-type: worker
        machine.openshift.io/cluster-api-machineset: jcallen2-ztbvp-worker-fd-wldn-120-dc
    spec:
      lifecycleHooks: {}
      metadata: {}
      providerSpec:
        value:
          numCoresPerSocket: 4
          diskGiB: 120
          snapshot: ''
          userDataSecret:
            name: worker-user-data
          memoryMiB: 16384
          credentialsSecret:
            name: vsphere-cloud-credentials
          network:
            devices:
              - networkName: ci-vlan-826
          metadata: {}
          numCPUs: 4
          kind: VSphereMachineProviderSpec
          workspace:
            datacenter: wldn-120-DC
            datastore: /wldn-120-DC/datastore/wldn-120-cl01-vsan01
            resourcePool: /wldn-120-DC/host/wldn-120-cl01/Resources/ipi-ci-clusters
            server: vcenter-120.ci.ibmc.devcluster.openshift.com
          template: /wldn-120-DC/vm/rhcos-9.6.20251212-1-vmware.x86_64.ova
          apiVersion: machine.openshift.io/v1beta1


```
