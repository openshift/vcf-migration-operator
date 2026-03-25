import * as React from 'react';
import {
  Label,
  Spinner,
  Bullseye,
  EmptyState,
  EmptyStateHeader,
  EmptyStateIcon,
  EmptyStateBody,
  Title,
  PageSection,
  DescriptionList,
  DescriptionListGroup,
  DescriptionListTerm,
  DescriptionListDescription,
} from '@patternfly/react-core';
import {
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
} from '@patternfly/react-table';
import { CubesIcon } from '@patternfly/react-icons';
import { useK8sWatchResource, ResourceLink } from '@openshift-console/dynamic-plugin-sdk';
import {
  MachineSetModel,
  MachineModel,
  NodeModel,
  ControlPlaneMachineSetModel,
} from '../../models';
import { MachineTopologyGraph } from './MachineTopologyGraph';

/* ------------------------------------------------------------------ */
/* Types                                                               */
/* ------------------------------------------------------------------ */

/** Machine providerSpec.value.workspace — only the server field is needed. */
interface VSphereMachineProviderSpec {
  workspace?: { server?: string };
}

interface MachineSetKind {
  metadata: { name: string; namespace: string };
  spec: { replicas?: number };
  status?: { readyReplicas?: number; availableReplicas?: number };
}

interface MachineKind {
  metadata: { name: string; namespace: string; labels?: Record<string, string> };
  spec?: { providerSpec?: { value?: VSphereMachineProviderSpec } };
  status?: { phase?: string; nodeRef?: { name: string } };
}

interface NodeKind {
  metadata: { name: string };
  status?: {
    conditions?: Array<{ type: string; status: string }>;
  };
}

interface CPMSKind {
  metadata: { name: string; namespace: string };
  spec: {
    state?: string;
    replicas?: number;
  };
  status?: {
    replicas?: number;
    updatedReplicas?: number;
    readyReplicas?: number;
  };
}

const machineAPINamespace = 'openshift-machine-api';

/* ------------------------------------------------------------------ */
/* Shared row type used by both tables and the topology graph          */
/* ------------------------------------------------------------------ */

export interface MachineRow {
  machineSetName: string | null;
  machineName: string;
  machineNamespace: string;
  machinePhase: string;
  nodeName: string | null;
  nodeReady: boolean | null;
  role: 'master' | 'worker';
  vcenter: string;
}

/* ------------------------------------------------------------------ */
/* Props                                                               */
/* ------------------------------------------------------------------ */

export interface MachineTopologyViewProps {
  namespace: string;
}

/* ------------------------------------------------------------------ */
/* Component                                                           */
/* ------------------------------------------------------------------ */

