package main

import (
	"context"
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// newDatasource returns datasource.ServeOpts.
func newDatasource() datasource.ServeOpts {
	// creates a instance manager for your plugin. The function passed
	// into `NewInstanceManger` is called when the instance is created
	// for the first time or when a datasource configuration changed.
	im := datasource.NewInstanceManager(newDataSourceInstance)
	ds := &MongoDBDatasource{
		im: im,
	}

	return datasource.ServeOpts{
		QueryDataHandler:   ds,
		CheckHealthHandler: ds,
	}
}

type MongoDBDatasource struct {
	im instancemgmt.InstanceManager
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *MongoDBDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {

	log.DefaultLogger.Info("QueryData", "request", req)

	queryService, err := td.queryService(req.PluginContext)
	if err != nil {
		return nil, err
	}

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := td.query(ctx, queryService, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}
	return response, err
}

type queryModel struct {
	Format    string `json:"format"`
	QueryText string `json:"queryText"`
}

func (td *MongoDBDatasource) query(ctx context.Context, queryService *QueryService, query backend.DataQuery) backend.DataResponse {

	response := backend.DataResponse{}

	// Unmarshal the json into our queryModel
	var qm queryModel
	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}

	// Log a warning if `Format` is empty.
	if qm.Format == "" {
		log.DefaultLogger.Warn("format is empty. defaulting to time series")
	}

	ds := NewFieldBuilder(10)
	response.Error = queryService.RunQuery(ctx, qm.QueryText, func(record primitive.D) {
		ds.ProcessRecord(record)
	})
	if response.Error != nil {
		return response
	}

	// create data frame response
	frame := data.NewFrame("response", ds.BuildFields()...)
	response.Frames = append(response.Frames, frame)

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (td *MongoDBDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {

	queryService, err := td.queryService(req.PluginContext)
	if err == nil {
		err = queryService.Ping(ctx)
	}

	status := backend.HealthStatusOk
	message := "Data source is working"
	if err != nil {
		status = backend.HealthStatusError
		message = err.Error()
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

func (td *MongoDBDatasource) queryService(pluginContext backend.PluginContext) (*QueryService, error) {

	instance, err := td.im.Get(pluginContext)
	return instance.(*QueryService), err
}

func newDataSourceInstance(setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {

	url := setting.URL
	user := setting.User
	defaultDB := setting.Database

	secureData := setting.DecryptedSecureJSONData
	password, _ := secureData["password"]

	customSettings := unmarshalCustomSettings(setting.JSONData)
	value, ok := customSettings["maxResults"]
	maxResult := 1000
	if ok {
		maxResult = int(value.(float64))
	}

	return NewQueryService(context.Background(), url, defaultDB, user, password, maxResult)
}

func unmarshalCustomSettings(raw json.RawMessage) map[string]interface{} {

	var jsonSettings interface{}
	err := json.Unmarshal(raw, &jsonSettings)
	if err != nil {
		return make(map[string]interface{})
	}
	return jsonSettings.(map[string]interface{})
}
