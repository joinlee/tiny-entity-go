package models

import (
	"github.com/joinlee/tiny-entity-go"
)

type Account struct {
	tiny.IEntityObject[Account]
	Id         string  `tiny:"primaryKey" json:"id"`
	Username   string  `tiny:"type:varchar(255);notNull" json:"username"`
	Password   string  `tiny:"type:varchar(255);notNull" json:"password"`
	Status     *string `tiny:"type:varchar(255)" json:"status"`
	CreateTime int64   `tiny:"type:bigint" json:"createTime"`
}
