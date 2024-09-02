package work_schedule_console

import (
	"errors"

	"github.com/evgeniums/go-utils/pkg/app_context"
	"github.com/evgeniums/go-utils/pkg/console_tool"
	"github.com/evgeniums/go-utils/pkg/multitenancy"
	"github.com/evgeniums/go-utils/pkg/pubsub/pool_pubsub"
	"github.com/evgeniums/go-utils/pkg/work_schedule"
)

type WorkSchedulerBuilder[T work_schedule.Work] func(app app_context.Context) (work_schedule.WorkScheduler[T], error)

type WorkScheduleCommands[T work_schedule.Work] struct {
	console_tool.Commands[*WorkScheduleCommands[T]]

	MakeController WorkSchedulerBuilder[T]
}

func NewWorkScheduleCommands[T work_schedule.Work](name string, description string, makeController WorkSchedulerBuilder[T]) *WorkScheduleCommands[T] {
	w := &WorkScheduleCommands[T]{}
	w.Construct(w, name, description)
	w.MakeController = makeController
	w.LoadHandlers()
	return w
}

func (w *WorkScheduleCommands[T]) LoadHandlers() {
	w.AddHandlers(
		PostWork[T],
	)
}

type HandlerBase[T work_schedule.Work] struct {
	console_tool.HandlerBase[*WorkScheduleCommands[T]]
}

func (b *HandlerBase[T]) Context(data interface{}) (multitenancy.TenancyContext, work_schedule.WorkScheduler[T], error) {

	ctx, err := b.HandlerBase.Context(data)
	if err != nil {
		return ctx, nil, err
	}

	ctrl, err := b.Group.MakeController(ctx.App())
	if err != nil {
		return ctx, nil, err
	}

	return ctx, ctrl, nil
}

func DefaultControllerBuilder[T work_schedule.Work](app app_context.Context, workBuilder work_schedule.WorkBuilder[T], shedulerName string, pubsubTopicName string, configPath string) (work_schedule.WorkScheduler[T], error) {

	a, ok := app.(pool_pubsub.AppWithPubsub)
	if !ok {
		return nil, errors.New("must be application with pool pubsub")
	}

	// create profile work publisher
	workPublisher := work_schedule.NewPoolWorkPublisher[T](a.Pubsub(), pubsubTopicName)

	// create work scheduler
	workScheduler := work_schedule.NewWorkSchedule[T](shedulerName, work_schedule.Config[T]{
		WorkBuilder: workBuilder,
		WorkInvoker: workPublisher.InvokeWork,
	})
	err := workScheduler.Init(app, configPath)
	if err != nil {
		return nil, err
	}

	// done
	return workScheduler, nil
}

func MakeDefaultControllerBuilder[T work_schedule.Work](workBuilder work_schedule.WorkBuilder[T], shedulerName string, pubsubTopicName string, configPath string) WorkSchedulerBuilder[T] {
	return func(app app_context.Context) (work_schedule.WorkScheduler[T], error) {
		return DefaultControllerBuilder(app, workBuilder, shedulerName, pubsubTopicName, configPath)
	}
}
