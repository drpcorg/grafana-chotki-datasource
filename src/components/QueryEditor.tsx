import React, { ChangeEvent } from 'react';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import {
  InlineField,
  InlineFieldRow,
  Input,
  RadioButtonGroup,
  Select,
  Stack,
  Switch,
  TextArea,
} from '@grafana/ui';
import { DataSource } from '../datasource';
import { ChotkiDataSourceOptions, ChotkiQuery, QueryParamValue, RpcMethod } from '../types';

type Props = QueryEditorProps<DataSource, ChotkiQuery, ChotkiDataSourceOptions>;

type ParamType = 'string' | 'number' | 'boolean';

type ParamSchema = {
  name: string;
  label: string;
  type: ParamType;
  required?: boolean;
  placeholder?: string;
};

const methodOptions: Array<SelectableValue<RpcMethod>> = [
  { value: 'GetOwner', label: 'GetOwner' },
  { value: 'GetFullOwner', label: 'GetFullOwner' },
  { value: 'GetOwnerHits', label: 'GetOwnerHits' },
  { value: 'GetOwnerMetadata', label: 'GetOwnerMetadata' },
  { value: 'GetKey', label: 'GetKey' },
  { value: 'GetKeyHits', label: 'GetKeyHits' },
  { value: 'ListKeys', label: 'ListKeys' },
  { value: 'GetOwnersWithBalance', label: 'GetOwnersWithBalance' },
  { value: 'GetAllOwnerIds', label: 'GetAllOwnerIds' },
  { value: 'GetNodeCoreKey', label: 'GetNodeCoreKey' },
  { value: 'ListNodeCoreKeys', label: 'ListNodeCoreKeys' },
];

const formatOptions: Array<SelectableValue<'table' | 'stat'>> = [
  { value: 'table', label: 'table' },
  { value: 'stat', label: 'stat' },
];

const methodParams: Record<RpcMethod, ParamSchema[]> = {
  GetOwner: [{ name: 'ownerId', label: 'ownerId', type: 'string', required: true, placeholder: 'uuid' }],
  GetFullOwner: [
    { name: 'ownerId', label: 'ownerId', type: 'string', required: true, placeholder: 'uuid' },
    { name: 'loadBalance', label: 'loadBalance', type: 'boolean' },
  ],
  GetOwnerHits: [{ name: 'ownerId', label: 'ownerId', type: 'string', required: true, placeholder: 'uuid' }],
  GetOwnerMetadata: [{ name: 'ownerId', label: 'ownerId', type: 'string', required: true, placeholder: 'uuid' }],
  GetKey: [
    { name: 'keyId', label: 'keyId', type: 'string', required: true, placeholder: 'uuid' },
    { name: 'ownerId', label: 'ownerId', type: 'string', required: true, placeholder: 'uuid' },
  ],
  GetKeyHits: [{ name: 'keyId', label: 'keyId', type: 'string', required: true, placeholder: 'uuid' }],
  ListKeys: [
    { name: 'ownerId', label: 'ownerId', type: 'string', required: true, placeholder: 'uuid' },
    { name: 'lastKeyId', label: 'lastKeyId', type: 'string', placeholder: 'uuid (optional)' },
    { name: 'limit', label: 'limit', type: 'number' },
  ],
  GetOwnersWithBalance: [],
  GetAllOwnerIds: [],
  GetNodeCoreKey: [
    { name: 'keyId', label: 'keyId', type: 'string', required: true, placeholder: 'uuid' },
    { name: 'ownerId', label: 'ownerId', type: 'string', required: true, placeholder: 'uuid' },
  ],
  ListNodeCoreKeys: [
    { name: 'ownerId', label: 'ownerId', type: 'string', required: true, placeholder: 'uuid' },
    { name: 'lastKeyId', label: 'lastKeyId', type: 'string', placeholder: 'uuid (optional)' },
    { name: 'limit', label: 'limit', type: 'number' },
  ],
};

