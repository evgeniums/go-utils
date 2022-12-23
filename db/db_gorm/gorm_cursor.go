package db_gorm

import (
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

type GormCursor struct {
	Rows   gorm.Rows
	GormDB *GormDB

	Sql *sql.Rows
}

func (c *GormCursor) Close() error {
	err := c.Rows.Close()
	if err != nil {
		err = fmt.Errorf("failed to close rows")
		c.GormDB.Logger().Error("GormDB.Cursor", err)
	}
	return err
}

func (c *GormCursor) Scan(obj interface{}) error {
	err := c.GormDB.DB.ScanRows(c.Sql, obj)
	if err != nil {
		err = fmt.Errorf("failed to scan rows to object %v", ObjectTypeName(obj))
		c.GormDB.Logger().Error("GormDB.Cursor", err)
	}
	return err
}

func (c *GormCursor) Next() (bool, error) {
	next := c.Rows.Next()
	err := c.Rows.Err()
	if err != nil {
		err = fmt.Errorf("failed to read next rows")
		c.GormDB.Logger().Error("GormDB.Cursor", err)
	}
	return next, err
}
