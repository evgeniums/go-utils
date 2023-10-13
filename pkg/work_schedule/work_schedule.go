package work_schedule

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/background_worker"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/crud"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/app_with_multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context/default_op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type Work interface {
	common.Object
	GetReferenceId() string
	SetReferenceId(string)
	GetNextTime() time.Time
	SetNextTime(time.Time)
	IsAcquired() bool
}

type WorkBuilder[T Work] func() T

type WorkPublisher[T Work] func(ctx op_context.Context, work T, immediate bool, tenancy ...multitenancy.Tenancy) error

type WorkProducer[T Work] interface {
	NewWork(referenceId string) T
	PostWork(ctx op_context.Context, work T, immediate bool, tenancy ...multitenancy.Tenancy) error
	RemoveWork(ctx op_context.Context, referenceId string) error
}

type WorkProducerBase[T Work] struct {
	workBuilder WorkBuilder[T]
}

func (s *WorkProducerBase[T]) Construct(workBuilder WorkBuilder[T]) {
	s.workBuilder = workBuilder
}

func (s *WorkProducerBase[T]) NewWork(referenceId string) T {
	w := s.workBuilder()
	w.InitObject()
	w.SetReferenceId(referenceId)
	return w
}

type WorkBase struct {
	common.ObjectBase
	ReferenceId   string    `json:"reference_id" gorm:"uniqueIndex"`
	Acquired      bool      `json:"acquired" gorm:"index"`
	NextTime      time.Time `json:"next_time" gorm:"index"`
	AcquiringTime time.Time `json:"acquiring_time" gorm:"index"`
}

func (w *WorkBase) GetReferenceId() string {
	return w.ReferenceId
}

func (w *WorkBase) SetReferenceId(referenceId string) {
	w.ReferenceId = referenceId
}

func (w *WorkBase) IsAcquired() bool {
	return w.Acquired
}

func (w *WorkBase) GetNextTime() time.Time {
	return w.NextTime
}

func (w *WorkBase) SetNextTime(nextTime time.Time) {
	w.NextTime = nextTime
}

type WorkRunner[T Work] interface {
	Run(ctx op_context.Context, work T) (bool, time.Time, error)
}

type WorkScheduleConfig struct {
	PARALLEL_JOBS                       int `default:"8"`
	BUCKET_SIZE                         int `default:"32"`
	JOB_INVOKATION_INTERVAL_SECONDS     int `default:"300"`
	STUCK_JOBS_RELEASE_INTERVAL_SECONDS int `default:"900"`
}

type workItem[T Work] struct {
	work    T
	tenancy multitenancy.Tenancy
}

type WorkSchedule[T Work] struct {
	WorkScheduleConfig
	app_context.WithAppBase
	crud.WithCRUDBase
	background_worker.JobRunnerBase

	WorkProducerBase[T]

	name string

	mutex            sync.RWMutex
	runningWorks     map[string]bool
	pendingWorkCount atomic.Int32

	workRunner WorkRunner[T]

	queue chan workItem[T]

	lastStuckWorksCleanTime time.Time

	running       atomic.Bool
	workPublisher WorkPublisher[T]
}

type Config[T Work] struct {
	WorkBuilder   WorkBuilder[T]
	WorkRunner    WorkRunner[T]
	WorkPublisher WorkPublisher[T]
}

func New[T Work](name string, config Config[T], cruds ...crud.CRUD) *WorkSchedule[T] {
	s := &WorkSchedule[T]{
		name:       name,
		workRunner: config.WorkRunner,
	}
	s.WorkProducerBase.Construct(config.WorkBuilder)
	s.WithCRUDBase.Construct(cruds...)
	s.runningWorks = make(map[string]bool)
	s.queue = make(chan workItem[T])
	s.workPublisher = config.WorkPublisher
	if s.workPublisher == nil {
		s.workPublisher = func(ctx op_context.Context, work T, immediate bool, tenancy ...multitenancy.Tenancy) error {
			if immediate {
				s.queue <- workItem[T]{work: work, tenancy: utils.OptionalArg(nil, tenancy...)}
			}
			return nil
		}
	}
	return s
}

func (s *WorkSchedule[T]) Config() interface{} {
	return &s.WorkScheduleConfig
}

