package oplog_db

import (
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/oplog"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type OplogConteollerDb struct {
	Ctx op_context.Context
}

func (o *OplogConteollerDb) Write(op oplog.Oplog) error {
	err := db.DB(o.Ctx.Db()).Create(o.Ctx, op)
	if err != nil {
		o.Ctx.Logger().Error("failed to write oplog", err, logger.Fields{"oplog": utils.ObjectTypeName(op)})
	}
	return err
}

func (o *OplogConteollerDb) Read(filter *db.Filter, docs interface{}) (int64, error) {
	count, err := db.DB(o.Ctx.Db()).FindWithFilter(o.Ctx, filter, docs)
	if err != nil {
		o.Ctx.Logger().Error("failed to read oplog", err, logger.Fields{"oplog": utils.ObjectTypeName(docs)})
	}
	return count, err
}

func MakeOplogController(ctx op_context.Context) oplog.OplogController {
	return &OplogConteollerDb{Ctx: ctx}
}
