import { DataSourceInstanceSettings } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { MongoDBDataSourceOptions, MongoDBQuery } from './types';

export class DataSource extends DataSourceWithBackend<MongoDBQuery, MongoDBDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MongoDBDataSourceOptions>) {
    super(instanceSettings);
  }
  applyTemplateVariables(query: MongoDBQuery) {
    const templateSrv = getTemplateSrv();
    const queryText = query.queryText ? templateSrv.replace(query.queryText) : '';
    return {
      ...query,
      queryText: queryText,
    };
  }
}
