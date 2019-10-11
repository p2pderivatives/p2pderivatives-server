package database

import (
	"encoding/hex"
	"fmt"

	"github.com/jinzhu/gorm"
	// Needed for using sqlite
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// NewGormSqlite creates a new GormSqlite object.
func NewGormSqlite(filePath string) (*gorm.DB, error) {
	return gorm.Open("sqlite3", filePath)
}

// HandleGormError takes a gorm error and parameters to return a generic
// database error
func HandleGormError(
	err error, action string, typeName string, id []byte) error {
	message := fmt.Sprintf(
		"Could not %s %s with id %s", action, typeName, hex.EncodeToString(id))
	if gorm.IsRecordNotFoundError(err) {
		err = NewDbError(message, NotFound)
	} else {
		err = NewDbError(message, InternalError)
	}

	return err
}
