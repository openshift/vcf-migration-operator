
Task summary

Control plane machine sets are still not being configured properly. 
Review the manifest below.
The failureDomains are incorrect
The state is set to inactive
Find the cause and resolve this issue
Create a new branch, create a commit message and push

Expectations:

- Build and run unit tests after EVERY code change
- Test driven development, use govmomi simulator for vSphere specific tests
- Expert level at golang
- Will run golint, so best practices are followed
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


```

apiVersion: machine.openshift.io/v1
kind: ControlPlaneMachineSet
metadata:
  creationTimestamp: '2026-03-12T14:07:45Z'
  finalizers:
    - controlplanemachineset.machine.openshift.io
  generation: 1
  name: cluster
  namespace: openshift-machine-api
  resourceVersion: '39420'
  uid: e92bb02c-9876-4cb6-b7b2-b8050a621134
spec:
  replicas: 3
  selector:
    matchLabels:
      machine.openshift.io/cluster-api-cluster: jcallen2-ztbvp
      machine.openshift.io/cluster-api-machine-role: master
      machine.openshift.io/cluster-api-machine-type: master
  state: Inactive
  strategy:
    type: RollingUpdate
  template:
    machineType: machines_v1beta1_machine_openshift_io
    machines_v1beta1_machine_openshift_io:
      failureDomains:
        platform: VSphere
        vsphere:
          - name: us-east-1
      metadata:
        labels:
          machine.openshift.io/cluster-api-cluster: jcallen2-ztbvp
          machine.openshift.io/cluster-api-machine-role: master
          machine.openshift.io/cluster-api-machine-type: master
      spec:
        lifecycleHooks: {}
        metadata: {}
        providerSpec:
          value:
            numCoresPerSocket: 4
            diskGiB: 120
            snapshot: ''
            userDataSecret:
              name: master-user-data
            memoryMiB: 16384
            credentialsSecret:
              name: vsphere-cloud-credentials
            network:
              devices: null
            metadata: {}
            numCPUs: 4
            kind: VSphereMachineProviderSpec
            workspace: {}
            template: ''
            apiVersion: machine.openshift.io/v1beta1
status:
  conditions:
    - lastTransitionTime: '2026-03-12T14:07:45Z'
      message: ''
      observedGeneration: 1
      reason: AsExpected
      status: 'False'
      type: Error
    - lastTransitionTime: '2026-03-12T14:07:46Z'
      message: ''
      observedGeneration: 1
      reason: AllReplicasAvailable
      status: 'True'
      type: Available
    - lastTransitionTime: '2026-03-12T14:07:46Z'
      message: ''
      observedGeneration: 1
      reason: AsExpected
      status: 'False'
      type: Degraded
    - lastTransitionTime: '2026-03-12T14:07:46Z'
      message: ''
      observedGeneration: 1
      reason: AllReplicasUpdated
      status: 'False'
      type: Progressing
  observedGeneration: 1
  readyReplicas: 3
  replicas: 3
  updatedReplicas: 3


```