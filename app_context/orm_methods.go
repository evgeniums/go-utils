package app_context

// func FindByField(ctx *app_context.Context, field string, value string, obj interface{}, tx ...*gorm.DB) (bool, error) {
// 	db := ctx.DB
// 	if len(tx) != 0 {
// 		db = tx[0]
// 	}
// 	notFound, err := orm.FindByField(db, field, value, obj)
// 	if err != nil && !notFound {
// 		ctx.Log.WithFields(log.Fields{"field": field, "value": value, "error": err}).Errorf("%v.FindByField", ObjectTypeName(obj))
// 	}
// 	return notFound, err
// }

// func Find(ctx *app_context.Context, id interface{}, obj interface{}, tx ...*gorm.DB) (bool, error) {
// 	db := ctx.DB
// 	if len(tx) != 0 {
// 		db = tx[0]
// 	}
// 	notFound, err := orm.Find(db, id, obj)
// 	if err != nil && !notFound {
// 		ctx.Log.WithFields(log.Fields{"id": id, "error": err}).Errorf("%v.Find", ObjectTypeName(obj))
// 	}
// 	return notFound, err
// }

// func Delete(ctx *app_context.Context, obj orm.WithIdInterface, tx ...*gorm.DB) error {
// 	db := ctx.DB
// 	if len(tx) != 0 {
// 		db = tx[0]
// 	}
// 	err := orm.RemoveById(db, obj.ID(), obj)
// 	if err != nil {
// 		ctx.Log.WithFields(log.Fields{"id": obj.ID(), "error": err}).Errorf("%v.Delete", ObjectTypeName(obj))
// 	}
// 	return err
// }

// func DeleteByField(ctx *app_context.Context, field string, value interface{}, obj interface{}, tx ...*gorm.DB) error {
// 	db := ctx.DB
// 	if len(tx) != 0 {
// 		db = tx[0]
// 	}
// 	err := orm.RemoveByField(db, field, value, obj)
// 	if err != nil {
// 		ctx.Log.WithFields(log.Fields{"field": field, "value": value, "error": err}).Errorf("%v.DeleteByField", ObjectTypeName(obj))
// 	}
// 	return err
// }

// func DeleteByFields(ctx *app_context.Context, fields map[string]interface{}, obj interface{}, tx ...*gorm.DB) error {
// 	db := ctx.DB
// 	if len(tx) != 0 {
// 		db = tx[0]
// 	}
// 	err := orm.DeleteAllByFields(db, fields, obj)
// 	if err != nil {
// 		ctx.Log.WithFields(log.Fields{"fields": fields, "error": err}).Errorf("%v.DeleteByFields", ObjectTypeName(obj))
// 	}
// 	return err
// }

// func Create(ctx *app_context.Context, obj orm.BaseInterface, tx ...*gorm.DB) error {
// 	db := ctx.DB
// 	if len(tx) != 0 {
// 		db = tx[0]
// 	}
// 	err := orm.Create(db, obj)
// 	if err != nil {
// 		ctx.Log.WithFields(log.Fields{"id": obj.ID(), "error": err}).Errorf("%v.Create", ObjectTypeName(obj))
// 	}
// 	return err
// }

// func CreateDoc(ctx *app_context.Context, obj interface{}, tx ...*gorm.DB) error {
// 	db := ctx.DB
// 	if len(tx) != 0 {
// 		db = tx[0]
// 	}
// 	err := orm.Create(db, obj)
// 	if err != nil {
// 		ctx.Log.WithFields(log.Fields{"error": err}).Errorf("%v.Create", ObjectTypeName(obj))
// 	}
// 	return err
// }

// func Update(ctx *app_context.Context, obj interface{}, tx ...*gorm.DB) error {
// 	db := ctx.DB
// 	if len(tx) != 0 {
// 		db = tx[0]
// 	}
// 	err := orm.Update(db, obj)
// 	if err != nil {
// 		ctx.Log.WithFields(log.Fields{"error": err}).Errorf("%v.Update", ObjectTypeName(obj))
// 	}
// 	return err
// }

