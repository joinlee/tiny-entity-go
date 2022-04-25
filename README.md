# tiny-entity-go

## Table of Contents
  - [Install](#install)
  - [Introduction](#introduction)
  - [Define](#define)
  - [Query](#query)
  - [Command](#command)

## Install

```sh
$ go get -u github.com/joinlee/tiny-entity-go
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

```

If User with Account one on one mapping. you can define a field in User. Use keyword maaping, the params is the mode name that you will be refence.

```go
Account     *Account      `tiny:"mapping:Account" json:"account"`
```

### One To Many Mapping

```go
package models

import "github.com/joinlee/tiny-entity-go"

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
Then you can query like this;
```go
ctx := domain.NewTinyDataContext()
users := ctx.User.JoinOn(ctx.UserAddress, "Id", "UserId").ToList()
```

How to insert a data to DB.
```go
ctx := domain.NewTinyDataContext()
user := new(models.User)
user.Id = utils.GetGuid()
user.Name = "john lee" + fmt.Sprintf("%d", i)
user.Phone = "13245678765"
user.AccountId = account.Id

ctx.Create(user)
```

How to update date?
```go
ctx := domain.NewTinyDataContext()
usere := ctx.User.First()
//change user name 
user.Name = "new name";

ctx.Update(user)
```
And also you can use update with specific conditions, like this:
``` go
ctx := domain.NewTinyDataContext()
ctx.UpdateWith(ctx.User, []string{"Name = 'john lee'"}, "Age > 18")
```
