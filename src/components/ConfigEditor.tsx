import React, { ChangeEvent } from 'react';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { InlineField, Input, SecretInput, Switch, VerticalGroup } from '@grafana/ui';
import { ChotkiDataSourceOptions, ChotkiSecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<ChotkiDataSourceOptions, ChotkiSecureJsonData> {}

export function ConfigEditor({ onOptionsChange, options }: Props) {
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const updateJsonData = (patch: Partial<ChotkiDataSourceOptions>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        ...patch,
      },
    });
  };

  const updateSecureData = (patch: Partial<ChotkiSecureJsonData>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...secureJsonData,
        ...patch,
      },
    });
  };

  const onNumberChange = (key: keyof ChotkiDataSourceOptions) => (event: ChangeEvent<HTMLInputElement>) => {
    const parsed = Number(event.target.value);
    updateJsonData({ [key]: Number.isFinite(parsed) ? parsed : undefined } as Partial<ChotkiDataSourceOptions>);
  };

  const onStringChange = (key: keyof ChotkiDataSourceOptions) => (event: ChangeEvent<HTMLInputElement>) => {
    updateJsonData({ [key]: event.target.value } as Partial<ChotkiDataSourceOptions>);
  };

  const onBoolChange = (key: keyof ChotkiDataSourceOptions) => (event: ChangeEvent<HTMLInputElement>) => {
    updateJsonData({ [key]: event.currentTarget.checked } as Partial<ChotkiDataSourceOptions>);
  };

  const onSecretChange = (key: keyof ChotkiSecureJsonData) => (event: ChangeEvent<HTMLInputElement>) => {
    updateSecureData({ [key]: event.target.value });
  };

  const onSecretReset = (key: keyof ChotkiSecureJsonData) => () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...secureJsonFields,
        [key]: false,
      },
      secureJsonData: {
        ...secureJsonData,
        [key]: '',
      },
    });
  };

  return (
    <VerticalGroup spacing="md">
      <InlineField label="gRPC Address" labelWidth={20} required>
        <Input
          id="config-editor-grpc-address"
          value={jsonData.grpcAddress || ''}
          onChange={onStringChange('grpcAddress')}
          placeholder="host:port"
          width={50}
        />
      </InlineField>

      <InlineField label="Insecure" labelWidth={20}>
        <Switch value={Boolean(jsonData.insecure ?? true)} onChange={onBoolChange('insecure')} />
      </InlineField>

      <InlineField label="Timeout (ms)" labelWidth={20}>
        <Input
          id="config-editor-timeout-ms"
          value={jsonData.timeoutMs ?? 4000}
          onChange={onNumberChange('timeoutMs')}
          type="number"
          width={20}
        />
      </InlineField>

      <InlineField label="Default Limit" labelWidth={20}>
        <Input
          id="config-editor-default-limit"
          value={jsonData.defaultLimit ?? 200}
          onChange={onNumberChange('defaultLimit')}
          type="number"
          width={20}
        />
      </InlineField>

      <InlineField label="Hard Limit" labelWidth={20}>
        <Input
          id="config-editor-hard-limit"
          value={jsonData.hardLimit ?? 1000}
          onChange={onNumberChange('hardLimit')}
          type="number"
          width={20}
        />
      </InlineField>

      <InlineField label="Decode IDs" labelWidth={20}>
        <Switch value={Boolean(jsonData.decodeIds ?? true)} onChange={onBoolChange('decodeIds')} />
      </InlineField>

      <InlineField label="Decode Enums" labelWidth={20}>
        <Switch value={Boolean(jsonData.decodeEnums ?? true)} onChange={onBoolChange('decodeEnums')} />
      </InlineField>

      <InlineField label="Decode Timestamps" labelWidth={20}>
        <Switch value={Boolean(jsonData.decodeTimestamps ?? true)} onChange={onBoolChange('decodeTimestamps')} />
      </InlineField>

      <InlineField label="Auth Token" labelWidth={20}>
        <SecretInput
          id="config-editor-auth-token"
          isConfigured={Boolean(secureJsonFields?.authToken)}
          value={secureJsonData?.authToken}
          onChange={onSecretChange('authToken')}
          onReset={onSecretReset('authToken')}
          width={50}
        />
      </InlineField>

      <InlineField label="TLS CA Cert" labelWidth={20}>
        <SecretInput
          id="config-editor-tls-ca-cert"
          isConfigured={Boolean(secureJsonFields?.tlsCACert)}
          value={secureJsonData?.tlsCACert}
          onChange={onSecretChange('tlsCACert')}
          onReset={onSecretReset('tlsCACert')}
          width={50}
        />
      </InlineField>

      <InlineField label="TLS Client Cert" labelWidth={20}>
        <SecretInput
          id="config-editor-tls-client-cert"
          isConfigured={Boolean(secureJsonFields?.tlsClientCert)}
          value={secureJsonData?.tlsClientCert}
          onChange={onSecretChange('tlsClientCert')}
          onReset={onSecretReset('tlsClientCert')}
          width={50}
        />
      </InlineField>

      <InlineField label="TLS Client Key" labelWidth={20}>
        <SecretInput
          id="config-editor-tls-client-key"
          isConfigured={Boolean(secureJsonFields?.tlsClientKey)}
          value={secureJsonData?.tlsClientKey}
          onChange={onSecretChange('tlsClientKey')}
          onReset={onSecretReset('tlsClientKey')}
          width={50}
        />
      </InlineField>
    </VerticalGroup>
  );
}
