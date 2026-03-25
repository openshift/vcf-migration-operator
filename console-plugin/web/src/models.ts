import { K8sModel } from '@openshift-console/dynamic-plugin-sdk';

export const VmwareCloudFoundationMigrationModel: K8sModel = {
  kind: 'VmwareCloudFoundationMigration',
  label: 'VmwareCloudFoundationMigration',
  labelPlural: 'VmwareCloudFoundationMigrations',
  apiGroup: 'migration.openshift.io',
  apiVersion: 'v1alpha1',
  plural: 'vmwarecloudfoundationmigrations',
  abbr: 'vcfm',
  namespaced: true,
  crd: true,
};

export const MachineSetModel: K8sModel = {
  kind: 'MachineSet',
  label: 'MachineSet',
  labelPlural: 'MachineSets',
  apiGroup: 'machine.openshift.io',
  apiVersion: 'v1beta1',
  plural: 'machinesets',
  abbr: 'ms',
  namespaced: true,
};

export const MachineModel: K8sModel = {
  kind: 'Machine',
  label: 'Machine',
  labelPlural: 'Machines',
  apiGroup: 'machine.openshift.io',
  apiVersion: 'v1beta1',
  plural: 'machines',
  abbr: 'm',
  namespaced: true,
};

export const ControlPlaneMachineSetModel: K8sModel = {
  kind: 'ControlPlaneMachineSet',
  label: 'ControlPlaneMachineSet',
  labelPlural: 'ControlPlaneMachineSets',
  apiGroup: 'machine.openshift.io',
  apiVersion: 'v1',
  plural: 'controlplanemachinesets',
  abbr: 'cpms',
  namespaced: true,
};

export const NodeModel: K8sModel = {
  kind: 'Node',
  label: 'Node',
  labelPlural: 'Nodes',
  apiGroup: '',
  apiVersion: 'v1',
  plural: 'nodes',
  abbr: 'n',
  namespaced: false,
};

export interface SecretReference {
  name: string;
  namespace?: string;
}

export interface VSpherePlatformTopology {
  datacenter: string;
  computeCluster: string;
  datastore: string;
  networks?: string[];
  resourcePool?: string;
  template: string;
  folder?: string;
}

export interface VSpherePlatformFailureDomainSpec {
  name: string;
  region: string;
  zone: string;
  server: string;
  topology: VSpherePlatformTopology;
}

export interface VmwareCloudFoundationMigrationSpec {
  state: 'Pending' | 'Running' | 'Paused';
  targetVCenterCredentialsSecret: SecretReference;
  failureDomains: VSpherePlatformFailureDomainSpec[];
}

export interface VmwareCloudFoundationMigrationStatus {
  conditions?: Array<{
    type: string;
    status: string;
    reason?: string;
    message?: string;
    lastTransitionTime?: string;
  }>;
  startTime?: string;
  completionTime?: string;
}

export interface VmwareCloudFoundationMigrationKind {
  apiVersion: string;
  kind: string;
  metadata: {
    name: string;
    namespace: string;
    [key: string]: unknown;
  };
  spec: VmwareCloudFoundationMigrationSpec;
  status?: VmwareCloudFoundationMigrationStatus;
}
