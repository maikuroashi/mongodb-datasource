package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var queryRegex = regexp.MustCompile("(\\w+)\\.(\\w+)\\.(find|aggregate)\\s*\\(\\s*([^)]*)\\s*\\)(?:\\.sort\\(\\s*(.+)\\s*\\))?")
var dateLiteralRegex = regexp.MustCompile("^Date<(\\d+)>$")

type QueryService struct {
	mongoClient *mongo.Client
	defaultDB   string
	maxResult   int
}

type mongoQuery struct {
	Database   string
	Collection string
	Method     string
	Query      interface{}
	Projection interface{}
	Sort       interface{}
}

type DataHandler = func(primitive.D)

func NewQueryService(ctx context.Context, url string, defaultDB string, user string, password string, maxResult int) (*QueryService, error) {

	clientOptions := options.Client()
	clientOptions.ApplyURI(url)
	clientOptions.Auth = &options.Credential{Username: user, Password: password}

	client, err := mongo.Connect(ctx, clientOptions)

	return &QueryService{client, defaultDB, maxResult}, err
}

func (qs *QueryService) Dispose() {
	qs.mongoClient.Disconnect(context.Background())
}

func (qs *QueryService) Ping(ctx context.Context) error {
	return qs.mongoClient.Ping(ctx, nil)
}

func (qs *QueryService) RunQuery(ctx context.Context, queryString string, handler DataHandler) error {

	mongoQuery, err := parseQuery(queryString, qs.defaultDB)
	if err != nil {
		return err
	}

	var cur *mongo.Cursor
	switch mongoQuery.Method {
	case "find":
		cur, err = qs.find(ctx, mongoQuery)
	case "aggregate":
		cur, err = qs.aggregate(ctx, mongoQuery)
	default:
	}

	if err != nil {
		return err
	}

	count := 0
	defer cur.Close(ctx)
	for cur.Next(ctx) {

		count++
		if qs.maxResult > 0 && count > qs.maxResult {
			break
		}

		var rec primitive.D
		err := cur.Decode(&rec)
		if err != nil {
			return err
		}
		handler(rec)
	}

	err = cur.Err()
	return err
}

func (qs *QueryService) find(ctx context.Context, mongoQuery *mongoQuery) (*mongo.Cursor, error) {

	collection := qs.mongoClient.Database(mongoQuery.Database).Collection(mongoQuery.Collection)
	queryOptions := options.FindOptions{Projection: mongoQuery.Projection, Sort: mongoQuery.Sort}
	return collection.Find(ctx, mongoQuery.Query, &queryOptions)
}

func (qs *QueryService) aggregate(ctx context.Context, mongoQuery *mongoQuery) (*mongo.Cursor, error) {

	queryDoc, ok := mongoQuery.Query.(primitive.A)
	if !ok {
		return nil, errors.New("aggregate argument must be an array")
	}

	if mongoQuery.Sort != nil {
		length := len(queryDoc)
		queryDoc = make(primitive.A, length, length+1)
		queryDoc = append(queryDoc, bson.M{"$sort": mongoQuery.Sort})
	}

	collection := qs.mongoClient.Database(mongoQuery.Database).Collection(mongoQuery.Collection)
	return collection.Aggregate(ctx, queryDoc)
}

func parseQuery(queryString string, defaultDB string) (*mongoQuery, error) {

	match := queryRegex.FindStringSubmatch(queryString)
	if match == nil {
		err := fmt.Errorf("'%s' is not a valid MongoDB query expression", queryString)
		return nil, err
	}

	db := match[1]
	if db == "db" {
		db = defaultDB
	}

	collection := match[2]
	method := match[3]

	if method == "find" && match[4] == "" {
		match[4] = "{}"
	} else if method == "aggregate" && match[4] == "" {
		match[4] = "[]"
	}

	filter, projection, err := decodeArgs(match[4])
	if err != nil {
		err = fmt.Errorf("The args '%s' to the '%s' method are not a valid", match[4], match[3])
		return nil, err
	}
	patchLiterals(&filter)

	sort, _, err := decodeArgs(match[5])
	if err != nil {
		err = fmt.Errorf("The args '%s' to the '%s' method are not a valid", match[5], "sort")
		return nil, err
	}

	return &mongoQuery{
		Database:   db,
		Collection: collection,
		Method:     method,
		Query:      filter,
		Projection: projection,
		Sort:       sort,
	}, err
}

func decodeArgs(args string) (interface{}, interface{}, error) {

	jsonArray := fmt.Sprintf("[%s]", args)

	var result []interface{}
	var first interface{}
	var second interface{}
	err := bson.UnmarshalExtJSON([]byte(jsonArray), true, &result)

	if err == nil {
		if len(result) > 0 {
			first = result[0]
		}

		if len(result) > 1 {
			second = result[1]
		}
	}
	return first, second, err
}

func patchLiterals(value *interface{}) {

	switch (*value).(type) {

	case primitive.D:
		pd := (*value).(primitive.D)
		for i := range pd {

			literal, ok := pd[i].Value.(string)
			if ok {
				pd[i].Value = convertLiteral(literal)
			} else {
				patchLiterals(&pd[i].Value)
			}
		}

	case primitive.A:
		pa := (*value).(primitive.A)
		for i := range pa {

			literal, ok := pa[i].(string)
			if ok {
				pa[i] = convertLiteral(literal)
			} else {
				patchLiterals(&pa[i])
			}
		}
	}
}

func convertLiteral(value string) interface{} {

	var result interface{} = value

	match := dateLiteralRegex.FindStringSubmatch(value)
	if match != nil {
		unixTime, _ := strconv.ParseInt(match[1], 10, 64)
		result = primitive.DateTime(unixTime)
	}
	return result
}
