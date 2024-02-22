package mongodb_client

import (
	"context"
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDbClientConfig struct {
	MONGODB_URI string        `validate:"required" vmessage:"mongodb URI must be specified"`
	DB_NAME     string        `validate:"required" vmessage:"Name of database must be specified"`
	JSON_TAG    bool          `default:"true"`
	TIMEOUT     time.Duration `default:"30s"`
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
	if app.Cfg().IsSet("mongodb.json_tag") {
		m.JSON_TAG = app.Cfg().GetBool("mongodb.json_tag")
	}

	// load configuration
	err := object_config.LoadLogValidateApp(app, m, "mongodb_client", configPath...)
	if err != nil {
		return app.Logger().PushFatalStack("failed to load config of mongodb client", err)
	}

	// init mongodb client
	bsonOpts := &options.BSONOptions{
		UseJSONStructTags: m.JSON_TAG,
	}
	context, cancel := context.WithTimeout(context.TODO(), time.Duration(m.TIMEOUT))
	defer cancel()
	m.client, err = mongo.Connect(context, options.Client().ApplyURI(m.MONGODB_URI).SetBSONOptions(bsonOpts))
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

func (m *MongoDbClient) Context() context.Context {
	// TODO use context with deadline
	// ctx, _ := context.WithTimeout(context.TODO(), time.Duration(m.TIMEOUT))
	return context.TODO()
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
