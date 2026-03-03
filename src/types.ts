import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export type RpcMethod =
  | 'GetOwner'
  | 'GetFullOwner'
  | 'GetOwnerHits'
  | 'GetOwnerMetadata'
  | 'GetKey'
  | 'GetKeyHits'
  | 'ListKeys'
  | 'GetOwnersWithBalance'
  | 'GetAllOwnerIds'
  | 'GetNodeCoreKey'
  | 'ListNodeCoreKeys';

export type QueryFormat = 'table' | 'stat';
export type QueryEditorMode = 'builder' | 'raw';
export type QueryParamValue = string | number | boolean | string[] | number[];

export interface QueryOptions {
  format?: QueryFormat;
  decodeIds?: boolean;
  decodeEnums?: boolean;
  decodeTimestamps?: boolean;
  limit?: number;
}

export interface ChotkiQuery extends DataQuery {
  mode: 'rpc';
  method?: RpcMethod;
  params: Record<string, QueryParamValue>;
  options?: QueryOptions;
  editorMode?: QueryEditorMode;
  rawQuery?: string;
}

export const DEFAULT_QUERY: Partial<ChotkiQuery> = {
  mode: 'rpc',
  method: 'GetOwnerHits',
  params: {
    ownerId: '',
  },
  options: {
    format: 'table',
  },
  editorMode: 'builder',
};

export interface ChotkiDataSourceOptions extends DataSourceJsonData {
  grpcAddress?: string;
  insecure?: boolean;
  timeoutMs?: number;
  defaultLimit?: number;
  hardLimit?: number;
  decodeIds?: boolean;
  decodeEnums?: boolean;
  decodeTimestamps?: boolean;
}

export interface ChotkiSecureJsonData {
  authToken?: string;
  tlsCACert?: string;
  tlsClientCert?: string;
  tlsClientKey?: string;
}