// func UpdateField(ctx *app_context.Context, obj interface{}, field string, tx ...*gorm.DB) error {
// 	db := ctx.DB
// 	if len(tx) != 0 {
// 		db = tx[0]
// 	}
// 	err := orm.UpdateField(db, field, obj)
// 	if err != nil {
// 		ctx.Log.WithFields(log.Fields{"field": field, "error": err}).Errorf("%v.UpdateField", ObjectTypeName(obj))
// 	}
// 	return err
// }

// func UpdateFields(ctx *app_context.Context, obj interface{}, fields ...string) error {
// 	fs := append([]string{}, fields...)
// 	err := orm.UpdateFields(ctx.DB, fs, obj)
// 	if err != nil {
// 		ctx.Log.WithFields(log.Fields{"fields": fs, "error": err}).Errorf("%v.UpdateFields", ObjectTypeName(obj))
// 	}
// 	return err
// }

// func UpdateFieldsTx(ctx *app_context.Context, tx *gorm.DB, obj interface{}, fields ...string) error {
// 	fs := append([]string{}, fields...)
// 	err := orm.UpdateFields(tx, fs, obj)
// 	if err != nil {
// 		ctx.Log.WithFields(log.Fields{"fields": fs, "error": err}).Errorf("%v.UpdateFieldsTx", ObjectTypeName(obj))
// 	}
// 	return err
// }

// func CheckFound(notfound bool, err *error) bool {
// 	ok := *err == nil && !notfound
// 	if notfound {
// 		*err = errors.New("not found")
// 	}
// 	return ok
// }

// func CheckFoundNoError(notfound bool, err *error) bool {
// 	ok := *err == nil && !notfound
// 	if notfound {
// 		*err = nil
// 	}
// 	return ok
// }

// func CheckFoundDbError(notfound bool, err error) error {
// 	if err != nil && !notfound {
// 		return err
// 	}
// 	return nil
// }

// func FindByFields(ctx *app_context.Context, fields map[string]interface{}, obj interface{}, tx ...*gorm.DB) (bool, error) {
// 	db := ctx.DB
// 	if len(tx) != 0 {
// 		db = tx[0]
// 	}
// 	notFound, err := orm.FindByFields(db, fields, obj)
// 	if err != nil && !notFound {
// 		ctx.Log.WithFields(log.Fields{"fields": fields, "error": err}).Errorf("%v.FindByFields", ObjectTypeName(obj))
// 	}
// 	return notFound, err
// }

// var ObjectTypeName = orm.ObjectTypeName

// func FindNotIn(ctx *app_context.Context, fields map[string]interface{}, obj interface{}, tx ...*gorm.DB) error {
// 	db := ctx.DB
// 	if len(tx) != 0 {
// 		db = tx[0]
// 	}
// 	err := orm.FindNotIn(db, fields, obj)
// 	if err != nil {
// 		ctx.Log.WithFields(log.Fields{"fields": fields, "error": err}).Errorf("%v.FindNotIn", ObjectTypeName(obj))
// 	}
// 	return err
// }

// func FindAllByFields(ctx *app_context.Context, fields map[string]interface{}, obj interface{}, tx ...*gorm.DB) error {
// 	db := ctx.DB
// 	if len(tx) != 0 {
// 		db = tx[0]
// 	}
// 	err := orm.FindAllByFields(db, fields, obj)
// 	if err != nil {
// 		ctx.Log.WithFields(log.Fields{"fields": fields, "error": err}).Errorf("%v.FindNotIn", ObjectTypeName(obj))
// 	}
// 	return err
// }

// func DeleteAll(ctx *app_context.Context, obj interface{}, tx ...*gorm.DB) error {
// 	db := ctx.DB
// 	if len(tx) != 0 {
// 		db = tx[0]
// 	}
// 	err := orm.DeleteAll(db, obj)
// 	if err != nil {
// 		ctx.Log.WithFields(log.Fields{"error": err}).Errorf("%v.DeleteAll", ObjectTypeName(obj))
// 	}
// 	return err
// }
