package mongodb_client

import (
	"context"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDbClientConfig struct {
	MONGODB_URI string `validate:"required" vmessage:"mongodb URI must be specified"`
	DB_NAME     string `validate:"required" vmessage:"Name of database must be specified"`
}

type MongoDbClient struct {
	MongoDbClientConfig
	client *mongo.Client
	name   string
}

func New(name string, defaultDbName ...string) *MongoDbClient {
	m := &MongoDbClient{name: name}
	m.DB_NAME = utils.OptionalString("", defaultDbName...)
	return m
}

func (m *MongoDbClient) Config() interface{} {
	return &m.MongoDbClientConfig
}

func (m *MongoDbClient) Init(app app_context.Context, configPath ...string) error {

	if app.Cfg().IsSet("mongodb.mongodb_uri") {
		m.MONGODB_URI = app.Cfg().GetString("mongodb.mongodb_uri")
	}
	if app.Cfg().IsSet("mongodb.db_name") {
		m.DB_NAME = app.Cfg().GetString("mongodb.db_name")
	}

	// load configuration
	err := object_config.LoadLogValidateApp(app, m, "mongodb_client", configPath...)
	if err != nil {
		return app.Logger().PushFatalStack("failed to load config of mongodb client", err)
	}

	// init mongodb client
	m.client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(m.MONGODB_URI))
	if err != nil {
		return app.Logger().PushFatalStack("failed to connect to mongodb server", err)
	}

	// done
	app.Logger().Info("Connected to mongodb", logger.Fields{"mongodb_client": m.name})
	return nil
}

func (m *MongoDbClient) Client() *mongo.Client {
	return m.client
}

func (m *MongoDbClient) DbName() string {
	return m.DB_NAME
}

func (m *MongoDbClient) Name() string {
	return m.name
}

func (m *MongoDbClient) Shutdown(ctx context.Context) error {
	if m.client != nil {
		return m.client.Disconnect(ctx)
	}
	return nil
}
