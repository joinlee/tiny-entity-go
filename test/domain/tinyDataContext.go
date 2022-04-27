package domain

import (
	"github.com/joinlee/tiny-entity-go"
	tinyMysql "github.com/joinlee/tiny-entity-go/mysql"
	"github.com/joinlee/tiny-entity-go/test/domain/models"
)

type TinyDataContext struct {
	*tinyMysql.MysqlDataContext
	Account     *models.Account
	User        *models.User
	UserAddress *models.UserAddress
}

func NewTinyDataContext() *TinyDataContext {
	ctx := &TinyDataContext{}
	ctx.MysqlDataContext = tinyMysql.NewMysqlDataContext(tiny.DataContextOptions{
		Host:            "localhost",
		Port:            "3306",
		Username:        "root",
		Password:        "123456",
		DataBaseName:    "tinygotest",
		CharSet:         "utf8",
		ConnectionLimit: 50,
	})

	ctx.Account = &models.Account{
		IEntityObject: tinyMysql.NewEntityObjectMysql[models.Account](ctx.MysqlDataContext, "Account")}
	ctx.RegistModel(ctx.Account)
	ctx.User = &models.User{
		IEntityObject: tinyMysql.NewEntityObjectMysql[models.User](ctx.MysqlDataContext, "User")}
	ctx.RegistModel(ctx.User)
	ctx.UserAddress = &models.UserAddress{
		IEntityObject: tinyMysql.NewEntityObjectMysql[models.UserAddress](ctx.MysqlDataContext, "UserAddress")}
	ctx.RegistModel(ctx.UserAddress)
	return ctx
}
func (this *TinyDataContext) CreateDatabase() {
	this.MysqlDataContext.CreateDatabase()
	this.CreateTable(this.Account)
	this.CreateTable(this.User)
	this.CreateTable(this.UserAddress)
}
func (this *TinyDataContext) GetEntityList() map[string]tiny.Entity {
	list := make(map[string]tiny.Entity)
	list["Account"] = this.Account
	list["User"] = this.User
	list["UserAddress"] = this.UserAddress
	return list
}
