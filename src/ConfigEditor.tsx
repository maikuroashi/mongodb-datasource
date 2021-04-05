import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MongoDBDataSourceOptions, MongoDBSecureJsonData } from './types';

const { SecretFormField, FormField } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<MongoDBDataSourceOptions> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onMaxResultsChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      maxResults: parseInt(event.target.value, 10),
    };
    onOptionsChange({ ...options, jsonData });
  };
  onUrlChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({ ...options, url: event.target.value });
  };
  onDatabaseChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({ ...options, database: event.target.value });
  };
  onUserChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({ ...options, user: event.target.value });
  };

  // Secure field (only sent to the backend)
  onPasswordChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        password: event.target.value,
      },
    });
  };

  onResetPassword = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        password: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        password: '',
      },
    });
  };

  render() {
    const { options } = this.props;
    const { jsonData, secureJsonFields } = options;
    const secureJsonData = (options.secureJsonData || {}) as MongoDBSecureJsonData;

    return (
      <div>
        <h3 className="page-heading">MongoDB Connection</h3>
        <div className="gf-form-group">
          <div className="gf-form">
            <FormField
              label="URL"
              labelWidth={6}
              inputWidth={20}
              onChange={this.onUrlChange}
              value={options.url || ''}
              placeholder="mongodb://localhost:27017"
            />
          </div>
          <div className="gf-form">
            <FormField
              label="Database"
              labelWidth={6}
              inputWidth={20}
              onChange={this.onDatabaseChange}
              value={options.database || ''}
              placeholder="database name"
            />
          </div>
          <div className="gf-form-inline">
            <div className="gf-form">
              <FormField
                label="User"
                labelWidth={6}
                inputWidth={20}
                onChange={this.onUserChange}
                value={options.user || ''}
                placeholder="user"
              />
            </div>
            <div className="gf-form">
              <SecretFormField
                isConfigured={(secureJsonFields && secureJsonFields.password) as boolean}
                value={secureJsonData.password || ''}
                label="Password"
                placeholder="password"
                labelWidth={6}
                inputWidth={20}
                onReset={this.onResetPassword}
                onChange={this.onPasswordChange}
              />
            </div>
          </div>
        </div>
        <h3 className="page-heading">Query Limits</h3>
        <div className="gf-form-group">
          <div className="gf-form">
            <FormField
              label="Max Results"
              labelWidth={6}
              inputWidth={20}
              onChange={this.onMaxResultsChange}
              value={jsonData.maxResults || ''}
              placeholder="1000"
            />
          </div>
        </div>
      </div>
    );
  }
}
