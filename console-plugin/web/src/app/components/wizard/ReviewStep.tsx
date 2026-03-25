import * as React from 'react';
import {
  Card,
  CardBody,
  CardTitle,
  DescriptionList,
  DescriptionListGroup,
  DescriptionListTerm,
  DescriptionListDescription,
  Label,
  Stack,
  StackItem,
  TextContent,
  Text,
  TextVariants,
  Divider,
} from '@patternfly/react-core';
import type { VmwareCloudFoundationMigrationKind } from '../../../models';

export interface ReviewStepProps {
  migration: VmwareCloudFoundationMigrationKind;
}

export const ReviewStep: React.FC<ReviewStepProps> = ({ migration }) => (
  <Stack hasGutter>
    <StackItem>
      <TextContent>
        <Text component={TextVariants.p}>
          Review the migration configuration before creating the resource.
        </Text>
      </TextContent>
    </StackItem>
    <StackItem>
      <Card isPlain isCompact>
        <CardTitle>General</CardTitle>
        <CardBody>
          <DescriptionList isHorizontal termWidth="12ch">
            <DescriptionListGroup>
              <DescriptionListTerm>Name</DescriptionListTerm>
              <DescriptionListDescription>{migration.metadata.name}</DescriptionListDescription>
            </DescriptionListGroup>
            <DescriptionListGroup>
              <DescriptionListTerm>Namespace</DescriptionListTerm>
              <DescriptionListDescription>{migration.metadata.namespace}</DescriptionListDescription>
            </DescriptionListGroup>
            <DescriptionListGroup>
              <DescriptionListTerm>State</DescriptionListTerm>
              <DescriptionListDescription>
                <Label color="blue">{migration.spec.state}</Label>
              </DescriptionListDescription>
            </DescriptionListGroup>
          </DescriptionList>
        </CardBody>
      </Card>
    </StackItem>
    <StackItem>
      <Divider />
    </StackItem>
    <StackItem>
      <Card isPlain isCompact>
        <CardTitle>Credentials</CardTitle>
        <CardBody>
          <DescriptionList isHorizontal termWidth="12ch">
            <DescriptionListGroup>
              <DescriptionListTerm>Secret name</DescriptionListTerm>
              <DescriptionListDescription>
                {migration.spec.targetVCenterCredentialsSecret.name || '(not set)'}
              </DescriptionListDescription>
            </DescriptionListGroup>
            <DescriptionListGroup>
              <DescriptionListTerm>Secret namespace</DescriptionListTerm>
              <DescriptionListDescription>
                {migration.spec.targetVCenterCredentialsSecret.namespace || '(default)'}
              </DescriptionListDescription>
            </DescriptionListGroup>
          </DescriptionList>
        </CardBody>
      </Card>
    </StackItem>
    {migration.spec.failureDomains?.length > 0 && (
      <>
        <StackItem>
          <Divider />
        </StackItem>
        <StackItem>
          <TextContent>
            <Text component={TextVariants.h3}>
              Failure domains ({migration.spec.failureDomains.length})
            </Text>
          </TextContent>
        </StackItem>
        {migration.spec.failureDomains.map((fd, i) => (
          <StackItem key={i}>
            <Card isCompact>
              <CardTitle>{fd.name || `Failure domain ${i + 1}`}</CardTitle>
              <CardBody>
                <DescriptionList isHorizontal isCompact termWidth="14ch">
                  <DescriptionListGroup>
                    <DescriptionListTerm>Server</DescriptionListTerm>
                    <DescriptionListDescription>{fd.server || '-'}</DescriptionListDescription>
                  </DescriptionListGroup>
                  <DescriptionListGroup>
                    <DescriptionListTerm>Region</DescriptionListTerm>
                    <DescriptionListDescription>{fd.region || '-'}</DescriptionListDescription>
                  </DescriptionListGroup>
                  <DescriptionListGroup>
                    <DescriptionListTerm>Zone</DescriptionListTerm>
                    <DescriptionListDescription>{fd.zone || '-'}</DescriptionListDescription>
                  </DescriptionListGroup>
                  <DescriptionListGroup>
                    <DescriptionListTerm>Datacenter</DescriptionListTerm>
                    <DescriptionListDescription>{fd.topology.datacenter || '-'}</DescriptionListDescription>
                  </DescriptionListGroup>
                  <DescriptionListGroup>
                    <DescriptionListTerm>Compute cluster</DescriptionListTerm>
                    <DescriptionListDescription>{fd.topology.computeCluster || '-'}</DescriptionListDescription>
                  </DescriptionListGroup>
                  <DescriptionListGroup>
                    <DescriptionListTerm>Datastore</DescriptionListTerm>
                    <DescriptionListDescription>{fd.topology.datastore || '-'}</DescriptionListDescription>
                  </DescriptionListGroup>
                  <DescriptionListGroup>
                    <DescriptionListTerm>Network</DescriptionListTerm>
                    <DescriptionListDescription>
                      {fd.topology.networks?.join(', ') || '-'}
                    </DescriptionListDescription>
                  </DescriptionListGroup>
                  <DescriptionListGroup>
                    <DescriptionListTerm>Template</DescriptionListTerm>
                    <DescriptionListDescription>{fd.topology.template || '-'}</DescriptionListDescription>
                  </DescriptionListGroup>
                  {fd.topology.folder && (
                    <DescriptionListGroup>
                      <DescriptionListTerm>Folder</DescriptionListTerm>
                      <DescriptionListDescription>{fd.topology.folder}</DescriptionListDescription>
                    </DescriptionListGroup>
                  )}
                  {fd.topology.resourcePool && (
                    <DescriptionListGroup>
                      <DescriptionListTerm>Resource pool</DescriptionListTerm>
                      <DescriptionListDescription>{fd.topology.resourcePool}</DescriptionListDescription>
                    </DescriptionListGroup>
                  )}
                </DescriptionList>
              </CardBody>
            </Card>
          </StackItem>
        ))}
      </>
    )}
  </Stack>
);
