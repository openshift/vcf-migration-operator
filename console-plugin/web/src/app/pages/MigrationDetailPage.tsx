import * as React from 'react';
import { useParams, useHistory, useLocation } from 'react-router-dom';
import {
  PageSection,
  Title,
  Button,
  Breadcrumb,
  BreadcrumbItem,
  Card,
  CardBody,
  CardTitle,
  DescriptionList,
  DescriptionListGroup,
  DescriptionListTerm,
  DescriptionListDescription,
  Divider,
  Dropdown,
  DropdownItem,
  DropdownList,
  Flex,
  FlexItem,
  Label,
  MenuToggle,
  ProgressStepper,
  ProgressStep,
  Stack,
  StackItem,
  Spinner,
  Bullseye,
  Alert,
  Tab,
  Tabs,
  TabTitleText,
} from '@patternfly/react-core';
import { DownloadIcon } from '@patternfly/react-icons';
import { useK8sWatchResource, k8sPatch, k8sDelete, ResourceYAMLEditor, consoleFetch } from '@openshift-console/dynamic-plugin-sdk';
import { VmwareCloudFoundationMigrationModel } from '../../models';
import type { VmwareCloudFoundationMigrationKind } from '../../models';
import { EventStream } from '../components/EventStream';
import { MachineTopologyView } from '../components/MachineTopologyView';
import { MigrationLogs } from '../components/MigrationLogs';

const migrationGVK = {
  group: 'migration.openshift.io',
  version: 'v1alpha1',
  kind: 'VmwareCloudFoundationMigration',
};

const conditionOrder = [
  'InfrastructurePrepared',
  'DestinationInitialized',
  'MultiSiteConfigured',
  'WorkloadMigrated',
  'SourceCleaned',
  'Ready',
];

