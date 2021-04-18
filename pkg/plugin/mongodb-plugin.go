package plugin

import (
	"context"
	"encoding/json"

	"github.com/maikuroashi/mongodb-datasource/pkg/field"
	"github.com/maikuroashi/mongodb-datasource/pkg/query"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// newDatasource returns datasource.ServeOpts.
func NewDatasource() datasource.ServeOpts {
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

	instance, err := td.pluginInstance(req.PluginContext)
	if err != nil {
		return nil, err
	}

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := instance.query(ctx, q)

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

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (td *MongoDBDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {

	instance, err := td.pluginInstance(req.PluginContext)
	if err == nil {
		err = instance.queryService.Ping(ctx)
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

func (td *MongoDBDatasource) pluginInstance(pluginContext backend.PluginContext) (*pluginInstance, error) {

	instance, err := td.im.Get(pluginContext)
	return instance.(*pluginInstance), err
}

type pluginInstance struct {
	queryService query.QueryService
	maxResult    int
}

func newDataSourceInstance(setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {

	url := setting.URL
	user := setting.User
	defaultDB := setting.Database

	secureData := setting.DecryptedSecureJSONData
	password := secureData["password"]

	customSettings := unmarshalCustomSettings(setting.JSONData)
	value, ok := customSettings["maxResults"]
	maxResult := 1000
	if ok {
		maxResult = int(value.(float64))
	}

	queryService, err := query.NewQueryService(context.Background(), url, defaultDB, user, password)
	if err != nil {
		return nil, err
	}

	return &pluginInstance{
		queryService: queryService,
		maxResult:    maxResult,
	}, err
}

func (is *pluginInstance) query(ctx context.Context, query backend.DataQuery) backend.DataResponse {

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

	ds := field.NewFieldBuilder(10)
	response.Error = is.queryService.RunQuery(ctx, qm.QueryText, is.maxResult, ds.ProcessRecord)
	if response.Error != nil {
		return response
	}

	// create data frame response
	frame := data.NewFrame("response", ds.BuildFields()...)
	response.Frames = append(response.Frames, frame)

	return response
}

func (is *pluginInstance) Dispose() {
	is.queryService.Disconnect(context.Background())
}

func unmarshalCustomSettings(raw json.RawMessage) map[string]interface{} {

	var jsonSettings interface{}
	err := json.Unmarshal(raw, &jsonSettings)
	if err != nil {
		return make(map[string]interface{})
	}
	return jsonSettings.(map[string]interface{})
}
