package models

import "github.com/joinlee/tiny-entity-go"

type User struct {
	tiny.IEntityObject[User]
	Id          string        `tiny:"primaryKey" json:"id"`
	Name        string        `tiny:"varchar(255);notNull;index" json:"name"`
	Phone       string        `tiny:"varchar(255);default:123" json:"phone"`
	Age         *int          `tiny:"int(10)" json:"age"`
	AccountId   string        `tiny:"varchar(32);index" json:"accountId"`
	Account     *Account      `tiny:"mapping:Account" json:"account"`
	UserAddress []UserAddress `tiny:"mapping:UserAddress" json:"userAddress"`
}
