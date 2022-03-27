package test

import (
	"fmt"
	"testing"

	"github.com/joinlee/tiny-entity-go/test/domain"
	"github.com/joinlee/tiny-entity-go/test/domain/models"
	"github.com/joinlee/tiny-entity-go/utils"
)

func TestGCTX(t *testing.T) {
	codeGenerator := GetCodeGenerator()
	codeGenerator.GenerateCtxFile()
}

func TestGDB(t *testing.T) {
	ctx := domain.NewTinyDataContext()
	ctx.CreateDatabase()
}

func TestGOP(t *testing.T) {
	ctx := domain.NewTinyDataContext()
	codeGenerator := GetCodeGenerator()
	codeGenerator.AutoMigration(ctx)
}

func TestCreate(t *testing.T) {
	ctx := domain.NewTinyDataContext()

	account := new(models.Account)
	account.Id = utils.GetGuid()
	account.Username = "likecheng"
	account.Password = "123"
	account.Status = ""

	ctx.Create(account)

	targetAccount := ctx.Account.Where("Id = ?", account.Id).First()

	if targetAccount.Id != account.Id {
		t.Errorf("input: %s, output: %s", account.Id, targetAccount.Id)
	}

	targetAccount = ctx.Account.Where("Id = ?", "123").First()
	if targetAccount != nil {
		t.Errorf("targetAccount is not nil ")
	}
}

func TestToList(t *testing.T) {
	ctx := domain.NewTinyDataContext()
	accounts := ctx.Account.ToList()

	fmt.Printf("list length : %d", len(accounts))
}