function buildRawQuery(method?: RpcMethod, params?: Record<string, QueryParamValue>, options?: ChotkiQuery['options']): string {
  return JSON.stringify(
    {
      method: method ?? 'GetOwnerHits',
      params: params ?? { ownerId: '' },
      options: options ?? { format: 'table' },
    },
    null,
    2
  );
}

function parseRawQuery(raw: string): Partial<ChotkiQuery> | null {
  const trimmed = raw.trim();
  if (!trimmed) {
    return null;
  }

  try {
    const parsed = JSON.parse(trimmed) as {
      method?: RpcMethod;
      params?: Record<string, QueryParamValue>;
      options?: ChotkiQuery['options'];
    };

    if (!parsed.method) {
      return null;
    }

    return {
      mode: 'rpc',
      method: parsed.method,
      params: parsed.params ?? {},
      options: parsed.options,
    };
  } catch {
    return null;
  }
}

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const editorMode = query.editorMode ?? 'builder';
  const selectedMethod = query.method ?? 'GetOwnerHits';
  const params = query.params ?? {};
  const options = query.options ?? {};

  const setQuery = (patch: Partial<ChotkiQuery>, run = false) => {
    onChange({
      ...query,
      mode: 'rpc',
      ...patch,
    });
    if (run) {
      onRunQuery();
    }
  };

  const setOptions = (patch: Partial<NonNullable<ChotkiQuery['options']>>) => {
    const nextOptions = {
      ...options,
      ...patch,
    };

    const sanitized = Object.fromEntries(Object.entries(nextOptions).filter(([, value]) => value !== undefined));
    setQuery({ options: sanitized }, true);
  };

  const onEditorModeChange = (mode: 'builder' | 'raw') => {
    if (mode === 'raw') {
      setQuery(
        {
          editorMode: mode,
          rawQuery: query.rawQuery || buildRawQuery(selectedMethod, params, options),
        },
        false
      );
      return;
    }

    const parsed = parseRawQuery(query.rawQuery || '');
    setQuery(
      {
        editorMode: mode,
        ...(parsed ?? {}),
      },
      false
    );
  };

  const onMethodChange = (value?: SelectableValue<RpcMethod>) => {
    const method = value?.value;
    if (!method) {
      return;
    }

    const schema = methodParams[method] ?? [];
    const nextParams: Record<string, QueryParamValue> = {};

    schema.forEach((param) => {
      const current = params[param.name];
      if (current !== undefined) {
        nextParams[param.name] = current;
        return;
      }

      if (param.type === 'boolean') {
        nextParams[param.name] = false;
      } else if (param.required) {
        nextParams[param.name] = '';
      }
    });

    setQuery(
      {
        method,
        params: nextParams,
        rawQuery: buildRawQuery(method, nextParams, options),
      },
      true
    );
  };

  const onStringParamChange = (name: string) => (event: ChangeEvent<HTMLInputElement>) => {
    const nextParams = { ...params, [name]: event.target.value };
    setQuery({ params: nextParams, rawQuery: buildRawQuery(selectedMethod, nextParams, options) }, false);
  };

  const onNumberParamChange = (name: string) => (event: ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value.trim();
    const nextParams = { ...params };
    if (value === '') {
      delete nextParams[name];
    } else {
      nextParams[name] = Number(value);
    }
    setQuery({ params: nextParams, rawQuery: buildRawQuery(selectedMethod, nextParams, options) }, false);
  };

  const onBooleanParamChange = (name: string) => (event: ChangeEvent<HTMLInputElement>) => {
    const nextParams = { ...params, [name]: event.currentTarget.checked };
    setQuery({ params: nextParams, rawQuery: buildRawQuery(selectedMethod, nextParams, options) }, true);
  };

  const onRawQueryChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    const rawQuery = event.target.value;
    const parsed = parseRawQuery(rawQuery);

    setQuery(
      {
        rawQuery,
        ...(parsed ?? {}),
      },
      false
    );
  };

  const paramSchema = methodParams[selectedMethod] ?? [];

  return (
    <Stack direction="column" gap={1}>
      <InlineField label="Mode" labelWidth={14}>
        <RadioButtonGroup
          id="query-editor-mode"
          aria-label="Mode"
          options={[
            { label: 'Builder', value: 'builder' },
            { label: 'Raw JSON', value: 'raw' },
          ]}
          value={editorMode}
          onChange={(value) => onEditorModeChange(value as 'builder' | 'raw')}
        />
      </InlineField>

      {editorMode === 'builder' && (
        <>
          <InlineField label="Method" labelWidth={14}>
            <Select<RpcMethod>
              inputId="query-editor-method"
              aria-label="Method"
              options={methodOptions}
              value={methodOptions.find((item) => item.value === selectedMethod)}
              onChange={onMethodChange}
              width={40}
            />
          </InlineField>

          {paramSchema.length > 0 && (
            <InlineFieldRow>
              {paramSchema.map((param) => (
                <InlineField label={param.label} key={param.name} labelWidth={14}>
                  {param.type === 'boolean' ? (
                    <Switch
                      value={Boolean(params[param.name] ?? false)}
                      onChange={onBooleanParamChange(param.name)}
                      id={`query-editor-param-${param.name}`}
                    />
                  ) : (
                    <Input
                      id={`query-editor-param-${param.name}`}
                      value={String(params[param.name] ?? '')}
                      onChange={param.type === 'number' ? onNumberParamChange(param.name) : onStringParamChange(param.name)}
                      placeholder={param.placeholder}
                      required={param.required}
                      width={30}
                      type={param.type === 'number' ? 'number' : 'text'}
                    />
                  )}
                </InlineField>
              ))}
            </InlineFieldRow>
          )}

          <InlineField label="Format" labelWidth={14}>
            <Select<'table' | 'stat'>
              inputId="query-editor-format"
              options={formatOptions}
              value={formatOptions.find((item) => item.value === (options.format ?? 'table'))}
              onChange={(value) => setOptions({ format: value.value as 'table' | 'stat' })}
              width={20}
            />
          </InlineField>

          <InlineField label="Limit" labelWidth={14}>
            <Input
              id="query-editor-limit"
              value={options.limit ?? ''}
              onChange={(event) => {
                const trimmed = event.currentTarget.value.trim();
                if (!trimmed) {
                  setOptions({ limit: undefined });
                  return;
                }
                const parsed = Number(trimmed);
                setOptions({ limit: Number.isFinite(parsed) ? parsed : undefined });
              }}
              type="number"
              width={20}
            />
          </InlineField>

          <InlineField label="Decode IDs" labelWidth={14}>
            <Switch
              id="query-editor-decode-ids"
              label="Decode IDs"
              value={Boolean(options.decodeIds ?? true)}
              onChange={(event) => setOptions({ decodeIds: event.currentTarget.checked })}
            />
          </InlineField>

          <InlineField label="Decode Enums" labelWidth={14}>
            <Switch
              id="query-editor-decode-enums"
              label="Decode Enums"
              value={Boolean(options.decodeEnums ?? true)}
              onChange={(event) => setOptions({ decodeEnums: event.currentTarget.checked })}
            />
          </InlineField>

          <InlineField label="Decode Time" labelWidth={14}>
            <Switch
              id="query-editor-decode-time"
              label="Decode Time"
              value={Boolean(options.decodeTimestamps ?? true)}
              onChange={(event) => setOptions({ decodeTimestamps: event.currentTarget.checked })}
            />
          </InlineField>
        </>
      )}

      {editorMode === 'raw' && (
        <InlineField label="Raw JSON" labelWidth={14} grow>
          <TextArea
            id="query-editor-raw-json"
            value={query.rawQuery || buildRawQuery(selectedMethod, params, options)}
            onChange={onRawQueryChange}
            rows={14}
          />
        </InlineField>
      )}
    </Stack>
  );
}
