package tenancy_manager

import (
	"errors"
	"sync"

	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/crypt_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/customer"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pool_pubsub"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_subscriber"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const (
	TENANCY_DATABASE_ROLE string = "tenancy_db"
)

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
	PoolPubsub                 pool_pubsub.PoolPubsub
	tenancyNotificationHandler *TenancyNotificationHandler
}

func NewTenancyManager(pools pool.PoolStore, poolPubsub pool_pubsub.PoolPubsub, dbModels []interface{}) *TenancyManager {
	m := &TenancyManager{}
	m.Pools = pools
	m.tenanciesById = make(map[string]multitenancy.Tenancy)
	m.tenanciesByPath = make(map[string]multitenancy.Tenancy)
	m.DbModels = append(dbModels, multitenancy.DbInternalModels()...)
	m.PoolPubsub = poolPubsub

	m.tenancyNotificationHandler = &TenancyNotificationHandler{manager: m}
	m.tenancyNotificationHandler.Init("tenancy_manager")
	m.PubsubTopic.Subscribe(m.tenancyNotificationHandler)

	return m
}

func (t *TenancyManager) Config() interface{} {
	return &t.TenancyManagerConfig
}

func (t *TenancyManager) SetController(controller multitenancy.TenancyController) {
	t.Controller = controller
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

	// subscribe to pubsub notifications and load tenancies only if self pool is defined
	selfPool, selfPoolErr := t.Pools.SelfPool()
	if selfPoolErr == nil {

		// subscribe to notifications
		t.PubsubTopic.TopicBase = pubsub_subscriber.New(multitenancy.PubsubTopicName, multitenancy.NewPubsubNotification)
		err = t.PoolPubsub.SubscribeSelfPool(&t.PubsubTopic)
		if err != nil {
			c.SetError(err)
			return ctx.Logger().PushFatalStack("failed to subscribe to pubsub notifications", err)
		}

		// load tenancies
		err = t.LoadTenancies(ctx, selfPool)
		if err != nil {
			c.SetError(err)
			return ctx.Logger().PushFatalStack("failed to load tenancies", err)
		}
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
	c := ctx.TraceInMethod("TenancyManager.LoadTenancyFromData", logger.Fields{"tenancy": tenancyDb.GetID(), "customer": tenancyDb.CUSTOMER_ID, "role": tenancyDb.ROLE})
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
	tenancyItem, err := t.Controller.Find(ctx, id)
	if err != nil {
		c.SetMessage("failed to find tenancy")
		return nil, err
	}
	tenancyDb := &tenancyItem.TenancyDb
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

func (t *TenancyManager) FindCustomer(ctx op_context.Context, c op_context.CallContext, id string) (*customer.Customer, error) {
	owner, err := t.Customers.Find(ctx, id)
	if err != nil {
		c.SetMessage("failed to find customer")
		return nil, err
	}
	if owner == nil {
		// try to find by login
		owner, err = t.Customers.FindByLogin(ctx, id)
		if err != nil {
			c.SetMessage("failed to find customer")
			return nil, err
		}
	}
	if owner == nil {
		err = errors.New("customer not found")
		ctx.SetGenericErrorCode(customer.ErrorCodeCustomerNotFound, true)
		return nil, err
	}
	return owner, nil
}

func (t *TenancyManager) FindPool(ctx op_context.Context, c op_context.CallContext, id string) (pool.Pool, error) {
	p, err := t.Pools.Pool(id)
	if err != nil {
		p, err = t.Pools.PoolByName(id)
		if err != nil {
			ctx.SetGenericErrorCode(pool.ErrorCodePoolNotFound)
			c.SetMessage("unknown pool")
			return nil, err
		}
	}
	return p, nil
}

func (t *TenancyManager) CheckDuplicateRole(ctx op_context.Context, c op_context.CallContext, customerId string, role string) error {
	c.SetLoggerField("customer_id", customerId)
	c.SetLoggerField("role", role)
	fields := db.Fields{"customer_id": customerId, "role": role}
	exists, err := t.Controller.Exists(ctx, fields)
	if err != nil {
		c.SetMessage("failed to check existence of tenancy")
		return err
	}
	if exists {
		err = errors.New("tenancy already exists with such role for that customer")
		ctx.SetGenericErrorCode(multitenancy.ErrorCodeTenancyConflictRole)
		return err
	}
	return nil
}

func (t *TenancyManager) CheckDuplicatePath(ctx op_context.Context, c op_context.CallContext, poolId string, path string) error {
	c.SetLoggerField("pool_id", poolId)
	c.SetLoggerField("path", path)
	fields := db.Fields{"pool_id": poolId, "path": path}
	exists, err := t.Controller.Exists(ctx, fields)
	if err != nil {
		c.SetMessage("failed to check existence of tenancy")
		return err
	}
	if exists {
		err = errors.New("tenancy already exists with such path in that pool")
		ctx.SetGenericErrorCode(multitenancy.ErrorCodeTenancyConflictPath)
		return err
	}
	return nil
}

func (t *TenancyManager) CreateTenancy(ctx op_context.Context, data *multitenancy.TenancyData) (*multitenancy.TenancyItem, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("TenancyManager.LoadTenancy", logger.Fields{"customer": data.CUSTOMER_ID, "role": data.ROLE})
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// find customer
	customer, err := t.FindCustomer(ctx, c, data.CUSTOMER_ID)
	if err != nil {
		return nil, err
	}

	// check if pool exists
	pool, err := t.FindPool(ctx, c, data.POOL_ID)
	if err != nil {
		return nil, err
	}

	// check if tenancy with such role for this customer exists
	err = t.CheckDuplicateRole(ctx, c, data.CUSTOMER_ID, data.ROLE)
	if err != nil {
		return nil, err
	}

	// create
	tenancy := NewTenancy(t)
	tenancy.InitObject()
	tenancy.TenancyData = *data
	tenancy.CUSTOMER_ID = customer.GetID()
	tenancy.POOL_ID = pool.GetID()
	if tenancy.PATH == "" {
		tenancy.PATH = crypt_utils.GenerateString()
	}
	if tenancy.DBNAME == "" {
		tenancy.DBNAME = utils.ConcatStrings(t.DB_PREFIX, "_", customer.Login(), "_", data.ROLE)
	}

	// check if tenancy with such path in that pool
	err = t.CheckDuplicatePath(ctx, c, data.POOL_ID, data.PATH)
	if err != nil {
		return nil, err
	}

	// connect to database server
	err = tenancy.ConnectDatabase(ctx)
	if err != nil {
		return nil, err
	}

	// create database
	err = tenancy.Db().AutoMigrate(ctx, t.DbModels)
	if err != nil {
		ctx.SetGenericErrorCode(multitenancy.ErrorCodeTenancyDbInitializationFailed)
		c.SetMessage("failed to initialize database models")
		return nil, err
	}

	// set tenancy meta in the database
	meta := &multitenancy.TenancyMeta{}
	meta.ObjectBase = tenancy.ObjectBase
	err = tenancy.Db().Create(ctx, meta)
	if err != nil {
		c.SetMessage("failed to save tenancy meta in created database")
		return nil, err
	}

	// create item
	item := &multitenancy.TenancyItem{}
	item.TenancyDb = tenancy.TenancyDb
	item.CustomerLogin = customer.Login()

	// done
	return item, nil
}

func (t *TenancyManager) LoadTenancies(ctx op_context.Context, selfPool pool.Pool) error {

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

	// load tenancies only for self pool
	filter := db.NewFilter()
	filter.AddField("pool_id", selfPool.GetID())
	tenancies, _, err := t.Controller.List(ctx, nil)
	if err != nil {
		c.SetMessage("failed to load tenancies from database")
		return err
	}

	// load each tenancy
	for _, tenancy := range tenancies {
		_, err = t.LoadTenancyFromData(ctx, &tenancy.TenancyDb)
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
			c.SetLoggerField("tenancy", multitenancy.TenancyDisplay(tenancy))
			c.SetLoggerField("tenancy_id", tenancy.GetID())
			return c.SetError(err)
		}
	}

	return nil
}

// TODO subscribe to customer blocking
