package db_gorm

import (
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

type GormCursor struct {
	rows   gorm.Rows
	gormDB *GormDB

	sql *sql.Rows
}

func (c *GormCursor) Close() error {
	err := c.rows.Close()
	if err != nil {
		err = fmt.Errorf("failed to close rows")
		c.gormDB.Logger().Error("GormDB.Cursor", err)
	}
	return err
}

func (c *GormCursor) Scan(obj interface{}) error {
	err := c.gormDB.db.ScanRows(c.sql, obj)
	if err != nil {
		err = fmt.Errorf("failed to scan rows to object %v", ObjectTypeName(obj))
		c.gormDB.Logger().Error("GormDB.Cursor", err)
	}
	return err
}

func (c *GormCursor) Next() (bool, error) {
	next := c.rows.Next()
	err := c.rows.Err()
	if err != nil {
		err = fmt.Errorf("failed to read next rows")
		c.gormDB.Logger().Error("GormDB.Cursor", err)
	}
	return next, err
}
