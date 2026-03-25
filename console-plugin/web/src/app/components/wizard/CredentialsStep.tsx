import * as React from 'react';
import {
  Form,
  FormGroup,
  FormHelperText,
  FormSection,
  HelperText,
  HelperTextItem,
  TextInput,
  Checkbox,
} from '@patternfly/react-core';

export interface CredentialsStepProps {
  server: string;
  onServerChange: (v: string) => void;
  username: string;
  onUsernameChange: (v: string) => void;
  password: string;
  onPasswordChange: (v: string) => void;
  useSecretRef: boolean;
  onUseSecretRefChange: (v: boolean) => void;
  secretName: string;
  onSecretNameChange: (v: string) => void;
  secretNamespace: string;
  onSecretNamespaceChange: (v: string) => void;
  createSecret: boolean;
  onCreateSecretChange: (v: boolean) => void;
  migrationName: string;
  onMigrationNameChange: (v: string) => void;
  migrationNamespace: string;
  onMigrationNamespaceChange: (v: string) => void;
}

export const CredentialsStep: React.FC<CredentialsStepProps> = (props) => (
  <Form>
    <FormSection title="Migration details">
      <FormGroup label="Migration name" isRequired fieldId="migration-name">
        <TextInput
          id="migration-name"
          value={props.migrationName}
          onChange={(_e, v) => props.onMigrationNameChange(v)}
          placeholder="vcf-migration"
        />
        <FormHelperText>
          <HelperText>
            <HelperTextItem>A unique name for this migration resource</HelperTextItem>
          </HelperText>
        </FormHelperText>
      </FormGroup>
      <FormGroup label="Namespace" isRequired fieldId="migration-namespace">
        <TextInput
          id="migration-namespace"
          value={props.migrationNamespace}
          onChange={(_e, v) => props.onMigrationNamespaceChange(v)}
        />
        <FormHelperText>
          <HelperText>
            <HelperTextItem>The namespace where the migration resource will be created</HelperTextItem>
          </HelperText>
        </FormHelperText>
      </FormGroup>
    </FormSection>
    <FormSection title="Target vCenter">
      <FormGroup label="vCenter server" isRequired fieldId="server">
        <TextInput
          id="server"
          value={props.server}
          onChange={(_e, v) => props.onServerChange(v)}
          placeholder="vcenter.example.com"
        />
        <FormHelperText>
          <HelperText>
            <HelperTextItem>FQDN or IP address of the target vCenter server</HelperTextItem>
          </HelperText>
        </FormHelperText>
      </FormGroup>
      <FormGroup fieldId="use-secret">
        <Checkbox
          id="use-secret"
          label="Use existing secret for credentials"
          isChecked={props.useSecretRef}
          onChange={(_e, v) => props.onUseSecretRefChange(v)}
          description="Reference an existing Kubernetes secret containing vCenter credentials"
        />
      </FormGroup>
      {props.useSecretRef ? (
        <>
          <FormGroup label="Secret name" isRequired fieldId="secret-name">
            <TextInput
              id="secret-name"
              value={props.secretName}
              onChange={(_e, v) => props.onSecretNameChange(v)}
            />
          </FormGroup>
          <FormGroup label="Secret namespace" fieldId="secret-namespace">
            <TextInput
              id="secret-namespace"
              value={props.secretNamespace}
              onChange={(_e, v) => props.onSecretNamespaceChange(v)}
            />
          </FormGroup>
        </>
      ) : (
        <>
          <FormGroup label="Username" isRequired fieldId="username">
            <TextInput
              id="username"
              value={props.username}
              onChange={(_e, v) => props.onUsernameChange(v)}
            />
          </FormGroup>
          <FormGroup label="Password" isRequired fieldId="password">
            <TextInput
              id="password"
              type="password"
              value={props.password}
              onChange={(_e, v) => props.onPasswordChange(v)}
            />
          </FormGroup>
          <FormGroup fieldId="create-secret">
            <Checkbox
              id="create-secret"
              label="Create a secret from these credentials (recommended)"
              isChecked={props.createSecret}
              onChange={(_e, v) => props.onCreateSecretChange(v)}
              description="Stores credentials securely as a Kubernetes secret for the operator to use"
            />
          </FormGroup>
        </>
      )}
    </FormSection>
  </Form>
);