func (s *WorkSchedule[T]) Init(app app_context.Context, configPath ...string) error {

	s.WithAppBase.Init(app)

	err := object_config.LoadLogValidateApp(app, s, "work_schedule", configPath...)
	if err != nil {
		return app.Logger().PushFatalStack("failed to load configuration of WorkSchedule", err)
	}

	for i := 0; i < s.PARALLEL_JOBS; i++ {
		go s.worker()
	}

	return nil
}

func (s *WorkSchedule[T]) StopJob() {
	close(s.queue)
}

func (s *WorkSchedule[T]) RunJob() {
	s.ProcessWorks()
}

func (s *WorkSchedule[T]) PostWork(ctx op_context.Context, work T, immediate bool, tenancy ...multitenancy.Tenancy) error {

	// setup
	var err error
	c := ctx.TraceInMethod("WorkSchedule.PostWork")
	defer ctx.TraceOutMethod()

	// set next time
	if work.GetNextTime() == utils.TimeNil {
		work.SetNextTime(time.Now().Add(time.Second * time.Duration(s.JOB_INVOKATION_INTERVAL_SECONDS)))
	}

	// save in database
	_, err = s.CRUD().CreateDup(ctx, work, true)
	if err != nil {
		c.SetLoggerField("work_reference_id", work.GetReferenceId())
		c.SetMessage("failed to save work in database")
		return c.SetError(err)
	}

	// add to queue if immediate
	err = s.workPublisher(ctx, work, immediate, tenancy...)
	if err != nil {
		c.SetMessage("failed to publish work")
		return nil
	}

	// done
	return nil
}

func (s *WorkSchedule[T]) RemoveWork(ctx op_context.Context, referenceId string) error {

	// setup
	var err error
	c := ctx.TraceInMethod("WorkSchedule.AddWork")
	defer ctx.TraceOutMethod()

	// delete from database
	err = s.CRUD().DeleteByFields(ctx, db.Fields{"reference_id": referenceId}, s.workBuilder())
	if err != nil {
		c.SetLoggerField("work_reference_id", referenceId)
		c.SetMessage("failed to delete work from database")
		return c.SetError(err)
	}

	// done
	return nil
}

func (s *WorkSchedule[T]) ProcessWorks() {

	if !s.running.CompareAndSwap(false, true) {
		return
	}
	defer s.running.Store(false)

	// TODO support multitenancy

	ctx := default_op_context.BackgroundOpContext(s.App(), s.name)
	defer ctx.Close()
	c := ctx.TraceInMethod("WorkSchedule.RunJob")

	filter := db.NewFilter()

	// prepare filter
	filter.AddField("acquired", false)
	filter.AddInterval("next_time", nil, time.Now())
	filter.SetSorting("next_time", db.SORT_ASC)

	// process works
	for {

		// check if runner is stopped
		if s.Stopper().IsStopped() {
			break
		}

		// release stuck works
		if time.Since(s.lastStuckWorksCleanTime).Seconds() >= float64(s.STUCK_JOBS_RELEASE_INTERVAL_SECONDS) {
			s.lastStuckWorksCleanTime = time.Now()
			f := db.NewFilter()
			beforeTimestamp := time.Now().Add(-time.Second * time.Duration(s.STUCK_JOBS_RELEASE_INTERVAL_SECONDS))
			f.AddInterval("acquiring_time", nil, beforeTimestamp)
			f.AddField("acquired", true)
			e := s.CRUD().UpdateWithFilter(ctx, s.workBuilder(), f, db.Fields{"acquired": false})
			if e != nil {
				c.Logger().Error("failed to release stuck works", e)
			}
		}

		// check number of works currently pending or being processed
		s.mutex.RLock()
		filter.Limit = s.BUCKET_SIZE - len(s.runningWorks) - int(s.pendingWorkCount.Load())
		s.mutex.RUnlock()
		if filter.Limit <= 0 {
			c.Logger().Info("all bucket size is used, skipping")
			break
		}

		// read works from database
		var works []T
		_, err := s.CRUD().List(ctx, filter, &works)
		if err != nil {
			c.SetMessage("failed to read works from database")
			c.SetError(err)
			return
		}

		// stop cycle if there are no works
		if len(works) == 0 || s.Stopper().IsStopped() {
			break
		}

		// enqueue works to workers
		for _, work := range works {
			if s.Stopper().IsStopped() {
				break
			}
			s.pendingWorkCount.Add(1)
			// TODO support multitenancy
			s.queue <- workItem[T]{work: work, tenancy: nil}
		}
	}
}

