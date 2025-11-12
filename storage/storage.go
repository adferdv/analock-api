package storage

import (
	"database/sql"
)

type Storage interface {
	Get(uint) (interface{}, error)
	Create(interface{}) error
	Update(interface{}) error
	Delete(uint) error
	Scan(*sql.Rows) (interface{}, error)
}
