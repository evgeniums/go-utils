package tenancy_manager

import (
	"errors"
	"sync"

	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/crypt_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/customer"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_subscriber"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const (
	TENANCY_DATABASE_ROLE string = "tenancy_db"
)

type TenancyManagerConfig struct {
	MULTITENANCY bool
	DB_PREFIX    string `validate:"required,aphanum" vmessage:"Invalid prefix for names of databases"`
}

func (t *TenancyManagerConfig) IsMultiTenancy() bool {
	return t.MULTITENANCY
}

type TenancyManager struct {
	TenancyManagerConfig
	mutex                      sync.Mutex
	tenanciesById              map[string]multitenancy.Tenancy
	tenanciesByPath            map[string]multitenancy.Tenancy
	Controller                 multitenancy.TenancyController
	Pools                      pool.PoolStore
	Customers                  customer.CustomerController
	DbModels                   []interface{}
	PubsubTopic                multitenancy.PubsubTopic
	tenancyNotificationHandler *TenancyNotificationHandler
}

func NewTenancyManager(subscriber pubsub_subscriber.Subscriber, pools pool.PoolStore, controller multitenancy.TenancyController, dbModels []interface{}) *TenancyManager {
	m := &TenancyManager{}
	m.Pools = pools
	m.Controller = controller
	m.tenanciesById = make(map[string]multitenancy.Tenancy)
	m.tenanciesByPath = make(map[string]multitenancy.Tenancy)
	m.DbModels = dbModels

	m.PubsubTopic.TopicBase = pubsub_subscriber.New(multitenancy.PubsubTopicName, multitenancy.NewPubsubNotification)
	subscriber.Subscribe(&m.PubsubTopic)

	m.tenancyNotificationHandler = &TenancyNotificationHandler{manager: m}
	m.tenancyNotificationHandler.Init("tenancy_manager")
	m.PubsubTopic.Subscribe(m.tenancyNotificationHandler)

	return m
}

func (t *TenancyManager) Config() interface{} {
	return &t.TenancyManagerConfig
}

func (t *TenancyManager) Init(ctx op_context.Context, configPath ...string) error {

	c := ctx.TraceInMethod("TenancyManager.Init")
	defer ctx.TraceOutMethod()

	// init manager
	err := object_config.LoadLogValidate(ctx.App().Cfg(), ctx.Logger(), ctx.App().Validator(), t, "multitenancy", configPath...)
	if err != nil {
		c.SetError(err)
		return ctx.Logger().PushFatalStack("failed to init tenancy manager", err)
	}

	// subscribe to tenancy notifications

	// load tenancies
	err = t.LoadTenancies(ctx)
	if err != nil {
		c.SetError(err)
		return ctx.Logger().PushFatalStack("failed to load tenancies", err)
	}

	// done
	return nil
}

func (t *TenancyManager) Tenancy(id string) (multitenancy.Tenancy, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	tenancy, ok := t.tenanciesById[id]
	if !ok {
		return nil, errors.New("unknown tenancy")
	}
	return tenancy, nil
}

func (t *TenancyManager) TenancyByPath(path string) (multitenancy.Tenancy, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	tenancy, ok := t.tenanciesByPath[path]
	if !ok {
		return nil, errors.New("tenancy not found")
	}
	return tenancy, nil
}

func (t *TenancyManager) UnloadTenancy(id string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	tenancy, ok := t.tenanciesById[id]
	if ok {
		delete(t.tenanciesById, id)
		delete(t.tenanciesByPath, tenancy.Path())
	}
}

func (t *TenancyManager) LoadTenancyFromData(ctx op_context.Context, tenancyDb *multitenancy.TenancyDb) (multitenancy.Tenancy, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("TenancyManager.LoadTenancyFromData", logger.Fields{"tenancy": tenancyDb.GetID(), "customer": tenancyDb.CUSTOMER, "role": tenancyDb.ROLE})
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// TODO use tenancy builder to support derived tenancy types
	// init tenancy
	tenancy := NewTenancy(t)
	err = tenancy.Init(ctx, tenancyDb)
	if err != nil {
		return nil, err
	}

	// keep it
	t.mutex.Lock()
	t.tenanciesById[tenancy.GetID()] = tenancy
	t.tenanciesByPath[tenancy.Path()] = tenancy
	t.mutex.Unlock()

	// done
	return tenancy, nil
}

func (t *TenancyManager) LoadTenancy(ctx op_context.Context, id string) (multitenancy.Tenancy, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("TenancyManager.LoadTenancy", logger.Fields{"tenancy": id})
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// load from database
	tenancyDb, err := t.Controller.Find(ctx, id)
	if err != nil {
		c.SetMessage("failed to find tenancy")
		return nil, err
	}
	if tenancyDb == nil {
		err := errors.New("tenancy not found")
		return nil, err
	}

	// init tenancy
	tenancy, err := t.LoadTenancyFromData(ctx, tenancyDb)
	if err != nil {
		return nil, err
	}

	// done
	return tenancy, nil
}

func (t *TenancyManager) CreateTenancy(ctx op_context.Context, data *multitenancy.TenancyData) (*multitenancy.TenancyDb, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("TenancyManager.LoadTenancy", logger.Fields{"customer": data.CUSTOMER, "role": data.ROLE})
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// check if customer exists
	owner, err := t.Customers.Find(ctx, data.CUSTOMER)
	if err != nil {
		c.SetMessage("failed to find customer")
		return nil, err
	}
	if owner == nil {
		// try to find by login
		owner, err = t.Customers.FindByLogin(ctx, data.CUSTOMER)
		if err != nil {
			c.SetMessage("failed to find customer")
			return nil, err
		}
	}
	if owner == nil {
		err = errors.New("customer not found")
		// TODO load customer errors in tenancies service
		ctx.SetGenericErrorCode(customer.ErrorCodeCustomerNotFound, true)
		return nil, err
	}

	// check if pool exists
	_, err = t.Pools.Pool(data.POOL_ID)
	if err != nil {
		ctx.SetGenericErrorCode(pool.ErrorCodePoolNotFound)
		c.SetMessage("unknown pool")
		return nil, err
	}

	// create
	tenancy := NewTenancy(t)
	tenancy.InitObject()
	tenancy.TenancyData = *data
	tenancy.CUSTOMER = owner.GetID()
	if tenancy.PATH == "" {
		tenancy.PATH = crypt_utils.GenerateString()
	}
	if tenancy.DBNAME == "" {
		tenancy.DBNAME = utils.ConcatStrings(t.DB_PREFIX, "_", owner.Login(), "_", data.ROLE)
	}

	// connect to database server
	err = tenancy.ConnectDatabase(ctx)
	if err != nil {
		return nil, err
	}

	// create database
	err = tenancy.Db().AutoMigrate(ctx, t.DbModels)
	if err != nil {
		genErr := generic_error.New(pool.ErrorCodeServiceInitializationFailed, "Failed to init models of tenancy database")
		ctx.SetGenericError(genErr)
		err = genErr
		return nil, err
	}

	// done
	return &tenancy.TenancyDb, nil
}

func (t *TenancyManager) LoadTenancies(ctx op_context.Context) error {

	// setup
	var err error
	c := ctx.TraceInMethod("TenancyManager.LoadTenancies")
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// find tenancies
	tenancies, _, err := t.Controller.List(ctx, nil)
	if err != nil {
		c.SetMessage("failed to load tenancies from database")
		return err
	}

	// load each tenancy
	for _, tenancy := range tenancies {
		_, err = t.LoadTenancyFromData(ctx, tenancy)
		if err != nil {
			return err
		}
	}

	// done
	return nil
}

func (t *TenancyManager) MigrateDatabase(ctx op_context.Context) error {

	c := ctx.TraceInMethod("TenancyManager.MigrateDatabase")
	defer ctx.TraceOutMethod()

	for _, tenancy := range t.tenanciesById {
		err := tenancy.Db().AutoMigrate(ctx, t.DbModels)
		if err != nil {
			c.SetLoggerField("tenancy", tenancy.Display())
			c.SetLoggerField("tenancy_id", tenancy.GetID())
			return c.SetError(err)
		}
	}

	return nil
}

type TenancyNotificationHandler struct {
	pubsub_subscriber.SubscriberClientBase
	manager *TenancyManager
}

func (t *TenancyNotificationHandler) Handle(ctx op_context.Context, msg *multitenancy.PubsubNotification) error {

	c := ctx.TraceInMethod("TenancyNotificationHandler.Handle")
	defer ctx.TraceOutMethod()

	if msg.Operation == multitenancy.OpDelete {
		t.manager.UnloadTenancy(msg.Tenancy)
	} else {
		_, err := t.manager.LoadTenancy(ctx, msg.Tenancy)
		if err != nil {
			return c.SetError(err)
		}
	}

	return nil
}

// TODO subscribe to customer blocking
