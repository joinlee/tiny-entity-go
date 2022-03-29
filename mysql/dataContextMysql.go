package tinyMysql

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/joinlee/tiny-entity-go"
	"github.com/joinlee/tiny-entity-go/tagDefine"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlDataContext struct {
	// db           *sql.DB
	option tiny.DataContextOptions
	// tx           *sql.Tx
	// tranCount    int
	conStr       string
	entityRefMap map[string]reflect.Type

	*tiny.DataContextBase
}

func NewMysqlDataContext(opt tiny.DataContextOptions) *MysqlDataContext {
	ctx := &MysqlDataContext{}

	conStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&multiStatements=true",
		opt.Username,
		opt.Password,
		opt.Host,
		opt.Port,
		opt.DataBaseName,
		opt.CharSet)
	ctx.conStr = conStr

	ctx.DataContextBase = tiny.NewDataContextBase(opt)
	ctx.Db = tiny.GetDB(conStr, opt.ConnectionLimit, "mysql")

	ctx.option = opt
	ctx.entityRefMap = make(map[string]reflect.Type)

	return ctx
}

//插入数据到数据库
func (this *MysqlDataContext) Create(entity tiny.Entity) {
	sql := this.CreateSql(entity)
	this.Submit(sql)
}

func (this *MysqlDataContext) CreateBatch(entities []tiny.Entity) {
	if len(entities) > 0 {
		sql := this.CreateBatchSql(entities)
		this.Submit(sql)
	}
}

//更新数据到数据库
func (this *MysqlDataContext) Update(entity tiny.Entity) {
	sql := this.UpdateSql(entity)
	this.Submit(sql)
}

//批量更新数据表中的数据
//entity 实体对象
//fields 需要更新的字段列表，传入参数例子：[ Username = 'lkc', age = 18 ]
//queryStr 条件参数 例子：gender = 'male'
func (this *MysqlDataContext) UpdateWith(entity tiny.Entity, fields interface{}, queryStr interface{}) {
	sql := this.UpdateWithSql(entity, fields, queryStr)
	this.Submit(sql)
}

//通过实体Id 删除数据
func (this *MysqlDataContext) Delete(entity tiny.Entity) {
	sql := this.DeleteSql(entity)
	this.Submit(sql)
}

//通过指定条件删除数据
func (this *MysqlDataContext) DeleteWith(entity tiny.Entity, queryStr interface{}, args ...interface{}) {
	sql := this.DeleteWithSql(entity, queryStr, args...)
	this.Submit(sql)
}

func (this *MysqlDataContext) CreateDatabase() {
	conStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s",
		this.option.Username,
		this.option.Password,
		this.option.Host,
		this.option.Port,
		"mysql",
		this.option.CharSet)

	db, err := sql.Open("mysql", conStr)
	if err != nil {
		db.Close()
		panic(err)
	}

	sql := this.CreateDatabaseSQL()

	_, err1 := db.Exec(sql)
	if err1 != nil {
		db.Close()
		panic(err)
	}

	db.Close()
}

func (this *MysqlDataContext) DeleteDatabase() {
}

func (this *MysqlDataContext) CreateTable(entity tiny.Entity) {
	sqlStr := this.CreateTableSQL(entity)
	rows, err := this.Db.Query(sqlStr)
	rows.Close()
	if err != nil {
		panic(err)
	}
}

func (this *MysqlDataContext) RegistModel(entity tiny.Entity) {
	t := reflect.TypeOf(entity).Elem()
	this.entityRefMap[t.Name()] = t
}

func (t *MysqlDataContext) GetEntityFieldsDefineInfo(entity interface{}) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	et := reflect.TypeOf(entity).Elem()
	for i := 0; i < et.NumField(); i++ {
		fd := et.Field(i)
		defineStr, has := t.GetFieldDefineStr(fd)
		if !has {
			continue
		}
		defineMap := t.FormatDefine(defineStr)

		_, isMapping := defineMap[tagDefine.Mapping]
		if isMapping {
			continue
		}

		for key, v := range defineMap {
			if v == "" {
				defineMap[key] = true
			}
		}
		result[fd.Name] = defineMap
	}

	return result
}

func (t *MysqlDataContext) AlterTableDropColumn(tableName string, columnName string) string {
	return fmt.Sprintf("ALTER TABLE `%s` Drop `%s`; ", tableName, columnName)
}

func (t *MysqlDataContext) AlterTableAddColumn(tableName string, columnName string) string {
	return fmt.Sprintf("ALTER TABLE `%s` Add %s; ", tableName, columnName)
}

func (t *MysqlDataContext) AlterTableAlterColumn(tableName string, oldColumnName string, newColumnName string, changeSql string) string {
	return fmt.Sprintf("ALTER TABLE `%s` CHANGE `%s` `%s` %s; ", tableName, oldColumnName, newColumnName, changeSql)
}

func (this *MysqlDataContext) GetColumnSqls(defineMap map[string]interface{}, fieldName string, action string, delIndexSql bool, tableName string) (columnSql string, indexSql string) {
	return this.DataContextBase.GetColumnSqls(defineMap, fieldName, action, delIndexSql, tableName)
}
