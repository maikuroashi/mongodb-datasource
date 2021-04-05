import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface MongoDBQuery extends DataQuery {
  queryText: string;
}

export const defaultQuery: Partial<MongoDBQuery> = {
  queryText: 'db.mycollection.find()',
};

/**
 * These are options configured for each DataSource instance
 */
export interface MongoDBDataSourceOptions extends DataSourceJsonData {
  maxResults: number;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MongoDBSecureJsonData {
  password?: string;
}
