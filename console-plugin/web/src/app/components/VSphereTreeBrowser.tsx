import * as React from 'react';
import {
  FormGroup,
  FormSelect,
  FormSelectOption,
  Button,
  Spinner,
  Alert,
} from '@patternfly/react-core';
import { useVSphereConnect, useVSphereList } from '../hooks/useVSphereBrowse';

export interface VSphereBrowserProps {
  server: string;
  username?: string;
  password?: string;
  secretRef?: { name: string; namespace?: string };
  datacenter: string;
  onDatacenterChange: (dc: string) => void;
  cluster: string;
  onClusterChange: (c: string) => void;
  datastore: string;
  onDatastoreChange: (d: string) => void;
  network: string;
  onNetworkChange: (n: string) => void;
  resourcePool: string;
  onResourcePoolChange: (rp: string) => void;
  template: string;
  onTemplateChange: (t: string) => void;
  folder: string;
  onFolderChange: (f: string) => void;
  onConnectSuccess?: (datacenters: string[]) => void;
}

export const VSphereTreeBrowser: React.FC<VSphereBrowserProps> = (props) => {
  const { connect, loading: connectLoading, error: connectError } = useVSphereConnect();
  const [datacenters, setDatacenters] = React.useState<string[]>([]);
  const [connected, setConnected] = React.useState(false);

  const clusterParams = {
    server: props.server,
    datacenter: props.datacenter,
    secretName: props.secretRef?.name,
    secretNamespace: props.secretRef?.namespace,
    username: props.username,
    password: props.password,
  };
  const { items: clusters, loading: clustersLoading, fetchList: fetchClusters } = useVSphereList('clusters', clusterParams);
  const { items: datastores, loading: dsLoading, fetchList: fetchDatastores } = useVSphereList('datastores', clusterParams);
  const { items: networks, loading: netLoading, fetchList: fetchNetworks } = useVSphereList('networks', clusterParams);
  const { items: resourcePools, loading: rpLoading, fetchList: fetchResourcePools } = useVSphereList('resourcepools', clusterParams);
  const { items: templates, loading: tmplLoading, fetchList: fetchTemplates } = useVSphereList('templates', clusterParams);
  const { items: folders, loading: folderLoading, fetchList: fetchFolders } = useVSphereList('folders', clusterParams);

  const handleConnect = React.useCallback(async () => {
    const result = await connect({
      server: props.server,
      username: props.username,
      password: props.password,
      secretRef: props.secretRef,
    });
    if (result?.datacenters?.length) {
      setDatacenters(result.datacenters);
      setConnected(true);
      props.onConnectSuccess?.(result.datacenters);
      if (result.datacenters.length && !props.datacenter) {
        props.onDatacenterChange(result.datacenters[0]);
      }
    }
  }, [connect, props]);

  React.useEffect(() => {
    if (connected && props.datacenter) {
      fetchClusters();
      fetchDatastores();
      fetchNetworks();
      fetchResourcePools();
      fetchTemplates();
      fetchFolders();
    }
  }, [connected, props.datacenter, fetchClusters, fetchDatastores, fetchNetworks, fetchResourcePools, fetchTemplates, fetchFolders]);

  return (
    <>
      {!connected && (
        <FormGroup label="Connect to vCenter" fieldId="vsphere-connect">
          <Button variant="secondary" onClick={handleConnect} isDisabled={connectLoading || !props.server}>
            {connectLoading ? (
              <>
                <Spinner size="sm" className="pf-v5-u-mr-sm" /> Connecting...
              </>
            ) : (
              'Test connection & load datacenters'
            )}
          </Button>
          {connectError && (
            <Alert variant="danger" title={connectError} isInline className="pf-v5-u-mt-sm" />
          )}
        </FormGroup>
      )}

      {connected && (
        <>
          <FormGroup label="Datacenter" isRequired fieldId="vsphere-datacenter">
            <FormSelect
              id="vsphere-datacenter"
              value={props.datacenter}
              onChange={(_e, v) => props.onDatacenterChange(v)}
              aria-label="Datacenter"
            >
              <FormSelectOption value="" label="Select datacenter" />
              {datacenters.map((dc) => (
                <FormSelectOption key={dc} value={dc} label={dc} />
              ))}
            </FormSelect>
          </FormGroup>

          <FormGroup label="Compute cluster" isRequired fieldId="vsphere-cluster">
            <FormSelect
              id="vsphere-cluster"
              value={props.cluster}
              onChange={(_e, v) => props.onClusterChange(v)}
              aria-label="Cluster"
              isDisabled={clustersLoading}
            >
              <FormSelectOption value="" label={clustersLoading ? 'Loading...' : 'Select cluster'} />
              {clusters.map((c) => (
                <FormSelectOption key={c} value={c} label={c} />
              ))}
            </FormSelect>
          </FormGroup>

          <FormGroup label="Datastore" isRequired fieldId="vsphere-datastore">
            <FormSelect
              id="vsphere-datastore"
              value={props.datastore}
              onChange={(_e, v) => props.onDatastoreChange(v)}
              aria-label="Datastore"
              isDisabled={dsLoading}
            >
              <FormSelectOption value="" label={dsLoading ? 'Loading...' : 'Select datastore'} />
              {datastores.map((d) => (
                <FormSelectOption key={d} value={d} label={d} />
              ))}
            </FormSelect>
          </FormGroup>

          <FormGroup label="Network" isRequired fieldId="vsphere-network">
            <FormSelect
              id="vsphere-network"
              value={props.network}
              onChange={(_e, v) => props.onNetworkChange(v)}
              aria-label="Network"
              isDisabled={netLoading}
            >
              <FormSelectOption value="" label={netLoading ? 'Loading...' : 'Select network'} />
              {networks.map((n) => (
                <FormSelectOption key={n} value={n} label={n} />
              ))}
            </FormSelect>
          </FormGroup>

          <FormGroup label="Resource pool" fieldId="vsphere-resource-pool">
            <FormSelect
              id="vsphere-resource-pool"
              value={props.resourcePool}
              onChange={(_e, v) => props.onResourcePoolChange(v)}
              aria-label="Resource pool"
              isDisabled={rpLoading}
            >
              <FormSelectOption value="" label={rpLoading ? 'Loading...' : 'Select resource pool'} />
              {resourcePools.map((rp) => (
                <FormSelectOption key={rp} value={rp} label={rp} />
              ))}
            </FormSelect>
          </FormGroup>

          <FormGroup label="Template (RHCOS)" isRequired fieldId="vsphere-template">
            <FormSelect
              id="vsphere-template"
              value={props.template}
              onChange={(_e, v) => props.onTemplateChange(v)}
              aria-label="Template"
              isDisabled={tmplLoading}
            >
              <FormSelectOption value="" label={tmplLoading ? 'Loading...' : 'Select template'} />
              {templates.map((t) => (
                <FormSelectOption key={t} value={t} label={t} />
              ))}
            </FormSelect>
          </FormGroup>

          <FormGroup label="Folder" fieldId="vsphere-folder">
            <FormSelect
              id="vsphere-folder"
              value={props.folder}
              onChange={(_e, v) => props.onFolderChange(v)}
              aria-label="Folder"
              isDisabled={folderLoading}
            >
              <FormSelectOption value="" label={folderLoading ? 'Loading...' : 'Select folder'} />
              {folders.map((f) => (
                <FormSelectOption key={f} value={f} label={f} />
              ))}
            </FormSelect>
          </FormGroup>
        </>
      )}
    </>
  );
};
