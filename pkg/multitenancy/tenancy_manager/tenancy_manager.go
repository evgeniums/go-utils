package tenancy_manager

import (
	"errors"
	"sync"

	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/crypt_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/customer"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pool_pubsub"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_subscriber"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type TenancyNotificationHandler struct {
	pubsub_subscriber.SubscriberClientBase
	manager *TenancyManager
}

func (t *TenancyNotificationHandler) Handle(ctx op_context.Context, msg *multitenancy.PubsubNotification) error {

	c := ctx.TraceInMethod("TenancyNotificationHandler.Handle")
	defer ctx.TraceOutMethod()

	t.manager.UnloadTenancy(msg.Tenancy)
	if msg.Operation != multitenancy.OpDelete {
		_, err := t.manager.LoadTenancy(ctx, msg.Tenancy)
		if err != nil {
			return c.SetError(err)
		}
	}

	return nil
}

type TenancyManagerConfig struct {
	MULTITENANCY bool
	DB_PREFIX    string `validate:"required,alphanum" vmessage:"Invalid prefix for names of databases" default:"tenancy"`
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
	PubsubTopic                *multitenancy.PubsubTopic
	PoolPubsub                 pool_pubsub.PoolPubsub
	tenancyNotificationHandler *TenancyNotificationHandler

	selfTopicSubscription   string
	poolTopicsSubscriptions map[string]string

	tenancyDbModels *multitenancy.TenancyDbModels
}

func NewTenancyManager(pools pool.PoolStore, poolPubsub pool_pubsub.PoolPubsub, tenancyDbModels *multitenancy.TenancyDbModels) *TenancyManager {
	m := &TenancyManager{}
	m.Pools = pools
	m.tenanciesById = make(map[string]multitenancy.Tenancy)
	m.tenanciesByPath = make(map[string]multitenancy.Tenancy)
	m.tenancyDbModels = tenancyDbModels
	m.PoolPubsub = poolPubsub
	m.PubsubTopic = &multitenancy.PubsubTopic{}

	m.tenancyNotificationHandler = &TenancyNotificationHandler{manager: m}
	m.tenancyNotificationHandler.Init("tenancy_manager")

	return m
}

func (t *TenancyManager) Config() interface{} {
	return &t.TenancyManagerConfig
}

func (t *TenancyManager) SetController(controller multitenancy.TenancyController) {
	t.Controller = controller
}

func (t *TenancyManager) SetCustomerController(controller customer.CustomerController) {
	t.Customers = controller
}

func (t *TenancyManager) Init(ctx op_context.Context, configPath ...string) error {

	c := ctx.TraceInMethod("TenancyManager.Init")
	defer ctx.TraceOutMethod()

	// init manager
	app := ctx.App()
	cfg := app.Cfg()
	log := app.Logger()
	vld := app.Validator()
	err := object_config.LoadLogValidate(cfg, log, vld, t, "multitenancy", configPath...)
	if err != nil {
		c.SetError(err)
		return ctx.Logger().PushFatalStack("failed to init tenancy manager", err)
	}

	// get self pool
	selfPool, selfPoolErr := t.Pools.SelfPool()

	// subscribe to pubsub notifications
	t.PubsubTopic.TopicBase = pubsub_subscriber.New(multitenancy.PubsubTopicName, multitenancy.NewPubsubNotification)
	if selfPoolErr == nil && selfPool != nil {
		// subscribe to notifications only from self pool
		if selfPool.IsActive() {
			t.selfTopicSubscription, err = t.PoolPubsub.SubscribeSelfPool(ctx, t.PubsubTopic)
			if err != nil {
				c.SetError(err)
				return ctx.Logger().PushFatalStack("failed to subscribe to pubsub notifications in self pool", err)
			}
		}
	} else {
		// subscribe to notifications from all pools
		t.poolTopicsSubscriptions, err = t.PoolPubsub.SubscribePools(ctx, t.PubsubTopic)
		if err != nil {
			c.SetError(err)
			return ctx.Logger().PushFatalStack("failed to subscribe to pubsub notifications in all pools", err)
		}
	}
	t.PubsubTopic.Subscribe(t.tenancyNotificationHandler)

	// load tenancies
	err = t.LoadTenancies(ctx, selfPool)
	if err != nil {
		c.SetError(err)
		return ctx.Logger().PushFatalStack("failed to load tenancies", err)
	}

	// done
	return nil
}

