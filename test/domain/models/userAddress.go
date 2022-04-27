package models

import "github.com/shishisongsong/tiny-entity-go"

type UserAddress struct {
	tiny.IEntityObject[UserAddress]
	Id          string `tiny:"primaryKey" json:"id"`
	UserId      string `tiny:"varchar(32);notNull;index" json:"userId"`
	Address     string `tiny:"varchar(500)" json:"address"`
	Phone       string `tiny:"varchar(32)" json:"phone"`
	ReciverName string `tiny:"varchar(200)" json:"reciverName"`
}
