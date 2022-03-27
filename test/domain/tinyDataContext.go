package domain

import (
	"github.com/joinlee/tiny-entity-go"
	tinyMysql "github.com/joinlee/tiny-entity-go/mysql"
	"github.com/joinlee/tiny-entity-go/test/domain/models"
)

type TinyDataContext struct {
	*tinyMysql.MysqlDataContext
	Account *models.Account
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
	return ctx
}
func (this *TinyDataContext) CreateDatabase() {
	this.MysqlDataContext.CreateDatabase()
	this.CreateTable(this.Account)
}
func (this *TinyDataContext) GetEntityList() map[string]tiny.Entity {
	list := make(map[string]tiny.Entity)
	list["Account"] = this.Account
	return list
}
