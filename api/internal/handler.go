package internal

import "gorm.io/gorm"

type InternalHandler struct {
	Db *gorm.DB
}
