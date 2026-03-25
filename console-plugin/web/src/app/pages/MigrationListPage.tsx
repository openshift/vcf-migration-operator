import * as React from 'react';
import {
  PageSection,
  Title,
  Button,
  Toolbar,
  ToolbarContent,
  ToolbarItem,
  EmptyState,
  EmptyStateBody,
  EmptyStateFooter,
  EmptyStateActions,
  EmptyStateHeader,
  EmptyStateIcon,
  Spinner,
  Bullseye,
  Label,
  Alert,
  Dropdown,
  DropdownItem,
  DropdownList,
  Divider,
  MenuToggle,
} from '@patternfly/react-core';
import {
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
} from '@patternfly/react-table';
import { CubesIcon, EllipsisVIcon } from '@patternfly/react-icons';
import { useHistory } from 'react-router-dom';
import { useK8sWatchResource, k8sPatch, k8sDelete } from '@openshift-console/dynamic-plugin-sdk';
import { VmwareCloudFoundationMigrationModel } from '../../models';
import type { VmwareCloudFoundationMigrationKind } from '../../models';

const migrationGVK = {
  group: 'migration.openshift.io',
  version: 'v1alpha1',
  kind: 'VmwareCloudFoundationMigration',
};

const migrationStates = ['Pending', 'Running', 'Paused'] as const;

const getStateColor = (state: string): 'blue' | 'green' | 'orange' | 'grey' => {
  switch (state) {
    case 'Running':
      return 'blue';
    case 'Paused':
      return 'orange';
    default:
      return 'grey';
  }
};

const getReadyColor = (status: string): 'green' | 'red' | 'grey' => {
  switch (status) {
    case 'True':
      return 'green';
    case 'False':
      return 'red';
    default:
      return 'grey';
  }
};

const formatAge = (timestamp: unknown): string => {
  if (!timestamp || typeof timestamp !== 'string') return '-';
  const created = new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - created.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  if (diffMins < 60) return `${diffMins}m`;
  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return `${diffHours}h`;
  const diffDays = Math.floor(diffHours / 24);
  return `${diffDays}d`;
};

interface RowActionsProps {
  migration: VmwareCloudFoundationMigrationKind;
  onError: (msg: string) => void;
}

const RowActions: React.FC<RowActionsProps> = ({ migration, onError }) => {
  const [isOpen, setIsOpen] = React.useState(false);

  const handleSetState = React.useCallback(
    async (state: string) => {
      setIsOpen(false);
      try {
        await k8sPatch({
          model: VmwareCloudFoundationMigrationModel,
          resource: migration,
          data: [{ op: 'replace', path: '/spec/state', value: state }],
        });
      } catch (e) {
        onError(`Failed to set state to ${state}: ${e instanceof Error ? e.message : String(e)}`);
      }
    },
    [migration, onError],
  );

  const handleDelete = React.useCallback(async () => {
    setIsOpen(false);
    try {
      await k8sDelete({
        model: VmwareCloudFoundationMigrationModel,
        resource: migration,
      });
    } catch (e) {
      onError(`Failed to delete: ${e instanceof Error ? e.message : String(e)}`);
    }
  }, [migration, onError]);

  return (
    <Dropdown
      isOpen={isOpen}
      onSelect={() => setIsOpen(false)}
      onOpenChange={setIsOpen}
      toggle={(toggleRef) => (
        <MenuToggle
          ref={toggleRef}
          variant="plain"
          onClick={(e) => {
            e.stopPropagation();
            setIsOpen((prev) => !prev);
          }}
          isExpanded={isOpen}
          aria-label="Actions"
        >
          <EllipsisVIcon />
        </MenuToggle>
      )}
      popperProps={{ position: 'right' }}
    >
      <DropdownList>
        {migrationStates.map((state) => (
          <DropdownItem
            key={state}
            onClick={(e) => {
              e.stopPropagation();
              handleSetState(state);
            }}
            isDisabled={migration.spec.state === state}
            description={migration.spec.state === state ? 'Current state' : undefined}
          >
            Set {state}
          </DropdownItem>
        ))}
        <Divider />
        <DropdownItem
          key="delete"
          onClick={(e) => {
            e.stopPropagation();
            handleDelete();
          }}
          isDanger
        >
          Delete
        </DropdownItem>
      </DropdownList>
    </Dropdown>
  );
};

