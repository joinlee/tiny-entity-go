package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/joinlee/tiny-entity-go"
	"github.com/joinlee/tiny-entity-go/test/domain"
	"github.com/joinlee/tiny-entity-go/test/domain/models"
	"github.com/joinlee/tiny-entity-go/utils"
)

func SetEnv() {
	os.Setenv("TINY_LOG", "ON")
}

func TestGCTX(t *testing.T) {
	ctx := domain.NewTinyDataContext()
	codeGenerator := GetCodeGenerator(ctx)
	codeGenerator.GenerateCtxFile()
}

func TestGDB(t *testing.T) {
	ctx := domain.NewTinyDataContext()
	ctx.CreateDatabase()
}

func TestGOP(t *testing.T) {
	ctx := domain.NewTinyDataContext()
	codeGenerator := GetCodeGenerator(ctx)
	codeGenerator.AutoMigration()
}

func TestCreate(t *testing.T) {
	SetEnv()
	ctx := domain.NewTinyDataContext()

	tiny.Transaction(ctx, func(ctx *domain.TinyDataContext) {
		ctx.DeleteWith(ctx.Account, "")

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

		output := ctx.Account.Where("Id = ?", "123").First()
		if output != nil {
			t.Errorf("targetAccount is not nil ")
		}

		targetAccount.Username = "lkc"
		ctx.Update(targetAccount)

		output2 := ctx.Account.Where("Id = ?", account.Id).First()
		if output2.Username != "lkc" {
			t.Errorf("output2.Username != lkc")
		}

		ctx.Delete(targetAccount)

		output3 := ctx.Account.Count()
		if output3 != 0 {
			t.Errorf("output3 is not 0")
		}
	})
}

func TestToList(t *testing.T) {
	ctx := domain.NewTinyDataContext()
	accounts := ctx.Account.ToList()

	fmt.Printf("list length : %d", len(accounts))
}
