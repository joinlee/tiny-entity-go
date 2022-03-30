package domain 
import ( 
 "github.com/joinlee/tiny-entity-go/test/domain/models" 
"github.com/joinlee/tiny-entity-go" 
tinyKing "github.com/joinlee/tiny-entity-go/kingbase" 
) 
type TinyDataContextKingBase struct { 
*tinyKing.KingDataContext 
Account *models.Account 
User *models.User 
} 
func NewTinyDataContextKingBase() *TinyDataContextKingBase { 
ctx := &TinyDataContextKingBase{} 
ctx.KingDataContext = tinyKing.NewKingDataContext(tiny.DataContextOptions{ 
Host:            "localhost", 
Port:            "54321", 
Username:            "root", 
Password:            "123456", 
DataBaseName:            "tinygotest", 
CharSet:            "utf8", 
ConnectionLimit:            50, 
}) 

ctx.Account = &models.Account{ 
IEntityObject: tinyKing.NewEntityObjectKing[models.Account](ctx.KingDataContext, "Account"),}
ctx.RegistModel(ctx.Account)
ctx.User = &models.User{ 
IEntityObject: tinyKing.NewEntityObjectKing[models.User](ctx.KingDataContext, "User"),}
ctx.RegistModel(ctx.User)
return ctx } 
func (this *TinyDataContextKingBase) CreateDatabase() { 
this.KingDataContext.CreateDatabase() 
this.CreateTable(this.Account) 
this.CreateTable(this.User) 
} 
func (this *TinyDataContextKingBase) GetEntityList() map[string]tiny.Entity { 
list := make(map[string]tiny.Entity) 
list["Account"] = this.Account 
list["User"] = this.User 
return list } 
