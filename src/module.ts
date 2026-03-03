import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { ChotkiDataSourceOptions, ChotkiQuery } from './types';

export const plugin = new DataSourcePlugin<DataSource, ChotkiQuery, ChotkiDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
