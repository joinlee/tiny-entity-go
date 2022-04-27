package tinyKing

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/shishisongsong/tiny-entity-go"
	"github.com/shishisongsong/tiny-entity-go/tagDefine"

	_ "github.com/shishisongsong/kingbase-driver"
)

type KingDataContext struct {
	option tiny.DataContextOptions
	tx     *sql.Tx
	conStr string
	// entityRefMap map[string]reflect.Type

	*tiny.DataContextBase
}

func NewKingDataContext(opt tiny.DataContextOptions) *KingDataContext {
	ctx := &KingDataContext{}
	conStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		opt.Host,
		opt.Username,
		opt.Password,
		opt.DataBaseName)
	ctx.conStr = conStr

	ctx.DataContextBase = tiny.NewDataContextBase(opt)
	ctx.Db = tiny.GetDB(conStr, opt.ConnectionLimit, "kingbase")

	ctx.option = opt
	// ctx.entityRefMap = make(map[string]reflect.Type)

	return ctx
}

func (this *KingDataContext) FilterQuotes(str string) string {
	return strings.ReplaceAll(str, "`", "\"")
}

//插入数据到数据库
func (this *KingDataContext) Create(entity tiny.Entity) {
	sql := this.CreateSql(entity)
	sql = this.FilterQuotes(sql)
	this.Submit(sql)
}

func (this *KingDataContext) CreateBatch(entities []tiny.Entity) {
	if len(entities) > 0 {
		sql := this.CreateBatchSql(entities)
		sql = this.FilterQuotes(sql)
		this.Submit(sql)
	}
}

//更新数据到数据库
func (this *KingDataContext) Update(entity tiny.Entity) {
	sql := this.UpdateSql(entity)
	sql = this.FilterQuotes(sql)
	this.Submit(sql)
}

//批量更新数据表中的数据
//entity 实体对象
//fields 需要更新的字段列表，传入参数例子：[ Username = 'lkc', age = 18 ]
//queryStr 条件参数 例子：gender = 'male'
func (this *KingDataContext) UpdateWith(entity tiny.Entity, fields interface{}, queryStr interface{}) {
	sql := this.UpdateWithSql(entity, fields, queryStr)
	sql = this.FilterQuotes(sql)
	this.Submit(sql)
}

//通过实体Id 删除数据
func (this *KingDataContext) Delete(entity tiny.Entity) {
	sql := this.DeleteSql(entity)
	sql = this.FilterQuotes(sql)
	this.Submit(sql)
}

//通过指定条件删除数据
func (this *KingDataContext) DeleteWith(entity tiny.Entity, queryStr string, args ...interface{}) {
	sql := this.DeleteWithSql(entity, queryStr, args...)
	sql = this.FilterQuotes(sql)
	this.Submit(sql)
}

func (this *KingDataContext) CreateDatabase() {
	conStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=security sslmode=disable",
		this.option.Host,
		this.option.Port,
		this.option.Username,
		this.option.Password)

	db, err := sql.Open("kingbase", conStr)
	err = db.Ping()
	if err != nil {
		db.Close()
		panic(err)
	}

	sql := fmt.Sprintf("CREATE DATABASE \"%s\" encoding utf8;", this.option.DataBaseName)
	tiny.Log(sql)

	_, err1 := db.Exec(sql)
	if err1 != nil {
		db.Close()
		fmt.Println(err1)
	}

	db.Close()
}

func (this *KingDataContext) DeleteDatabase() {
}

func (this *KingDataContext) CreateTable(entity tiny.Entity) {
	sqlStr := this.CreateTableSQL(entity)
	sqlStr = this.FilterQuotes(sqlStr)
	tiny.Log(sqlStr)
	rows, err := this.Db.Query(sqlStr)
	if err != nil {
		panic(err)
	}
	rows.Close()
}

func (this *KingDataContext) GetColumnSqls(defineMap map[string]interface{}, fieldName string, action string, delIndexSql bool, tableName string) (columnSql string, indexSql string) {
	columnSqlList := make([]string, 0)
	_, isPk := defineMap[tagDefine.PRIMARY_KEY]
	column, has := defineMap[tagDefine.Column]
	if !has {
		column = fieldName
	}

	dataType, has := defineMap[tagDefine.Type]
	if !has {
		dataType = "varchar(255)"
		if isPk {
			dataType = "varchar(32)"
		}
	}

	if strings.Contains(fmt.Sprintf("%s", dataType), "tinyint") {
		dataType = "bool"
	}

	if strings.Contains(fmt.Sprintf("%s", dataType), "int(") {
		dataType = "integer"
	}

	if strings.Contains(fmt.Sprintf("%s", dataType), "longtext") {
		dataType = "text"
	}

	valueStr := ""
	notNullItem := defineMap[tagDefine.NOT_NULL]
	if notNullItem != nil {
		valueStr = " NOT NULL "
	} else {
		valueStr = " NULL "
	}

	if isPk {
		valueStr = " NOT NULL "
	}

	defaultValue, has := defineMap[tagDefine.DEFAULT]
	if has {
		valueStr = fmt.Sprintf("DEFAULT %s", defaultValue)
	}

	if action == "alter" {
		columnSqlList = append(columnSqlList, fmt.Sprintf("%s %s", dataType, valueStr))
	} else {
		s := fmt.Sprintf("\"%s\" %s %s", column, dataType, valueStr)
		if isPk {
			s += " PRIMARY KEY "
		}
		columnSqlList = append(columnSqlList, s)
	}

	_, has = defineMap[tagDefine.INDEX]
	if delIndexSql {
		indexSql = fmt.Sprintf("DROP INDEX \"idx_%s\"; ", column)
	}
	if has {
		indexSql += fmt.Sprintf("CREATE INDEX IF NOT EXISTS \"idx_%s\" ON \"%s\"(\"%s\"); ", column, tableName, column)
	}

	return strings.Join(columnSqlList, ""), indexSql
}

