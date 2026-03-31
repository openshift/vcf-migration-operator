import * as React from 'react';
import {
  Button,
  Card,
  CardBody,
  CardHeader,
  CardTitle,
  Checkbox,
  Form,
  FormGroup,
  FormHelperText,
  HelperText,
  HelperTextItem,
  TextInput,
} from '@patternfly/react-core';
import { TrashIcon } from '@patternfly/react-icons';
import type { VSpherePlatformFailureDomainSpec } from '../../../models';
import { useVSphereList } from '../../hooks/useVSphereBrowse';
import { TypeaheadSelect } from './TypeaheadSelect';

export interface FailureDomainEditorProps {
  index: number;
  domain: VSpherePlatformFailureDomainSpec;
  onUpdate: (fd: VSpherePlatformFailureDomainSpec) => void;
  onRemove: () => void;
  datacenters: string[];
  server: string;
  username?: string;
  password?: string;
  secretRef?: { name: string; namespace?: string };
}

const sanitize = (s: string): string =>
  s.replace(/[^a-zA-Z0-9-]/g, '-').replace(/--+/g, '-').replace(/^-|-$/g, '').toLowerCase();

const lastSegment = (path: string): string => {
  const parts = path.split('/').filter(Boolean);
  return parts[parts.length - 1] || path;
};