func (t *TenancyManager) Close() {

	t.mutex.Lock()

	for _, tenancy := range t.tenanciesById {
		tenancy.Db().Close()
	}
	t.tenanciesById = make(map[string]multitenancy.Tenancy)
	t.tenanciesByPath = make(map[string]multitenancy.Tenancy)

	t.mutex.Unlock()

	if t.PubsubTopic != nil {
		t.PoolPubsub.UnsubscribePools(t.PubsubTopic.Name())
		t.PoolPubsub.UnsubscribeSelfPool(t.PubsubTopic.Name())
	}
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

func (t *TenancyManager) Tenancies() []multitenancy.Tenancy {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	tenancies := utils.AllMapValues(t.tenanciesById)
	return tenancies
}

func (t *TenancyManager) UnloadTenancy(id string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	tenancy, ok := t.tenanciesById[id]
	if ok {
		tenancy.Db().Close()
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
	skip, err := tenancy.Init(ctx, tenancyDb)
	if err != nil {
		return nil, err
	}
	if skip {
		return nil, nil
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
		if ctx.GenericError() != nil && ctx.GenericError().Code() == generic_error.ErrorCodeNotFound {
			ctx.ClearError()
			// try to find by login
			owner, err = t.Customers.FindByLogin(ctx, id)
		}
	}
	if err != nil {
		c.SetMessage("failed to find customer")
		return nil, err
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
	p, err := t.FindPool(ctx, c, data.POOL_ID)
	if err != nil {
		return nil, err
	}
	if !p.IsActive() {
		ctx.SetGenericErrorCode(pool.ErrorCodePoolNotActive)
		err := errors.New("pool not active")
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
	tenancy.POOL_ID = p.GetID()
	tenancy.TenancyBaseData.Pool = p
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
	err = tenancy.ConnectDatabase(ctx, true)
	if err != nil {
		return nil, err
	}
	defer tenancy.Db().Close()

	// create database
	err = multitenancy.UpgradeTenancyDatabase(ctx, tenancy, t.tenancyDbModels)
	if err != nil {
		ctx.SetGenericErrorCode(multitenancy.ErrorCodeTenancyDbInitializationFailed)
		c.SetMessage("failed to initialize database models")
		return nil, err
	}

	// check if there are any tenancy metas in database
	var metas []multitenancy.TenancyMeta
	_, err = tenancy.Db().FindWithFilter(ctx, nil, &metas)
	if err != nil {
		c.SetMessage("failed to check tenancy metas in created database")
		return nil, err
	}
	if len(metas) != 0 {
		// database already belongs to some tenancy
		if metas[0].GetID() != tenancy.GetID() {
			ctx.SetGenericErrorCode(multitenancy.ErrorCodeForeignDatabase)
			err = errors.New("created database already belongs to other tenancy")
			return nil, err
		}
	} else {
		// set tenancy meta in the database
		meta := &multitenancy.TenancyMeta{}
		meta.ObjectBase = tenancy.ObjectBase
		err = tenancy.Db().Create(ctx, meta)
		if err != nil {
			c.SetMessage("failed to save tenancy meta in created database")
			return nil, err
		}
	}

	// create tenancy item
	item := &multitenancy.TenancyItem{}
	item.TenancyDb = tenancy.TenancyDb
	item.CustomerLogin = customer.Login()
	item.PoolName = p.Name()

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

	// load tenancies
	filter := db.NewFilter()
	if selfPool != nil {
		// load tenancies only for self pool
		filter.AddField("pool_id", selfPool.GetID())
	}
	tenancies, _, err := t.Controller.List(ctx, filter)
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

func (t *TenancyManager) TenancyController() multitenancy.TenancyController {
	return t.Controller
}

// TODO subscribe to customer blocking