export const MachineTopologyView: React.FC<MachineTopologyViewProps> = () => {
  /* K8s watches */
  const [machineSets, msLoaded] = useK8sWatchResource<MachineSetKind[]>({
    groupVersionKind: {
      group: MachineSetModel.apiGroup,
      version: MachineSetModel.apiVersion,
      kind: MachineSetModel.kind,
    },
    namespace: machineAPINamespace,
    isList: true,
    namespaced: true,
  });
  const [machines, mLoaded] = useK8sWatchResource<MachineKind[]>({
    groupVersionKind: {
      group: MachineModel.apiGroup,
      version: MachineModel.apiVersion,
      kind: MachineModel.kind,
    },
    namespace: machineAPINamespace,
    isList: true,
    namespaced: true,
  });
  const [nodes, nLoaded] = useK8sWatchResource<NodeKind[]>({
    groupVersionKind: {
      group: NodeModel.apiGroup,
      version: NodeModel.apiVersion,
      kind: NodeModel.kind,
    },
    isList: true,
    namespaced: false,
  });
  const [cpmsList, cpmsLoaded] = useK8sWatchResource<CPMSKind[]>({
    groupVersionKind: {
      group: ControlPlaneMachineSetModel.apiGroup,
      version: ControlPlaneMachineSetModel.apiVersion,
      kind: ControlPlaneMachineSetModel.kind,
    },
    namespace: machineAPINamespace,
    isList: true,
    namespaced: true,
  });

  const loaded = msLoaded && mLoaded && nLoaded && cpmsLoaded;

  /* Build lookup maps */
  const nodeMap = React.useMemo(() => {
    const m: Record<string, NodeKind> = {};
    nodes?.forEach((n) => { m[n.metadata.name] = n; });
    return m;
  }, [nodes]);

  const msNames = React.useMemo(
    () => new Set(machineSets?.map((ms) => ms.metadata.name) ?? []),
    [machineSets],
  );

  /* Build rows */
  const rows: MachineRow[] = React.useMemo(() => {
    if (!loaded) return [];
    const result: MachineRow[] = [];

    machines?.forEach((m) => {
      const msLabel = m.metadata.labels?.['machine.openshift.io/cluster-api-machineset'] ?? null;
      const roleLabel = m.metadata.labels?.['machine.openshift.io/cluster-api-machine-role'] ?? '';
      const nodeName = m.status?.nodeRef?.name ?? null;
      let nodeReady: boolean | null = null;
      if (nodeName) {
        const node = nodeMap[nodeName];
        const cond = node?.status?.conditions?.find((c) => c.type === 'Ready');
        nodeReady = cond?.status === 'True';
      }
      const role: 'master' | 'worker' = roleLabel === 'master' ? 'master' : 'worker';
      const vcenter = m.spec?.providerSpec?.value?.workspace?.server ?? 'unknown';

      result.push({
        machineSetName: msLabel && msNames.has(msLabel) ? msLabel : null,
        machineName: m.metadata.name,
        machineNamespace: m.metadata.namespace,
        machinePhase: m.status?.phase ?? 'Unknown',
        nodeName,
        nodeReady,
        role,
        vcenter,
      });
    });

    result.sort((a, b) => a.machineName.localeCompare(b.machineName));
    return result;
  }, [loaded, machines, msNames, nodeMap]);

  const controlPlaneRows = React.useMemo(() => rows.filter((r) => r.role === 'master'), [rows]);
  const workerRows = React.useMemo(() => rows.filter((r) => r.role === 'worker'), [rows]);
  const cpms = cpmsList?.[0] ?? null;

  /* Loading */
  if (!loaded) {
    return (
      <Bullseye>
        <Spinner size="lg" aria-label="Loading machine topology" />
      </Bullseye>
    );
  }

  /* Empty */
  if (!machines?.length) {
    return (
      <EmptyState>
        <EmptyStateHeader
          titleText="No machines found"
          headingLevel="h4"
          icon={<EmptyStateIcon icon={CubesIcon} />}
        />
        <EmptyStateBody>
          No machines were found in the openshift-machine-api namespace.
        </EmptyStateBody>
      </EmptyState>
    );
  }

  return (
    <>
      {/* Control Plane Section */}
      <PageSection variant="light" className="pf-v5-u-pb-lg">
        <Title headingLevel="h2" className="pf-v5-u-mb-md">Control Plane</Title>
        {cpms && <CPMSSummary cpms={cpms} />}
        <MachineTable rows={controlPlaneRows} />
      </PageSection>

      {/* Compute Section */}
      <PageSection variant="light" className="pf-v5-u-pb-lg">
        <Title headingLevel="h2" className="pf-v5-u-mb-md">Compute</Title>
        <MachineTable rows={workerRows} />
      </PageSection>

      {/* Topology Graph */}
      <PageSection variant="light">
        <Title headingLevel="h2" className="pf-v5-u-mb-md">Topology</Title>
        <MachineTopologyGraph rows={rows} />
      </PageSection>
    </>
  );
};

/* ------------------------------------------------------------------ */
/* CPMS summary                                                        */
/* ------------------------------------------------------------------ */

