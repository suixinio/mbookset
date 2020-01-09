package models

import (
	"time"
)

// Model represents meta data of entity.
type Model struct {
	ID uint64 `orm:"pk;unique;column(id)" json:"id"`
	//ID        uint64     `orm:"primary_key" json:"id"`
	CreatedAt time.Time  `orm:"auto_now_add;column(created_at);null" json:"createdAt"`
	UpdatedAt time.Time  `orm:"auto_now;column(updated_at);null" json:"updatedAt"`
	DeletedAt *time.Time `orm:"column(deleted_at);null "sql:"index" json:"deletedAt"`
}