export const MigrationListPage: React.FC = () => {
  const history = useHistory();
  const [actionError, setActionError] = React.useState<string | null>(null);
  const [migrations, loaded, loadError] = useK8sWatchResource<
    VmwareCloudFoundationMigrationKind[]
  >({
    groupVersionKind: migrationGVK,
    isList: true,
    namespaced: true,
  });

  const getReadyCondition = (m: VmwareCloudFoundationMigrationKind) => {
    const cond = m.status?.conditions?.find((c) => c.type === 'Ready');
    return cond?.status ?? 'Unknown';
  };

  return (
    <>
      <PageSection variant="light">
        <Toolbar>
          <ToolbarContent>
            <ToolbarItem>
              <Title headingLevel="h1">VCF Migrations</Title>
            </ToolbarItem>
            <ToolbarItem align={{ default: 'alignRight' }}>
              <Button variant="primary" onClick={() => history.push('/vcf-migration/create')}>
                Create migration
              </Button>
            </ToolbarItem>
          </ToolbarContent>
        </Toolbar>
      </PageSection>
      <PageSection>
        {actionError && (
          <Alert
            variant="danger"
            title="Action failed"
            isInline
            className="pf-v5-u-mb-md"
            actionClose={<Button variant="plain" onClick={() => setActionError(null)}>Dismiss</Button>}
          >
            {actionError}
          </Alert>
        )}
        {loadError && (
          <Alert variant="danger" title="Failed to load migrations" isInline className="pf-v5-u-mb-md">
            {String(loadError)}
          </Alert>
        )}
        {!loaded && !loadError && (
          <Bullseye>
            <Spinner size="xl" aria-label="Loading migrations" />
          </Bullseye>
        )}
        {loaded && !loadError && (!migrations || migrations.length === 0) && (
          <EmptyState>
            <EmptyStateHeader
              titleText="No migrations"
              headingLevel="h4"
              icon={<EmptyStateIcon icon={CubesIcon} />}
            />
            <EmptyStateBody>
              No VCF migrations have been created yet. Create a migration to begin
              moving your OpenShift cluster to a new vCenter.
            </EmptyStateBody>
            <EmptyStateFooter>
              <EmptyStateActions>
                <Button variant="primary" onClick={() => history.push('/vcf-migration/create')}>
                  Create migration
                </Button>
              </EmptyStateActions>
            </EmptyStateFooter>
          </EmptyState>
        )}
        {loaded && !loadError && migrations?.length > 0 && (
          <Table aria-label="Migrations table">
            <Thead>
              <Tr>
                <Th>Name</Th>
                <Th>Namespace</Th>
                <Th>State</Th>
                <Th>Ready</Th>
                <Th>Age</Th>
                <Th screenReaderText="Actions" />
              </Tr>
            </Thead>
            <Tbody>
              {migrations.map((m) => {
                const readyStatus = getReadyCondition(m);
                return (
                  <Tr
                    key={`${m.metadata.namespace}-${m.metadata.name}`}
                    isClickable
                    onRowClick={() =>
                      history.push(`/vcf-migration/ns/${m.metadata.namespace}/${m.metadata.name}`)
                    }
                  >
                    <Td dataLabel="Name">{m.metadata.name}</Td>
                    <Td dataLabel="Namespace">{m.metadata.namespace}</Td>
                    <Td dataLabel="State">
                      <Label color={getStateColor(m.spec.state)}>{m.spec.state}</Label>
                    </Td>
                    <Td dataLabel="Ready">
                      <Label color={getReadyColor(readyStatus)}>{readyStatus}</Label>
                    </Td>
                    <Td dataLabel="Age">{formatAge(m.metadata.creationTimestamp)}</Td>
                    <Td isActionCell>
                      <RowActions migration={m} onError={setActionError} />
                    </Td>
                  </Tr>
                );
              })}
            </Tbody>
          </Table>
        )}
      </PageSection>
    </>
  );
};