const CPMSSummary: React.FC<{ cpms: CPMSKind }> = ({ cpms }) => (
  <DescriptionList isHorizontal isCompact className="pf-v5-u-mb-md">
    <DescriptionListGroup>
      <DescriptionListTerm>ControlPlaneMachineSet</DescriptionListTerm>
      <DescriptionListDescription>
        <ResourceLink
          groupVersionKind={{
            group: ControlPlaneMachineSetModel.apiGroup,
            version: ControlPlaneMachineSetModel.apiVersion,
            kind: ControlPlaneMachineSetModel.kind,
          }}
          name={cpms.metadata.name}
          namespace={cpms.metadata.namespace}
        />
      </DescriptionListDescription>
    </DescriptionListGroup>
    <DescriptionListGroup>
      <DescriptionListTerm>State</DescriptionListTerm>
      <DescriptionListDescription>
        <Label color={cpms.spec.state === 'Active' ? 'green' : 'grey'} isCompact>
          {cpms.spec.state ?? 'Unknown'}
        </Label>
      </DescriptionListDescription>
    </DescriptionListGroup>
    <DescriptionListGroup>
      <DescriptionListTerm>Replicas</DescriptionListTerm>
      <DescriptionListDescription>
        {cpms.status?.readyReplicas ?? 0} / {cpms.spec.replicas ?? 0} ready
      </DescriptionListDescription>
    </DescriptionListGroup>
  </DescriptionList>
);

/* ------------------------------------------------------------------ */
/* Shared machine table                                                */
/* ------------------------------------------------------------------ */

const MachineTable: React.FC<{ rows: MachineRow[] }> = ({ rows }) => {
  if (rows.length === 0) {
    return <EmptyStateBody>No machines in this category.</EmptyStateBody>;
  }

  return (
    <Table aria-label="Machines" variant="compact">
      <Thead>
        <Tr>
          <Th>MachineSet</Th>
          <Th>Machine</Th>
          <Th>Phase</Th>
          <Th>Node</Th>
          <Th>Status</Th>
          <Th>vCenter</Th>
        </Tr>
      </Thead>
      <Tbody>
        {rows.map((row) => (
          <Tr key={row.machineName}>
            <Td dataLabel="MachineSet">
              {row.machineSetName ? (
                <ResourceLink
                  groupVersionKind={{
                    group: MachineSetModel.apiGroup,
                    version: MachineSetModel.apiVersion,
                    kind: MachineSetModel.kind,
                  }}
                  name={row.machineSetName}
                  namespace={machineAPINamespace}
                />
              ) : (
                <span className="pf-v5-u-color-200">-</span>
              )}
            </Td>
            <Td dataLabel="Machine">
              <ResourceLink
                groupVersionKind={{
                  group: MachineModel.apiGroup,
                  version: MachineModel.apiVersion,
                  kind: MachineModel.kind,
                }}
                name={row.machineName}
                namespace={row.machineNamespace}
              />
            </Td>
            <Td dataLabel="Phase">
              <Label color={getPhaseColor(row.machinePhase)} isCompact>
                {row.machinePhase}
              </Label>
            </Td>
            <Td dataLabel="Node">
              {row.nodeName ? (
                <ResourceLink
                  groupVersionKind={{
                    group: NodeModel.apiGroup,
                    version: NodeModel.apiVersion,
                    kind: NodeModel.kind,
                  }}
                  name={row.nodeName}
                />
              ) : (
                <span className="pf-v5-u-color-200">-</span>
              )}
            </Td>
            <Td dataLabel="Status">
              {row.nodeReady === null ? (
                <Label color="grey" isCompact>Pending</Label>
              ) : row.nodeReady ? (
                <Label color="green" isCompact>Ready</Label>
              ) : (
                <Label color="red" isCompact>Not Ready</Label>
              )}
            </Td>
            <Td dataLabel="vCenter">
              {row.vcenter !== 'unknown' ? row.vcenter : (
                <span className="pf-v5-u-color-200">-</span>
              )}
            </Td>
          </Tr>
        ))}
      </Tbody>
    </Table>
  );
};

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

const getPhaseColor = (phase: string): 'blue' | 'green' | 'red' | 'grey' => {
  switch (phase) {
    case 'Running':
      return 'green';
    case 'Provisioning':
    case 'Provisioned':
      return 'blue';
    case 'Failed':
    case 'Deleting':
      return 'red';
    default:
      return 'grey';
  }
};
