import * as React from 'react';
import {
  Button,
  PageSection,
  Wizard,
  WizardHeader,
  WizardStep,
  WizardFooterWrapper,
  useWizardContext,
  Alert,
} from '@patternfly/react-core';
import { useHistory } from 'react-router-dom';
import { k8sCreate } from '@openshift-console/dynamic-plugin-sdk';
import { consoleFetch } from '@openshift-console/dynamic-plugin-sdk';
import { VmwareCloudFoundationMigrationModel } from '../../models';
import type { VmwareCloudFoundationMigrationKind, VSpherePlatformFailureDomainSpec } from '../../models';
import { CredentialsStep } from '../components/wizard/CredentialsStep';
import { FailureDomainStep } from '../components/wizard/FailureDomainStep';
import { ReviewStep } from '../components/wizard/ReviewStep';

interface CreateFooterProps {
  onCreate: () => Promise<void>;
  isCreating: boolean;
}

const CreateMigrationFooter: React.FC<CreateFooterProps> = ({ onCreate, isCreating }) => {
  const { goToPrevStep, close } = useWizardContext();
  return (
    <WizardFooterWrapper>
      <Button variant="primary" onClick={onCreate} isLoading={isCreating} isDisabled={isCreating}>
        Create migration
      </Button>
      <Button variant="secondary" onClick={goToPrevStep} isDisabled={isCreating}>
        Back
      </Button>
      <Button variant="link" onClick={close} isDisabled={isCreating}>
        Cancel
      </Button>
    </WizardFooterWrapper>
  );
};

export const MigrationWizard: React.FC = () => {
  const history = useHistory();
  const [migrationName, setMigrationName] = React.useState('');
  const [migrationNamespace, setMigrationNamespace] = React.useState('openshift-vcf-migration');
  const [server, setServer] = React.useState('');
  const [username, setUsername] = React.useState('');
  const [password, setPassword] = React.useState('');
  const [useSecretRef, setUseSecretRef] = React.useState(false);
  const [secretName, setSecretName] = React.useState('');
  const [secretNamespace, setSecretNamespace] = React.useState('openshift-vcf-migration');
  const [createSecret, setCreateSecret] = React.useState(false);
  const [failureDomains, setFailureDomains] = React.useState<VSpherePlatformFailureDomainSpec[]>([]);
  const [createError, setCreateError] = React.useState<string | null>(null);
  const [isCreating, setIsCreating] = React.useState(false);

  const buildMigration = React.useCallback((): VmwareCloudFoundationMigrationKind => {
    const effectiveName = migrationName || 'vcf-migration';
    return {
      apiVersion: `${VmwareCloudFoundationMigrationModel.apiGroup}/${VmwareCloudFoundationMigrationModel.apiVersion}`,
      kind: VmwareCloudFoundationMigrationModel.kind,
      metadata: {
        name: effectiveName,
        namespace: migrationNamespace,
      },
      spec: {
        state: 'Pending',
        targetVCenterCredentialsSecret: {
          name: useSecretRef ? secretName : (createSecret ? `${effectiveName}-vcenter-creds` : secretName),
          namespace: useSecretRef ? secretNamespace : (createSecret ? migrationNamespace : secretNamespace),
        },
        failureDomains,
      },
    };
  }, [migrationName, migrationNamespace, useSecretRef, secretName, secretNamespace, createSecret, failureDomains]);

  const handleCreate = React.useCallback(async () => {
    setCreateError(null);
    setIsCreating(true);
    const obj = buildMigration();
    try {
      if (createSecret && !useSecretRef && username && password && server) {
        const secretNs = obj.spec.targetVCenterCredentialsSecret.namespace || obj.metadata.namespace;
        const secretBody = {
          apiVersion: 'v1',
          kind: 'Secret',
          metadata: {
            name: obj.spec.targetVCenterCredentialsSecret.name,
            namespace: secretNs,
          },
          type: 'Opaque',
          stringData: {
            [`${server}.username`]: username,
            [`${server}.password`]: password,
          },
        };
        const res = await consoleFetch(`/api/kubernetes/api/v1/namespaces/${secretNs}/secrets`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(secretBody),
        });
        if (!res.ok) {
          const errData = await res.json().catch(() => ({}));
          const msg = (errData as Record<string, string>).message || res.statusText;
          throw new Error(`Failed to create credentials secret: ${msg}`);
        }
      }
      await k8sCreate({
        model: VmwareCloudFoundationMigrationModel,
        data: obj,
      });
      history.push(`/vcf-migration/ns/${obj.metadata.namespace}/${obj.metadata.name}`);
    } catch (e) {
      setCreateError(e instanceof Error ? e.message : String(e));
    } finally {
      setIsCreating(false);
    }
  }, [buildMigration, history, createSecret, useSecretRef, username, password, server]);

  const handleClose = React.useCallback(() => {
    history.push('/vcf-migration');
  }, [history]);

  return (
    <PageSection isFilled padding={{ default: 'noPadding' }}>
      <Wizard
        header={
          <WizardHeader
            title="Create VCF Migration"
            description="Configure target vCenter credentials and failure domains for the migration"
            onClose={handleClose}
            closeButtonAriaLabel="Close wizard"
          />
        }
        onClose={handleClose}
      >
        <WizardStep name="Credentials" id="credentials">
          <CredentialsStep
            server={server}
            onServerChange={setServer}
            username={username}
            onUsernameChange={setUsername}
            password={password}
            onPasswordChange={setPassword}
            useSecretRef={useSecretRef}
            onUseSecretRefChange={setUseSecretRef}
            secretName={secretName}
            onSecretNameChange={setSecretName}
            secretNamespace={secretNamespace}
            onSecretNamespaceChange={setSecretNamespace}
            createSecret={createSecret}
            onCreateSecretChange={setCreateSecret}
            migrationName={migrationName}
            migrationNamespace={migrationNamespace}
            onMigrationNameChange={setMigrationName}
            onMigrationNamespaceChange={setMigrationNamespace}
          />
        </WizardStep>
        <WizardStep name="Failure domains" id="failure-domains">
          <FailureDomainStep
            server={server}
            username={username}
            password={password}
            secretRef={useSecretRef ? { name: secretName, namespace: secretNamespace } : undefined}
            failureDomains={failureDomains}
            onFailureDomainsChange={setFailureDomains}
          />
        </WizardStep>
        <WizardStep
          name="Review"
          id="review"
          footer={<CreateMigrationFooter onCreate={handleCreate} isCreating={isCreating} />}
        >
          {createError && (
            <Alert variant="danger" title="Migration creation failed" isInline className="pf-v5-u-mb-md">
              {createError}
            </Alert>
          )}
          <ReviewStep migration={buildMigration()} />
        </WizardStep>
      </Wizard>
    </PageSection>
  );
};
