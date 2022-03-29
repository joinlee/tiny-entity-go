package tinyMysql

import (
	"fmt"

	"github.com/joinlee/tiny-entity-go"
)

type EntityObjectMysql[T tiny.Entity] struct {
	base *tiny.EntityObjectBase[T]
}

func NewEntityObjectMysql[T tiny.Entity](ctx *MysqlDataContext, tableName string) *EntityObjectMysql[T] {
	entity := &EntityObjectMysql[T]{}

	entity.base = &tiny.EntityObjectBase[T]{}
	entity.base.InitEntityObj(tableName)
	entity.base.SetCtx(ctx.DataContextBase)

	return entity
}

func (this *EntityObjectMysql[T]) TableName() string {
	return this.base.TableName
}

func (this *EntityObjectMysql[T]) GetIQueryObject() tiny.IQueryObject[T] {
	return this
}

func (this *EntityObjectMysql[T]) And() tiny.IQueryObject[T] {
	this.base.Ctx.AddToWhere("AND", false)
	return this
}

func (this *EntityObjectMysql[T]) Or() tiny.IQueryObject[T] {
	this.base.Ctx.AddToWhere("OR", false)
	return this
}

// 添加查询条件
/* queryStr 查询语句， args 条件参数
ex： ctx.User.Where("Id = ?", user.Id).Any() */
func (this *EntityObjectMysql[T]) Where(queryStr interface{}, args ...interface{}) tiny.IQueryObject[T] {
	this.base.WherePartHandle(this.base.TableName, queryStr, args)
	return this
}

//添加指定表的查询条件
/* entity 需要查询的实体 queryStr 查询语句， args 条件参数
entity 表示查询外键表的条件
ex： ctx.User.WhereWith(ctx.Account, "Id = ?", user.Id).Any() */
func (this *EntityObjectMysql[T]) WhereWith(entity tiny.Entity, queryStr interface{}, args ...interface{}) tiny.IQueryObject[T] {
	tableName := this.base.Ctx.GetEntityName(entity)
	this.base.WherePartHandle(tableName, queryStr, args)
	return this
}

func (this *EntityObjectMysql[T]) Contains(felid string, values interface{}) tiny.IQueryObject[T] {
	this.base.InPartHandle(this.base.TableName, felid, values)
	return this
}

func (this *EntityObjectMysql[T]) ContainsWith(entity tiny.Entity, felid string, values interface{}) tiny.IQueryObject[T] {
	tableName := this.base.Ctx.GetEntityName(entity)
	this.base.InPartHandle(tableName, felid, values)
	return this
}

func (this *EntityObjectMysql[T]) OrderBy(field interface{}) tiny.IQueryObject[T] {
	this.base.Ctx.AddToOrdereBy(field.(string), false, this.base.TableName)
	return this
}

func (this *EntityObjectMysql[T]) OrderByDesc(field interface{}) tiny.IQueryObject[T] {
	this.base.Ctx.AddToOrdereBy(field.(string), true, this.base.TableName)
	return this
}

func (this *EntityObjectMysql[T]) IndexOf() tiny.IQueryObject[T] {
	return this
}

func (this *EntityObjectMysql[T]) GroupBy(field interface{}) tiny.IResultQueryObject[T] {
	this.base.GroupByHandle(field.(string))
	return this
}

func (this *EntityObjectMysql[T]) Select(fields ...interface{}) tiny.IResultQueryObject[T] {
	this.base.SelectHandle(fields...)
	return this
}

func (this *EntityObjectMysql[T]) Take(count int) tiny.ITakeChildQueryObject[T] {
	this.base.Ctx.AddToLimt("take", count)
	return this
}
func (this *EntityObjectMysql[T]) Skip(count int) tiny.IAssembleResultQuery[T] {
	this.base.Ctx.AddToLimt("skip", count)
	return this
}

// 添加外联引用
/*
	fEntity 需要连接的实体， mField 主表的连接字段， fField 外联表的字段
*/
func (this *EntityObjectMysql[T]) JoinOn(fEntity tiny.Entity, mField string, fField string) tiny.IQueryObject[T] {
	this.base.JoinHandle(this.base.TableName, fEntity, mField, fField)
	return this
}

func (this *EntityObjectMysql[T]) JoinOnWith(mEntity tiny.Entity, fEntity tiny.Entity, mField string, fField string) tiny.IQueryObject[T] {
	mTableName := mEntity.TableName()
	this.base.JoinHandle(mTableName, fEntity, mField, fField)
	return this
}

func (this *EntityObjectMysql[T]) Max() float64 {
	this.base.Ctx.Clean()
	return 0
}

func (this *EntityObjectMysql[T]) Min() float64 {
	this.base.Ctx.Clean()
	return 0
}
func (this *EntityObjectMysql[T]) Count() int {
	return this.CountArgs(fmt.Sprintf("`%s`.`Id`", this.base.TableName))
}

func (this *EntityObjectMysql[T]) CountArgs(field string) int {
	return this.base.CountHandle(field)
}

func (this *EntityObjectMysql[T]) Any() bool {
	count := this.Count()
	return count > 0
}

func (this *EntityObjectMysql[T]) First() *T {
	return this.base.First()
}

func (this *EntityObjectMysql[T]) ToList() []T {
	return this.base.ToList()
}
