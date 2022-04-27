package test

import (
	"fmt"
	"os"
	"testing"

	_ "net/http/pprof"

	"github.com/shishisongsong/tiny-entity-go"
	"github.com/shishisongsong/tiny-entity-go/test/domain"
	"github.com/shishisongsong/tiny-entity-go/test/domain/models"
	"github.com/shishisongsong/tiny-entity-go/utils"

	tinyKing "github.com/shishisongsong/tiny-entity-go/kingbase"
	tinyMysql "github.com/shishisongsong/tiny-entity-go/mysql"
)

func SetEnv() {
	os.Setenv("TINY_LOG", "ON")
}

func TestGCTX(t *testing.T) {
	ctx := &tinyMysql.MysqlDataContext{}
	codeGenerator := GetMysqlCodeGenerator(ctx)
	codeGenerator.GenerateCtxFile()
}

func TestGDB(t *testing.T) {
	ctx := domain.NewTinyDataContext()
	ctx.CreateDatabase()
}

func TestGOP(t *testing.T) {
	ctx := domain.NewTinyDataContext()
	codeGenerator := GetMysqlCodeGenerator(ctx)
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
		// account.Status = ""
		account.CreateTime = 1000

		ctx.Create(account)

		targetAccount := ctx.Account.Where("Id = ?", account.Id).First()

		if targetAccount.CreateTime != account.CreateTime {
			t.Errorf("input: %d, output: %d", account.CreateTime, targetAccount.CreateTime)
		}

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

func TestQuery(t *testing.T) {
	SetEnv()
	ctx := domain.NewTinyDataContext()

	tiny.Transaction(ctx, func(ctx *domain.TinyDataContext) {
		ctx.DeleteWith(ctx.Account, "")
		ctx.DeleteWith(ctx.User, "")

		output0 := ctx.Account.ToList()
		if len(output0) != 0 {
			t.Errorf("output0 lenght is not 0")
		}

		count := 10000
		// prepare data
		for i := 0; i < count; i++ {
			account := new(models.Account)
			account.Id = utils.GetGuid()
			account.Username = "admin" + fmt.Sprintf("%d", i)
			account.Password = "123"
			// account.Status = ""

			user := new(models.User)
			user.Id = utils.GetGuid()
			user.Name = "john lee" + fmt.Sprintf("%d", i)
			user.Phone = "13245678765"
			user.AccountId = account.Id

			ctx.Create(account)
			ctx.Create(user)
		}

		output1 := ctx.User.JoinOn(ctx.Account, "AccountId", "Id").ToList()
		if len(output1) != count {
			t.Errorf("count not equals 10, IS length: %d", len(output1))
		}

		for _, item := range output1 {
			if item.AccountId != item.Account.Id {
				t.Errorf("users accountId is not equals Account object id")
			}

			if item.Account.Status != nil {
				t.Errorf("Account.Status need nil, now : %s", *item.Account.Status)
			}
		}

		output2 := ctx.User.Max("Name")
		if output2 != "john lee9" {
			t.Errorf("max name is : %s", output2)
		}

		output3 := ctx.User.Min("Name")
		if output3 != "john lee0" {
			t.Errorf("min name is : %s", output3)
		}
	})
}

func TestBigDataInsert(t *testing.T) {
	SetEnv()
	ctx := domain.NewTinyDataContext()

	tiny.Transaction(ctx, func(ctx *domain.TinyDataContext) {
		ctx.DeleteWith(ctx.Account, "")
		ctx.DeleteWith(ctx.User, "")

		output0 := ctx.Account.ToList()
		if len(output0) != 0 {
			t.Errorf("output0 lenght is not 0")
		}

		count := 10000
		// prepare data
		for i := 0; i < count; i++ {
			account := new(models.Account)
			account.Id = utils.GetGuid()
			account.Username = "admin" + fmt.Sprintf("%d", i)
			account.Password = "123"
			// account.Status = ""

			user := new(models.User)
			user.Id = utils.GetGuid()
			user.Name = "john lee" + fmt.Sprintf("%d", i)
			user.Phone = "13245678765"
			user.AccountId = account.Id

			ctx.Create(account)
			ctx.Create(user)
		}
	})
}

func TestBigData(t *testing.T) {
	SetEnv()
	ctx := domain.NewTinyDataContext()
	count := 10000
	output1 := ctx.User.JoinOn(ctx.Account, "AccountId", "Id").ToList()
	if len(output1) != count {
		t.Errorf("count not equals 10, IS length: %d", len(output1))
	}

	for _, item := range output1 {
		if item.AccountId != item.Account.Id {
			t.Errorf("users accountId is not equals Account object id")
		}

		if item.Account.Status != nil {
			t.Errorf("Account.Status need nil, now : %s", *item.Account.Status)
		}
	}
}

func TestJoinOneMany(t *testing.T) {
	SetEnv()
	ctx := domain.NewTinyDataContext()

	var userId = ""
	var userId2 = ""

	tiny.Transaction(ctx, func(ctx *domain.TinyDataContext) {
		ctx.DeleteWith(ctx.Account, "")
		ctx.DeleteWith(ctx.User, "")
		ctx.DeleteWith(ctx.UserAddress, "")

		account := new(models.Account)
		account.Id = utils.GetGuid()
		account.Username = "lkc"
		account.Password = "123"
		account.CreateTime = 11122233

		ctx.Create(account)

		user := new(models.User)
		user.Id = utils.GetGuid()
		user.Name = "likecheng"
		user.AccountId = account.Id

		userId = user.Id

		ctx.Create(user)

		user2 := new(models.User)
		user2.Id = utils.GetGuid()
		user2.Name = "likecheng"
		user2.AccountId = account.Id
		userId2 = user2.Id
		ctx.Create(user2)

		for i := 0; i < 5; i++ {
			userAddress := new(models.UserAddress)
			userAddress.Id = utils.GetGuid()
			userAddress.Phone = fmt.Sprintf("1598765456%d", i)
			userAddress.UserId = user.Id
			userAddress.ReciverName = fmt.Sprintf("lkc%d", i)
			userAddress.Address = "china sichuan chengdu wuhouqu"

			ctx.Create(userAddress)
		}
	})

	obj := ctx.User.JoinOn(ctx.UserAddress, "Id", "UserId").Where("Id = ?", userId).First()

	if len(obj.UserAddress) != 5 {
		t.Errorf("obj UserAddress list must be 5!, now %d", len(obj.UserAddress))
	}

	output2 := ctx.User.JoinOn(ctx.UserAddress, "Id", "UserId").Where("Id = ?", userId2).First()
	if output2.UserAddress == nil {
		t.Errorf("output2 userAddress is not be null")
	}
	if len(output2.UserAddress) != 0 {
		t.Errorf("obj UserAddress list must be 0!, now %d", len(output2.UserAddress))
	}
}

func TestGCTX_KB(t *testing.T) {
	ctx := &tinyKing.KingDataContext{}
	codeGenerator := GetKingBaseCodeGenerator(ctx)
	codeGenerator.GenerateCtxFile()
}

func TestGDB_KB(t *testing.T) {
	SetEnv()
	ctx := domain.NewTinyDataContextKingBase()
	ctx.CreateDatabase()
}

func TestGOP_KB(t *testing.T) {
	SetEnv()
	ctx := domain.NewTinyDataContextKingBase()
	codeGenerator := GetKingBaseCodeGenerator(ctx)
	codeGenerator.AutoMigration()
}

func TestCreate_KB(t *testing.T) {
	SetEnv()
	ctx := domain.NewTinyDataContext()

	tiny.Transaction(ctx, func(ctx *domain.TinyDataContext) {
		ctx.DeleteWith(ctx.Account, "")

		account := new(models.Account)
		account.Id = utils.GetGuid()
		account.Username = "likecheng"
		account.Password = "123"
		// account.Status = ""

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

func TestQuery_KB(t *testing.T) {
	SetEnv()
	ctx := domain.NewTinyDataContextKingBase()

	tiny.Transaction(ctx, func(ctx *domain.TinyDataContextKingBase) {
		ctx.DeleteWith(ctx.Account, "")
		ctx.DeleteWith(ctx.User, "")

		output0 := ctx.Account.ToList()
		if len(output0) != 0 {
			t.Errorf("output0 lenght is not 0")
		}

		// prepare data
		for i := 0; i < 10; i++ {
			account := new(models.Account)
			account.Id = utils.GetGuid()
			account.Username = "admin" + fmt.Sprintf("%d", i)
			account.Password = "123"
			// account.Status = ""

			user := new(models.User)
			user.Id = utils.GetGuid()
			user.Name = "john lee" + fmt.Sprintf("%d", i)
			user.Phone = "13245678765"
			user.AccountId = account.Id

			ctx.Create(account)
			ctx.Create(user)
		}

		output1 := ctx.User.JoinOn(ctx.Account, "AccountId", "Id").ToList()
		if len(output1) != 10 {
			t.Errorf("count not equals 10, IS length: %d", len(output1))
		}

		for _, item := range output1 {
			if item.AccountId != item.Account.Id {
				t.Errorf("users accountId is not equals Account object id")
			}
		}

		output2 := ctx.User.Where("Phone = ?", "13245678765").Max("Name")
		if output2 != "john lee9" {
			t.Errorf("max name is : %s", output2)
		}

		output3 := ctx.User.Where("Phone = ?", "13245678765").Min("Name")
		if output3 != "john lee0" {
			t.Errorf("min name is : %s", output3)
		}
	})
}