export const FailureDomainEditor: React.FC<FailureDomainEditorProps> = ({
  index,
  domain,
  onUpdate,
  onRemove,
  datacenters,
  server,
  username,
  password,
  secretRef,
}) => {
  const prevDerivedName = React.useRef('');
  const prevDerivedRegion = React.useRef('');
  const prevDerivedZone = React.useRef('');
  const [createFolder, setCreateFolder] = React.useState(!domain.topology.folder);

  const listParams = React.useMemo(() => ({
    server,
    datacenter: domain.topology.datacenter,
    secretName: secretRef?.name,
    secretNamespace: secretRef?.namespace,
    username,
    password,
  }), [server, domain.topology.datacenter, secretRef?.name, secretRef?.namespace, username, password]);

  const { items: clusters, loading: clustersLoading, fetchList: fetchClusters } = useVSphereList('clusters', listParams);
  const { items: datastores, loading: dsLoading, fetchList: fetchDatastores } = useVSphereList('datastores', listParams);
  const { items: networks, loading: netLoading, fetchList: fetchNetworks } = useVSphereList('networks', listParams);
  const { items: resourcePools, loading: rpLoading, fetchList: fetchResourcePools } = useVSphereList('resourcepools', listParams);
  const { items: templates, loading: tmplLoading, fetchList: fetchTemplates } = useVSphereList('templates', listParams);
  const { items: folders, loading: folderLoading, fetchList: fetchFolders } = useVSphereList('folders', listParams);

  React.useEffect(() => {
    if (domain.topology.datacenter) {
      fetchClusters();
      fetchDatastores();
      fetchNetworks();
      fetchResourcePools();
      fetchTemplates();
      fetchFolders();
    }
  }, [domain.topology.datacenter, fetchClusters, fetchDatastores, fetchNetworks, fetchResourcePools, fetchTemplates, fetchFolders]);

  const handleDatacenterChange = React.useCallback((dc: string) => {
    const dcName = lastSegment(dc);
    const derivedRegion = sanitize(dcName);
    const derivedName = `fd-${derivedRegion}`;

    const updates: Partial<VSpherePlatformFailureDomainSpec> = {
      topology: { ...domain.topology, datacenter: dc, computeCluster: '', datastore: '', networks: [], template: '', folder: undefined, resourcePool: undefined },
    };

    if (!domain.name || domain.name === prevDerivedName.current) {
      updates.name = derivedName;
    }
    if (!domain.region || domain.region === prevDerivedRegion.current) {
      updates.region = derivedRegion;
    }
    if (!domain.zone || domain.zone === prevDerivedZone.current) {
      updates.zone = '';
    }

    prevDerivedName.current = derivedName;
    prevDerivedRegion.current = derivedRegion;

    onUpdate({ ...domain, ...updates });
  }, [domain, onUpdate]);

  const handleClusterChange = React.useCallback((cluster: string) => {
    const clusterName = lastSegment(cluster);
    const derivedZone = sanitize(clusterName);

    const updates: Partial<VSpherePlatformFailureDomainSpec> = {
      topology: { ...domain.topology, computeCluster: cluster },
    };

    if (!domain.zone || domain.zone === prevDerivedZone.current) {
      updates.zone = derivedZone;
    }

    prevDerivedZone.current = derivedZone;

    onUpdate({ ...domain, ...updates });
  }, [domain, onUpdate]);

  const updateTopology = React.useCallback(
    (field: string, value: string | string[] | undefined) => {
      onUpdate({
        ...domain,
        topology: { ...domain.topology, [field]: value },
      });
    },
    [domain, onUpdate],
  );

  const handleCreateFolderChange = React.useCallback(
    (_e: React.FormEvent, checked: boolean) => {
      setCreateFolder(checked);
      if (checked) {
        updateTopology('folder', undefined);
      }
    },
    [updateTopology],
  );

  const dcNotSelected = !domain.topology.datacenter;
  const idPrefix = `fd-${index}`;

  return (
    <Card isCompact>
      <CardHeader
        actions={{
          actions: (
            <Button
              variant="plain"
              icon={<TrashIcon />}
              onClick={onRemove}
              aria-label="Remove failure domain"
            />
          ),
        }}
      >
        <CardTitle>
          Failure domain {index + 1}
          {domain.name ? `: ${domain.name}` : ''}
        </CardTitle>
      </CardHeader>
      <CardBody>
        <Form>
          <FormGroup label="Datacenter" isRequired fieldId={`${idPrefix}-dc`}>
            <TypeaheadSelect
              id={`${idPrefix}-dc`}
              items={datacenters}
              value={domain.topology.datacenter}
              onChange={handleDatacenterChange}
              placeholder="Select datacenter"
              aria-label="Datacenter"
            />
          </FormGroup>

          <FormGroup label="Compute cluster" isRequired fieldId={`${idPrefix}-cluster`}>
            <TypeaheadSelect
              id={`${idPrefix}-cluster`}
              items={clusters}
              value={domain.topology.computeCluster}
              onChange={handleClusterChange}
              placeholder="Select cluster"
              isDisabled={dcNotSelected}
              isLoading={clustersLoading}
              aria-label="Compute cluster"
            />
          </FormGroup>

          <FormGroup label="Datastore" isRequired fieldId={`${idPrefix}-ds`}>
            <TypeaheadSelect
              id={`${idPrefix}-ds`}
              items={datastores}
              value={domain.topology.datastore}
              onChange={(v) => updateTopology('datastore', v)}
              placeholder="Select datastore"
              isDisabled={dcNotSelected}
              isLoading={dsLoading}
              aria-label="Datastore"
            />
          </FormGroup>

          <FormGroup label="Network" isRequired fieldId={`${idPrefix}-net`}>
            <TypeaheadSelect
              id={`${idPrefix}-net`}
              items={networks}
              value={domain.topology.networks?.[0] ?? ''}
              onChange={(v) => updateTopology('networks', v ? [v] : [])}
              placeholder="Select network"
              isDisabled={dcNotSelected}
              isLoading={netLoading}
              aria-label="Network"
            />
          </FormGroup>

          <FormGroup label="Template (RHCOS)" isRequired fieldId={`${idPrefix}-tmpl`}>
            <TypeaheadSelect
              id={`${idPrefix}-tmpl`}
              items={templates}
              value={domain.topology.template}
              onChange={(v) => updateTopology('template', v)}
              placeholder="Select template"
              isDisabled={dcNotSelected}
              isLoading={tmplLoading}
              aria-label="Template"
            />
          </FormGroup>

          <FormGroup label="Resource pool" fieldId={`${idPrefix}-rp`}>
            <TypeaheadSelect
              id={`${idPrefix}-rp`}
              items={resourcePools}
              value={domain.topology.resourcePool ?? ''}
              onChange={(v) => updateTopology('resourcePool', v || undefined)}
              placeholder="Select resource pool (optional)"
              isDisabled={dcNotSelected}
              isLoading={rpLoading}
              aria-label="Resource pool"
            />
          </FormGroup>

          <FormGroup label="Folder" fieldId={`${idPrefix}-folder`}>
            <Checkbox
              id={`${idPrefix}-create-folder`}
              label="Create folder automatically"
              isChecked={createFolder}
              onChange={handleCreateFolderChange}
              className="pf-v5-u-mb-sm"
            />
            {!createFolder && (
              <TypeaheadSelect
                id={`${idPrefix}-folder`}
                items={folders}
                value={domain.topology.folder ?? ''}
                onChange={(v) => updateTopology('folder', v || undefined)}
                placeholder="Select existing folder"
                isDisabled={dcNotSelected}
                isLoading={folderLoading}
                aria-label="Folder"
              />
            )}
            <FormHelperText>
              <HelperText>
                <HelperTextItem>
                  {createFolder
                    ? 'The operator will create a VM folder using the cluster infrastructure ID'
                    : 'Select an existing VM folder on the target vCenter'}
                </HelperTextItem>
              </HelperText>
            </FormHelperText>
          </FormGroup>

          <FormGroup label="Name" isRequired fieldId={`${idPrefix}-name`}>
            <TextInput
              id={`${idPrefix}-name`}
              value={domain.name}
              onChange={(_e, v) => onUpdate({ ...domain, name: v })}
              placeholder="fd-1"
            />
            <FormHelperText>
              <HelperText>
                <HelperTextItem>Auto-derived from datacenter; editable</HelperTextItem>
              </HelperText>
            </FormHelperText>
          </FormGroup>

          <FormGroup label="Region" isRequired fieldId={`${idPrefix}-region`}>
            <TextInput
              id={`${idPrefix}-region`}
              value={domain.region}
              onChange={(_e, v) => onUpdate({ ...domain, region: v })}
              placeholder="region1"
            />
            <FormHelperText>
              <HelperText>
                <HelperTextItem>Auto-derived from datacenter; editable</HelperTextItem>
              </HelperText>
            </FormHelperText>
          </FormGroup>

          <FormGroup label="Zone" isRequired fieldId={`${idPrefix}-zone`}>
            <TextInput
              id={`${idPrefix}-zone`}
              value={domain.zone}
              onChange={(_e, v) => onUpdate({ ...domain, zone: v })}
              placeholder="zone1"
            />
            <FormHelperText>
              <HelperText>
                <HelperTextItem>Auto-derived from compute cluster; editable</HelperTextItem>
              </HelperText>
            </FormHelperText>
          </FormGroup>
        </Form>
      </CardBody>
    </Card>
  );
};