func (s *WorkSchedule[T]) DoWork(ctx op_context.Context, work T) (bool, error) {

	// setup
	var err error
	ctx.SetLoggerField("work_reference_id", work.GetReferenceId())
	c := ctx.TraceInMethod("WorkSchedule.DoWork")
	defer ctx.TraceOutMethod()

	// check if work is already acquired
	if work.IsAcquired() {
		c.Logger().Warn("work is acquired in input")
		return true, nil
	}

	// check if work is acquired by local process and acquire it by local process
	s.mutex.Lock()
	_, acquired := s.runningWorks[work.GetReferenceId()]
	if acquired {
		s.mutex.Unlock()
		c.Logger().Warn("work is acquired in local map")
		return true, nil
	}
	s.runningWorks[work.GetReferenceId()] = true
	s.mutex.Unlock()
	defer delete(s.runningWorks, work.GetReferenceId())

	// read work from database
	acquiredInDb := false
	acquireWorkInDb := func() error {

		// read work from database
		dbWork := s.workBuilder()
		found, err := s.CRUD().ReadForUpdate(ctx, db.Fields{"reference_id": work.GetReferenceId()}, dbWork)
		if err != nil {
			c.SetMessage("failed to read work from database")
			return err
		}
		if !found {
			// no need to update work in database
			return nil
		}

		// check if work is already acquired
		if work.IsAcquired() {
			acquiredInDb = true
			return nil
		}

		// acquire work
		nextTime := time.Now().Add(time.Second * time.Duration(s.JOB_INVOKATION_INTERVAL_SECONDS))
		err = s.CRUD().Update(ctx, dbWork, db.Fields{"acquired": true, "next_time": nextTime})
		if err != nil {
			c.SetMessage("failed to acquirer work in database")
			return err
		}

		// done
		return nil
	}
	if ctx.DbTransaction() != nil {
		err = acquireWorkInDb()
	} else {
		err = op_context.ExecDbTransaction(ctx, acquireWorkInDb)
	}
	if err != nil {
		return false, c.SetError(err)
	}

	// check if work is already acquired in db
	if acquiredInDb {
		c.Logger().Warn("work is acquired in db")
		return true, nil
	}

	// run work
	done, nextTime, err := s.workRunner.Run(ctx, work)
	updateProcessedWork := func() error {

		// read work from database
		dbWork := s.workBuilder()
		found, err := s.CRUD().ReadForUpdate(ctx, db.Fields{"id": work.GetID()}, dbWork)
		if err != nil {
			c.SetMessage("failed to read processed work from database")
			return err
		}
		if !found {
			// no need to update work in database
			return nil
		}

		// if done then just delete work
		if done {
			err = s.CRUD().Delete(ctx, dbWork)
			if err != nil {
				c.SetMessage("failed to delete processed work from database")
				return err
			}
			return nil
		}

		// release work
		if nextTime == utils.TimeNil {
			nextTime = time.Now().Add(time.Second * time.Duration(s.JOB_INVOKATION_INTERVAL_SECONDS))
		}
		err = s.CRUD().Update(ctx, dbWork, db.Fields{"acquired": false, "next_time": nextTime, "acquiring_time": time.Now()})
		if err != nil {
			c.SetMessage("failed to release work in database")
			return err
		}

		// done
		return nil
	}
	var err1 error
	if ctx.DbTransaction() != nil {
		err1 = updateProcessedWork()
	} else {
		err1 = op_context.ExecDbTransaction(ctx, updateProcessedWork)
	}
	if err1 != nil {
		if err == nil {
			return false, c.SetError(err1)
		}
		c.Logger().Error("failed to update processed work", err1)
		return false, c.SetError(err)
	}

	// done
	return false, nil
}

func (s *WorkSchedule[T]) worker() {
	for work := range s.queue {

		if s.Stopper().IsStopped() {
			break
		}

		if work.tenancy != nil {
			ctx := app_with_multitenancy.BackgroundOpContext(s.App(), work.tenancy, s.name)
			s.DoWork(ctx, work.work)
		} else {
			ctx := default_op_context.BackgroundOpContext(s.App(), s.name)
			s.DoWork(ctx, work.work)
		}
		s.pendingWorkCount.Add(-1)

		go s.ProcessWorks()
	}
}