const conditionLabels: Record<string, string> = {
  InfrastructurePrepared: 'Infrastructure prepared',
  DestinationInitialized: 'Destination initialized',
  MultiSiteConfigured: 'Multi-site configured',
  WorkloadMigrated: 'Workload migrated',
  SourceCleaned: 'Source cleaned',
  Ready: 'Ready',
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

export const MigrationDetailPage: React.FC = () => {
  const params = useParams<{ ns: string; name: string }>();
  const location = useLocation();
  const history = useHistory();

  const { ns, name } = React.useMemo(() => {
    if (params.ns && params.name) return params;
    const match = location.pathname.match(/\/vcf-migration\/ns\/([^/]+)\/([^/]+)/);
    if (match) return { ns: match[1], name: match[2] };
    return { ns: '', name: '' };
  }, [params, location.pathname]);

  const watchSpec = ns && name
    ? {
        groupVersionKind: migrationGVK,
        name,
        namespace: ns,
        namespaced: true,
        isList: false,
      }
    : null;

  const [migration, loaded, loadError] = useK8sWatchResource<VmwareCloudFoundationMigrationKind>(watchSpec as Parameters<typeof useK8sWatchResource>[0]);

  const [actionsOpen, setActionsOpen] = React.useState(false);
  const [actionError, setActionError] = React.useState<string | null>(null);
  const [activeTab, setActiveTab] = React.useState<string | number>('details');

  const handleSetState = React.useCallback(
    async (state: string) => {
      setActionsOpen(false);
      if (!migration) return;
      try {
        await k8sPatch({
          model: VmwareCloudFoundationMigrationModel,
          resource: migration,
          data: [{ op: 'replace', path: '/spec/state', value: state }],
        });
      } catch (e) {
        setActionError(`Failed to set state to ${state}: ${e instanceof Error ? e.message : String(e)}`);
      }
    },
    [migration],
  );

  const handleDelete = React.useCallback(async () => {
    setActionsOpen(false);
    if (!migration) return;
    try {
      await k8sDelete({
        model: VmwareCloudFoundationMigrationModel,
        resource: migration,
      });
      history.push('/vcf-migration');
    } catch (e) {
      setActionError(`Failed to delete: ${e instanceof Error ? e.message : String(e)}`);
    }
  }, [migration, history]);

  const handleDownloadMetadata = React.useCallback(async () => {
    if (!ns || !name) return;
    try {
      const url = `/api/proxy/plugin/vcf-migration-console/vcf-migration-api/metadata?namespace=${encodeURIComponent(ns)}&name=${encodeURIComponent(name)}`;
      const response = await consoleFetch(url);
      const blob = await (response as Response).blob();
      const link = document.createElement('a');
      link.href = URL.createObjectURL(blob);
      link.download = `${name}-metadata.json`;
      link.click();
      URL.revokeObjectURL(link.href);
    } catch (e) {
      setActionError(`Failed to download metadata: ${e instanceof Error ? e.message : String(e)}`);
    }
  }, [ns, name]);

  if (!ns || !name) {
    return (
      <PageSection>
        <Title headingLevel="h1">Migration not found</Title>
        <Button variant="link" onClick={() => history.push('/vcf-migration')}>
          Back to list
        </Button>
      </PageSection>
    );
  }

  if (loadError) {
    return (
      <PageSection>
        <Alert variant="danger" title="Error loading migration" isInline>
          {String(loadError)}
        </Alert>
        <Button variant="link" onClick={() => history.push('/vcf-migration')} className="pf-v5-u-mt-md">
          Back to list
        </Button>
      </PageSection>
    );
  }

  if (!loaded || !migration) {
    return (
      <Bullseye>
        <Spinner size="xl" aria-label="Loading migration" />
      </Bullseye>
    );
  }

  const getCondition = (type: string) =>
    migration.status?.conditions?.find((c) => c.type === type);
  const isConditionTrue = (type: string) => getCondition(type)?.status === 'True';

  return (
    <>
      {/* Header - matches OpenShift resource detail style */}
      <PageSection variant="light" className="pf-v5-u-pb-0">
        <Breadcrumb className="pf-v5-u-mb-sm">
          <BreadcrumbItem onClick={() => history.push('/vcf-migration')}>
            VCF Migrations
          </BreadcrumbItem>
          <BreadcrumbItem isActive>Migration details</BreadcrumbItem>
        </Breadcrumb>

        <Flex alignItems={{ default: 'alignItemsCenter' }}>
          <FlexItem>
            <Title headingLevel="h1" className="pf-v5-u-mr-sm" style={{ display: 'inline' }}>
              {migration.metadata.name}
            </Title>
          </FlexItem>
          <FlexItem>
            <Label color={getStateColor(migration.spec.state)} isCompact>
              {migration.spec.state}
            </Label>
          </FlexItem>
          <FlexItem align={{ default: 'alignRight' }}>
            <Dropdown
              isOpen={actionsOpen}
              onSelect={() => setActionsOpen(false)}
              onOpenChange={setActionsOpen}
              toggle={(toggleRef) => (
                <MenuToggle ref={toggleRef} variant="primary" onClick={() => setActionsOpen((prev) => !prev)} isExpanded={actionsOpen}>
                  Actions
                </MenuToggle>
              )}
              popperProps={{ position: 'right' }}
            >
              <DropdownList>
                {migrationStates.map((state) => (
                  <DropdownItem
                    key={state}
                    onClick={() => handleSetState(state)}
                    isDisabled={migration.spec.state === state}
                    description={migration.spec.state === state ? 'Current state' : undefined}
                  >
                    Set {state}
                  </DropdownItem>
                ))}
                <Divider />
                <DropdownItem key="delete" onClick={handleDelete} isDanger>
                  Delete
                </DropdownItem>
              </DropdownList>
            </Dropdown>
          </FlexItem>
        </Flex>

        {actionError && (
          <Alert
            variant="danger"
            title="Action failed"
            isInline
            className="pf-v5-u-mt-sm"
            actionClose={<Button variant="plain" onClick={() => setActionError(null)}>Dismiss</Button>}
          >
            {actionError}
          </Alert>
        )}

        <Tabs
          activeKey={activeTab}
          onSelect={(_e, key) => setActiveTab(key)}
          className="pf-v5-u-mt-md"
          style={{ marginBottom: -1 }}
        >
          <Tab eventKey="details" title={<TabTitleText>Details</TabTitleText>} />
          <Tab eventKey="yaml" title={<TabTitleText>YAML</TabTitleText>} />
          <Tab eventKey="machines" title={<TabTitleText>Machines</TabTitleText>} />
          <Tab eventKey="logs" title={<TabTitleText>Logs</TabTitleText>} />
          <Tab eventKey="events" title={<TabTitleText>Events</TabTitleText>} />
        </Tabs>
      </PageSection>

      {/* Details tab */}
      {activeTab === 'details' && (
        <PageSection>
          <Stack hasGutter>
            <StackItem>
              <Card>
                <CardTitle>Overview</CardTitle>
                <CardBody>
                  <DescriptionList isHorizontal columnModifier={{ default: '2Col' }}>
                    <DescriptionListGroup>
                      <DescriptionListTerm>Namespace</DescriptionListTerm>
                      <DescriptionListDescription>{migration.metadata.namespace}</DescriptionListDescription>
                    </DescriptionListGroup>
                    <DescriptionListGroup>
                      <DescriptionListTerm>State</DescriptionListTerm>
                      <DescriptionListDescription>
                        <Label color={getStateColor(migration.spec.state)}>{migration.spec.state}</Label>
                      </DescriptionListDescription>
                    </DescriptionListGroup>
                    <DescriptionListGroup>
                      <DescriptionListTerm>Start time</DescriptionListTerm>
                      <DescriptionListDescription>
                        {migration.status?.startTime
                          ? new Date(migration.status.startTime).toLocaleString()
                          : '-'}
                      </DescriptionListDescription>
                    </DescriptionListGroup>
                    <DescriptionListGroup>
                      <DescriptionListTerm>Completion time</DescriptionListTerm>
                      <DescriptionListDescription>
                        {migration.status?.completionTime
                          ? new Date(migration.status.completionTime).toLocaleString()
                          : '-'}
                      </DescriptionListDescription>
                    </DescriptionListGroup>
                  </DescriptionList>
                </CardBody>
              </Card>
            </StackItem>

            {isConditionTrue('SourceCleaned') && (
              <StackItem>
                <Card>
                  <CardTitle>Installer metadata</CardTitle>
                  <CardBody>
                    <Button
                      variant="secondary"
                      icon={<DownloadIcon />}
                      onClick={handleDownloadMetadata}
                    >
                      Download metadata.json
                    </Button>
                    <p className="pf-v5-u-mt-sm pf-v5-u-color-200" style={{ fontSize: 'var(--pf-v5-global--FontSize--sm)' }}>
                      Replacement installer metadata with destination vCenter configuration.
                      Use this file to destroy the cluster with <code>openshift-install destroy cluster</code>.
                    </p>
                  </CardBody>
                </Card>
              </StackItem>
            )}

            <StackItem>
              <Card>
                <CardTitle>Migration progress</CardTitle>
                <CardBody>
                  <ProgressStepper isVertical>
                    {conditionOrder.map((type) => {
                      const cond = getCondition(type);
                      const isDone = isConditionTrue(type);
                      const isCurrent =
                        cond?.status === 'False' &&
                        cond?.reason !== 'Failed';
                      let variant: 'default' | 'success' | 'danger' | 'info' = 'default';
                      if (isDone) variant = 'success';
                      else if (cond?.reason === 'Failed') variant = 'danger';
                      else if (isCurrent) variant = 'info';
                      return (
                        <ProgressStep
                          key={type}
                          variant={variant}
                          id={type}
                          titleId={`${type}-title`}
                          aria-label={conditionLabels[type] || type}
                          description={cond?.message ?? (isDone ? 'Complete' : 'Pending')}
                        >
                          {conditionLabels[type] || type}
                        </ProgressStep>
                      );
                    })}
                  </ProgressStepper>
                </CardBody>
              </Card>
            </StackItem>
          </Stack>
        </PageSection>
      )}

      {/* YAML tab */}
      {activeTab === 'yaml' && (
        <div style={{ display: 'flex', flex: 1, flexDirection: 'column', height: 'calc(100vh - 250px)', minHeight: 400 }}>
          <React.Suspense fallback={<Bullseye><Spinner size="xl" aria-label="Loading editor" /></Bullseye>}>
            <ResourceYAMLEditor initialResource={migration} />
          </React.Suspense>
        </div>
      )}

      {/* Machines tab */}
      {activeTab === 'machines' && (
        <PageSection>
          <MachineTopologyView namespace={ns} />
        </PageSection>
      )}

      {/* Logs tab */}
      {activeTab === 'logs' && (
        <div style={{ display: 'flex', flexDirection: 'column', height: 'calc(100vh - 250px)', minHeight: 400 }}>
          <MigrationLogs />
        </div>
      )}

      {/* Events tab */}
      {activeTab === 'events' && (
        <PageSection>
          <EventStream namespace={ns} name={name} />
        </PageSection>
      )}
    </>
  );
};
