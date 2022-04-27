# tiny-entity-go

## Table of Contents
  - [Install](#install)
  - [Introduction](#introduction)
  - [Define](#define)
  - [Query](#query)
  - [Command](#command)

## Install

```sh
$ go get -u github.com/shishisongsong/tiny-entity-go
```

## Introduction

This is a ORM framework support Mysql and Kingbase. In the design process, I refer to ling and Entity Framework. I think EF is a beautiful design. 
You can use this orm map domain model to database. May your domain mode been changed, you can use this orm migration to you DB. Keep your db physical mode and domain mode with same.

In the other orm framework like gorm , you can't get a domain mode very well. because gorm give you a univeresal mode,that feel not good. So in tiny-entity I use generic solve this problem. 

Cuse generic , you need use golang at version 1.18.

## Define

At the first you must define your domain mode.

you can define an entity model like this:
``` golang
package models

import "github.com/shishisongsong/tiny-entity-go"

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
```

### PrimaryKey
```go
Id          string        `tiny:"primaryKey" json:"id"`
```
use keyword 'primaryKey' to define a PK in a mode. This field data type will be varchar(32). I use guid as the PK, and not support auto increment now.

---
### Data Type
you can use varchar(len), bigint, int(11), decimal(10), bool define your field. These data type is same at most DB like Mysql.

keyword
```go
VARCHAR     = "varchar"
TEXT        = "text"
LONGTEXT    = "longtext"
DECIMAL     = "decimal"
INT         = "int"
BIGINT      = "bigint"
BOOL        = "tinyint"
ARRARY      = "arrary"
JSON        = "json"
```

If you want add index ,you can use the keyword "index".
``` go
Name        string        `tiny:"varchar(255);notNull;index" json:"name"`
```

Set field default value. 
``` go
Phone       string        `tiny:"varchar(255);default:123" json:"phone"`
```

If field may be a null value, you must use ptr define.
```go
Age         *int          `tiny:"int(10)" json:"age"`
```

### Mapping Other Mode
### One on One Mapping

```go
package models

import (
	"github.com/shishisongsong/tiny-entity-go"
)

type Account struct {
	tiny.IEntityObject[Account]
	Id         string  `tiny:"primaryKey" json:"id"`
	Username   string  `tiny:"type:varchar(255);notNull" json:"username"`
	Password   string  `tiny:"type:varchar(255);notNull" json:"password"`
	Status     *string `tiny:"type:varchar(255)" json:"status"`
	CreateTime int64   `tiny:"type:bigint" json:"createTime"`
}

```

If User with Account one on one mapping. you can define a field in User. Use keyword maaping, the params is the mode name that you will be refence.

```go
Account     *Account      `tiny:"mapping:Account" json:"account"`
```

### One To Many Mapping

```go
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
```

In User mode you can define field like this:
```go
UserAddress []UserAddress `tiny:"mapping:UserAddress" json:"userAddress"`
```


you can use PrimaryKey() define a primarykey for an entity , and also use Column() define a field.

you can use some parameters for Column() like this:
```ts
@Define.Column({ 
        DataType: Define.DataType.Decimal, 
        DataLength: 11, 
        DecimalPoint: 3 
})
```
method list:
PrimaryKey(opt?: PropertyDefineOption):
define a primarykey faster then Column().

Column(opt?: PropertyDefineOption)
define a field. and also you can use this function define a primaryKey.

Mapping(opt: PropertyDefineOption)
define a mapping. it is used to deal with database foreignKey .
```ts
interface PropertyDefineOption{
    DataType?: DataType;
    DefaultValue?: any;
    NotAllowNULL?: boolean;
    DataLength?: number;
    ColumnName?: string;
    IsPrimaryKey?: boolean;
    ForeignKey?: { ForeignTable: string; ForeignColumn: string; IsPhysics?: boolean; };
    DecimalPoint?: number;
    IsIndex?: boolean;
    Mapping?: string;
    MappingType?: MappingType;
    MappingKey?: { FKey: string, MKey?: string } | string;
}

enum DataType {
        VARCHAR,
        TEXT,
        LONGTEXT,
        Decimal,
        INT,
        BIGINT,
        BOOL,
        Array,
        JSON
    }
```


---
## Query

query datas from table, return array.
``` ts
let list = await ctx.Person.Where(x => x.age > age, { age }).ToList();
```

``` ts
let list = await ctx.Person.Where(x => x.name.indexOf($args1), { $args1: params.name }).ToList();
```

using left join:
``` ts
let list = await ctx.Person
            .Join(ctx.Account)
            .On((m, f) => m.id == f.personId)
            .Contains<Account>(x => x.amount, values2, ctx.Account)
            .ToList();
```

query single entity
```ts
let list = await ctx.Person.First(x => x.name.indexOf($args1), { $args1: params.name });
```

using transcation:
``` ts
await Transaction(new TestDataContext(), async (ctx) => {
                //insert 10 persons to database;
                for (let i = 0; i < 10; i++) {
                    let person = new Person();
                    person.id = Guid.GetGuid();
                    person.name = "likecheng" + i;
                    person.age = 30 + i;
                    person.birth = new Date("1987-12-1").getTime();
                    if (i == 9)
                        throw ' transaction error';
                    await ctx.Create(person);
                }
            });
```

## Command
you need install tiny-entity2 global
```sh
$ npm install tiny-entity2 -g
```
at first you need to create the tinyconfig.json in your project.
this file provide some options to command.
```json
{
    "outDir": "./test",
    "modelLoadPath": [
        "./test/models"
    ],
    "modelExportPath": [
        "./models"
    ],
    "ctxExportPath": "",
    "configFilePath": "./config",
    "outFileName": "testDataContext.ts",
    "databaseType": "mysql",
    "packageName": "../mysql/dataContextMysql"
}
```
property description:
* outDir: directory of file output.
* modelLoadPath: directory of model file import.
* modelExportPath: directory of model file export.
* ctxExportPath: directory of data context file export.
* configFilePath: directory of database config file .
* outFileName: the name of data context file.
* databaseType: database type, 'mysql' or 'sqlite'.
* packageName: path of data context engine. now only support dataContextMysql and dataContextSqlite

use command 'gctx' create a dataContext.ts file. 
```sh
tiny --gctx ./tingconfig.json
```
use command 'gdb' create a new database file .
```sh
tiny --gdb ./tingconfig.json
```
use comman d 'gop' create a update log and excute to database.
```sh
tiny --gop ./tingconfig.json
```

you can create different tinyconfig.json  for different domain models.




