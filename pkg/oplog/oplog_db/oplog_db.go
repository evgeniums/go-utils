package oplog_db

import (
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/oplog"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type OplogControllerDb struct {
	Ctx op_context.Context
}

func (o *OplogControllerDb) Write(op oplog.Oplog) error {
	err := db.DB(o.Ctx.Db()).Create(o.Ctx, op)
	if err != nil {
		o.Ctx.Logger().Error("failed to write oplog", err, logger.Fields{"oplog": utils.ObjectTypeName(op)})
	}
	return err
}

func (o *OplogControllerDb) Read(filter *db.Filter, docs interface{}) (int64, error) {
	count, err := db.DB(o.Ctx.Db()).FindWithFilter(o.Ctx, filter, docs)
	if err != nil {
		o.Ctx.Logger().Error("failed to read oplog", err, logger.Fields{"oplog": utils.ObjectTypeName(docs)})
	}
	return count, err
}

func MakeOplogController(ctx op_context.Context) oplog.OplogController {
	return &OplogControllerDb{Ctx: ctx}
}
