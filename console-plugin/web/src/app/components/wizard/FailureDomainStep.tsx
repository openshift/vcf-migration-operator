import * as React from 'react';
import {
  Alert,
  Bullseye,
  Button,
  EmptyState,
  EmptyStateBody,
  EmptyStateFooter,
  EmptyStateActions,
  EmptyStateHeader,
  EmptyStateIcon,
  Spinner,
  Stack,
  StackItem,
  TextContent,
  Text,
  TextVariants,
} from '@patternfly/react-core';
import { PlusCircleIcon, CubesIcon } from '@patternfly/react-icons';
import type { VSpherePlatformFailureDomainSpec } from '../../../models';
import { useVSphereConnect } from '../../hooks/useVSphereBrowse';
import { FailureDomainEditor } from './FailureDomainEditor';

export interface FailureDomainStepProps {
  server: string;
  username?: string;
  password?: string;
  secretRef?: { name: string; namespace?: string };
  failureDomains: VSpherePlatformFailureDomainSpec[];
  onFailureDomainsChange: (fd: VSpherePlatformFailureDomainSpec[]) => void;
}

const emptyFailureDomain = (server: string): VSpherePlatformFailureDomainSpec => ({
  name: '',
  region: '',
  zone: '',
  server,
  topology: {
    datacenter: '',
    computeCluster: '',
    datastore: '',
    networks: [],
    template: '',
    folder: '',
  },
});

export const FailureDomainStep: React.FC<FailureDomainStepProps> = (props) => {
  const { connect, loading: connecting, error: connectError } = useVSphereConnect();
  const [datacenters, setDatacenters] = React.useState<string[]>([]);
  const [connected, setConnected] = React.useState(false);
  const connectAttempted = React.useRef(false);

  React.useEffect(() => {
    if (connectAttempted.current || connected || !props.server) return;
    connectAttempted.current = true;

    connect({
      server: props.server,
      username: props.username,
      password: props.password,
      secretRef: props.secretRef,
    }).then((result) => {
      if (result?.datacenters?.length) {
        setDatacenters(result.datacenters);
        setConnected(true);
      }
    });
  }, [props.server, props.username, props.password, props.secretRef, connect, connected]);

  const handleRetry = React.useCallback(() => {
    connectAttempted.current = false;
    setConnected(false);
    setDatacenters([]);
    connect({
      server: props.server,
      username: props.username,
      password: props.password,
      secretRef: props.secretRef,
    }).then((result) => {
      if (result?.datacenters?.length) {
        setDatacenters(result.datacenters);
        setConnected(true);
      }
    });
  }, [props.server, props.username, props.password, props.secretRef, connect]);

  const addDomain = () => {
    props.onFailureDomainsChange([
      ...props.failureDomains,
      emptyFailureDomain(props.server),
    ]);
  };

  const removeDomain = (index: number) => {
    props.onFailureDomainsChange(props.failureDomains.filter((_, i) => i !== index));
  };

  const updateDomain = (index: number, fd: VSpherePlatformFailureDomainSpec) => {
    const next = [...props.failureDomains];
    next[index] = fd;
    props.onFailureDomainsChange(next);
  };

  if (connecting) {
    return (
      <Bullseye>
        <Stack hasGutter>
          <StackItem isFilled>
            <Bullseye>
              <Spinner size="xl" aria-label="Connecting to vCenter" />
            </Bullseye>
          </StackItem>
          <StackItem>
            <Bullseye>
              <TextContent>
                <Text component={TextVariants.p}>Connecting to {props.server}...</Text>
              </TextContent>
            </Bullseye>
          </StackItem>
        </Stack>
      </Bullseye>
    );
  }

  if (connectError && !connected) {
    return (
      <Stack hasGutter>
        <StackItem>
          <Alert variant="danger" title="Failed to connect to vCenter" isInline>
            {connectError}
          </Alert>
        </StackItem>
        <StackItem>
          <Button variant="secondary" onClick={handleRetry}>
            Retry connection
          </Button>
        </StackItem>
      </Stack>
    );
  }

  if (!connected) {
    return (
      <Bullseye>
        <Spinner size="xl" aria-label="Connecting to vCenter" />
      </Bullseye>
    );
  }

  if (props.failureDomains.length === 0) {
    return (
      <EmptyState>
        <EmptyStateHeader
          titleText="No failure domains configured"
          headingLevel="h4"
          icon={<EmptyStateIcon icon={CubesIcon} />}
        />
        <EmptyStateBody>
          Connected to {props.server} ({datacenters.length} datacenter{datacenters.length !== 1 ? 's' : ''} found).
          Add at least one failure domain to define the topology for the migration.
        </EmptyStateBody>
        <EmptyStateFooter>
          <EmptyStateActions>
            <Button variant="primary" icon={<PlusCircleIcon />} onClick={addDomain}>
              Add failure domain
            </Button>
          </EmptyStateActions>
        </EmptyStateFooter>
      </EmptyState>
    );
  }

  return (
    <Stack hasGutter>
      <StackItem>
        <Button variant="secondary" icon={<PlusCircleIcon />} onClick={addDomain}>
          Add failure domain
        </Button>
      </StackItem>
      {props.failureDomains.map((fd, i) => (
        <StackItem key={i}>
          <FailureDomainEditor
            index={i}
            domain={fd}
            onUpdate={(updated) => updateDomain(i, updated)}
            onRemove={() => removeDomain(i)}
            datacenters={datacenters}
            server={props.server}
            username={props.username}
            password={props.password}
            secretRef={props.secretRef}
          />
        </StackItem>
      ))}
    </Stack>
  );
};
