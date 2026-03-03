import { DataSourceInstanceSettings, CoreApp, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';

import { ChotkiDataSourceOptions, ChotkiQuery, DEFAULT_QUERY, QueryParamValue } from './types';

export class DataSource extends DataSourceWithBackend<ChotkiQuery, ChotkiDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<ChotkiDataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<ChotkiQuery> {
    return DEFAULT_QUERY;
  }

  applyTemplateVariables(query: ChotkiQuery, scopedVars: ScopedVars) {
    const replaceString = (value: string) => getTemplateSrv().replace(value, scopedVars);
    const replaceParamValue = (value: QueryParamValue): QueryParamValue => {
      if (typeof value === 'string') {
        return replaceString(value);
      }
      if (Array.isArray(value)) {
        if (value.every((item) => typeof item === 'string')) {
          return value.map((item) => replaceString(item));
        }
        return value;
      }
      return value;
    };

    const params = Object.fromEntries(
      Object.entries(query.params ?? {}).map(([key, value]) => [key, replaceParamValue(value)])
    );

    return {
      ...query,
      params,
      rawQuery: query.rawQuery ? replaceString(query.rawQuery) : query.rawQuery,
    };
  }

  filterQuery(query: ChotkiQuery): boolean {
    if (query.editorMode === 'raw') {
      return Boolean(query.rawQuery?.trim());
    }
    return Boolean(query.method);
  }
}
