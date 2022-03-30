/*
 * @Author: john lee
 * @Date: 2021-06-07 13:28:01
 * @LastEditors: john lee
 * @LastEditTime: 2022-03-30 09:31:26
 * @FilePath: \tiny-entity-go\dataContext.go
 * @Description:
 *
 * Copyright (c) 2022 by 用户/公司名, All Rights Reserved.
 */
package tiny

type IDataContext interface {
	//插入数据到数据库
	Create(entity Entity)
	// 批量插入数据到数据库
	CreateBatch(entities []Entity)
	//更新数据到数据库
	Update(entity Entity)
	//通过指定条件更新数据表中的数据
	//entity IEntity 实体对象
	//fields []string 需要更新的字段列表，传入参数例子：[ Username = 'lkc', age = 18 ]
	//queryStr string 条件参数 例子：gender = 'male'
	UpdateWith(entity Entity, fields interface{}, queryStr interface{})
	//通过实体Id 删除数据
	Delete(entity Entity)
	//通过指定条件删除数据
	//entity 实体对象
	//queryStr 条件参数 例子：gender = 'male'
	// args 参数值
	DeleteWith(entity Entity, queryStr string, args ...interface{})
	//开起事务
	BeginTranscation()
	//提交事务
	Commit()
	//直接查询
	Query(sqlStr string) map[int]map[string]string
	//回滚
	RollBack()
	//创建数据库
	CreateDatabase()
	//创建表并执行到数据库
	CreateTable(entity Entity)
	//获取创建表Sql语句
	CreateTableSQL(entity Entity) string
	//获取删除表Sql语句
	DropTableSQL(tableName string) string
	//删除数据库
	DeleteDatabase()
	//获取上下文实体列表
	GetEntityList() map[string]Entity
}

type IDataContextInterpreter interface {
	IDataContext
	GetEntityFieldsDefineInfo(entity interface{}) map[string]map[string]interface{}
	AlterTableDropColumn(tableName string, columnName string) string
	AlterTableAddColumn(tableName string, columnName string) string
	GetColumnSqls(defineMap map[string]interface{}, fieldName string, action string, delIndexSql bool, tableName string) (columnSql string, indexSql string)
	AlterTableAlterColumn(tableName string, oldColumnName string, newColumnName string, changeSql string) string
}
