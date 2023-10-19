package work_schedule_console

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/console_tool"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/work_schedule"
)

const PostWorkCmd string = "post-work"
const PostWorkDescription string = "Post work"

func PostWork[T work_schedule.Work]() console_tool.Handler[*WorkScheduleCommands[T]] {
	a := &PostWorkHandler[T]{}
	a.Init(PostWorkCmd, PostWorkDescription)
	return a
}

type PostWorkData struct {
	ReferenceType string `long:"reference_type" description:"Work reference type"`
	ReferenceId   string `long:"reference_id" description:"Work reference ID"`
	Mode          string `long:"mode" description:"Posting mode: direct | queued | schedule" validate:"oneof:direct queued schedule" vmessage:"Invalid mode"`
	Delay         int    `long:"delay" description:"Work invokation delay"`
}

type PostWorkHandler[T work_schedule.Work] struct {
	HandlerBase[T]
	PostWorkData
}

func (a *PostWorkHandler[T]) Data() interface{} {
	return &a.PostWorkData
}

func (a *PostWorkHandler[T]) Execute(args []string) error {

	ctx, controller, err := a.Context(a.Data())
	if err != nil {
		return err
	}
	defer ctx.Close()

	work := controller.NewWork(a.ReferenceId, a.ReferenceType)
	err = controller.PostWork(ctx, work, work_schedule.Mode(a.Mode), ctx.GetTenancy())
	if err != nil {
		return err
	}
	fmt.Printf("Posted work:\n%s\n", utils.DumpPrettyJson(work))
	return nil
}