func (this *KingDataContext) CreateTableSQL(entity tiny.Entity) string {
	sql := this.DropTableSQL(entity.TableName())
	sql = this.FilterQuotes(sql)
	etype := reflect.TypeOf(entity).Elem()
	tableName := entity.TableName()

	columnSqlList := make([]string, 0)
	indexSqls := make([]string, 0)
	commentSqls := make([]string, 0)
	for i := 0; i < etype.NumField(); i++ {
		sField := etype.Field(i)

		defineStr, isTableColumn := this.GetFieldDefineStr(sField)
		if !isTableColumn {
			continue
		}

		defineMap := this.FormatDefine(defineStr)
		_, isMapping := defineMap[tagDefine.Mapping]
		if isMapping {
			continue
		}

		colSql, indexSql := this.GetColumnSqls(defineMap, sField.Name, "init", false, tableName)
		columnSqlList = append(columnSqlList, colSql)
		indexSqls = append(indexSqls, indexSql)

		if comment, ok := defineMap[tagDefine.COMMENT]; ok {
			commentSql := fmt.Sprintf("COMMENT ON COLUMN \"%s\".\"%s\" IS %s;", tableName, sField.Name, comment)
			commentSqls = append(commentSqls, commentSql)
		}
	}
	sql += fmt.Sprintf("CREATE TABLE \"%s\" ( %s );", tableName, strings.Join(columnSqlList, ","))
	sql += strings.Join(indexSqls, "")
	sql += strings.Join(commentSqls, "")
	return sql
}

func (this *KingDataContext) RegistModel(entity tiny.Entity) {
	this.DataContextBase.RegistModel(entity)
}

func (this *KingDataContext) GetEntityFieldsDefineInfo(entity interface{}) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	et := reflect.TypeOf(entity).Elem()
	for i := 0; i < et.NumField(); i++ {
		fd := et.Field(i)
		defineStr, has := this.GetFieldDefineStr(fd)
		if !has {
			continue
		}
		defineMap := this.FormatDefine(defineStr)

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

func (this *KingDataContext) AlterTableDropColumn(tableName string, columnName string) string {
	return fmt.Sprintf("ALTER TABLE \"%s\" DROP COLUMN \"%s\"; ", tableName, columnName)
}

func (this *KingDataContext) AlterTableAddColumn(tableName string, columnName string) string {
	return fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN IF NOT EXISTS %s; ", tableName, columnName)
}

func (this *KingDataContext) AlterTableAlterColumn(tableName string, oldColumnName string, newColumnName string, changeSql string) string {
	sql := ""
	if oldColumnName != newColumnName {
		sql = fmt.Sprintf("ALTER TABLE \"%s\" RENAME COLUMN \"%s\" TO \"%s\"; ", tableName, oldColumnName, newColumnName)
	}

	changeFieldItems := make([]string, 0)
	tmp := strings.Split(changeSql, " ")
	for i, v := range tmp {
		if i == 0 {
			changeFieldItems = append(changeFieldItems, fmt.Sprintf("ALTER COLUMN \"%s\" TYPE %s", oldColumnName, v))
		}

		if v == "NULL" {
			if tmp[i-1] == "NOT" {
				changeFieldItems = append(changeFieldItems, fmt.Sprintf("ALTER COLUMN \"%s\" SET NOT NULL", oldColumnName))
			} else {
				changeFieldItems = append(changeFieldItems, fmt.Sprintf("ALTER COLUMN \"%s\" DROP NOT NULL", oldColumnName))
			}
		}

		if v == "DEFAULT" {
			changeFieldItems = append(changeFieldItems, fmt.Sprintf("ALTER COLUMN \"%s\" SET DEFAULT %s", oldColumnName, tmp[i+1]))
		}
	}

	sql += fmt.Sprintf("ALTER TABLE \"%s\" %s; ", tableName, strings.Join(changeFieldItems, ","))

	return sql
}

func (this *KingDataContext) GetEntityList() map[string]tiny.Entity {
	return nil
}

func (this *KingDataContext) AddComments(entity tiny.Entity) {
	commentSqls := make([]string, 0)
	fieldsDefineInfo := this.GetEntityFieldsDefineInfo(entity)
	for fName, defineMap := range fieldsDefineInfo {
		if comment, ok := defineMap[tagDefine.COMMENT]; ok {
			commentSql := fmt.Sprintf("COMMENT ON COLUMN \"%s\".\"%s\" IS %s;", entity.TableName(), fName, comment)
			commentSqls = append(commentSqls, commentSql)
		}
	}
	sqlStr := strings.Join(commentSqls, "")
	if sqlStr != "" {
		rows, err := this.Db.Query(sqlStr)
		if err != nil {
			panic(err)
		}
		rows.Close()
	}

}
