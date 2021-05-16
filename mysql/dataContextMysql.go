package tinyMysql

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/joinlee/tiny-entity-go"
	"github.com/joinlee/tiny-entity-go/tagDefine"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlDataContext struct {
	db            *sql.DB
	interpreter   *tiny.Interpreter
	option        MysqlDataOption
	querySentence []string
	tx            *sql.Tx
	tranCount     int
	conStr        string
	entityRefMap  map[string]reflect.Type
}

func NewMysqlDataContext(opt MysqlDataOption) *MysqlDataContext {
	ctx := &MysqlDataContext{}

	conStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&multiStatements=true",
		opt.Username,
		opt.Password,
		opt.Host,
		opt.Port,
		opt.DataBaseName,
		opt.CharSet)
	ctx.conStr = conStr
	db, err := sql.Open("mysql", conStr)
	if err != nil {
		panic(err)
	}

	ctx.db = db

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(opt.ConnectionLimit)
	db.SetMaxIdleConns(20)

	ctx.interpreter = &tiny.Interpreter{}
	ctx.option = opt
	ctx.querySentence = make([]string, 0)
	ctx.tranCount = 0
	ctx.entityRefMap = make(map[string]reflect.Type)

	return ctx
}

//插入数据到数据库
func (this *MysqlDataContext) Create(entity tiny.Entity) {
	tableName := reflect.TypeOf(entity).Elem().Name()
	sql := fmt.Sprintf("INSERT INTO `%s`", tableName)
	fields, values, _ := this.getKeyValueList(entity, true)
	sql += fmt.Sprintf(" (%s) VALUES (%s);", strings.Join(fields, ","), strings.Join(values, ","))

	if this.tx == nil {
		this.submit(sql)
	} else {
		this.querySentence = append(this.querySentence, sql)
	}
}

//更新数据到数据库
func (this *MysqlDataContext) Update(entity tiny.Entity) {
	tableName := reflect.TypeOf(entity).Elem().Name()
	sql := fmt.Sprintf("UPDATE `%s` SET", tableName)
	_, _, kvMap := this.getKeyValueList(entity, false)

	vList := make([]string, 0)
	idValue := ""
	for k, v := range kvMap {
		if k == "Id" {
			idValue = v
			continue
		}
		vList = append(vList, fmt.Sprintf("`%s`=%s", k, v))
	}
	sql += strings.Join(vList, ",") + " WHERE `Id` = " + idValue + ";"

	if this.tx == nil {
		this.submit(sql)
	} else {
		this.querySentence = append(this.querySentence, sql)
	}
}

//通过实体Id 删除数据
func (this *MysqlDataContext) Delete(entity tiny.Entity) {
	tableName := reflect.TypeOf(entity).Elem().Name()
	_, _, kvMap := this.getKeyValueList(entity, false)

	sql := fmt.Sprintf("DELETE FROM `%s` WHERE `%s`.`Id`= '%s' ;", tableName, tableName, kvMap["Id"])

	if this.tx == nil {
		this.submit(sql)
	} else {
		this.querySentence = append(this.querySentence, sql)
	}
}

//通过指定条件删除数据
func (this *MysqlDataContext) DeleteWith(entity tiny.Entity, queryStr interface{}, args ...interface{}) {
	qs := queryStr.(string)
	for _, value := range args {
		qs = strings.Replace(qs, "?", this.interpreter.TransValueToStr(value), 1)
	}
	qs = this.interpreter.FormatQuerySetence(qs, "")

	tableName := reflect.TypeOf(entity).Elem().Name()
	sql := fmt.Sprintf("DELETE FROM `%s` WHERE %s ;", tableName, qs)

	if this.tx == nil {
		this.submit(sql)
	} else {
		this.querySentence = append(this.querySentence, sql)
	}
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
		panic(err)
	}

	sql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET %s COLLATE utf8_unicode_ci;", this.option.DataBaseName, this.option.CharSet)
	tiny.Log(sql)
	_, err1 := db.Exec(sql)
	if err1 != nil {
		panic(err)
	}
}

func (this *MysqlDataContext) DeleteDatabase() {

}

func (this *MysqlDataContext) CreateTable(entity tiny.Entity) {
	sqlStr := this.CreateTableSQL(entity)
	_, err := this.db.Query(sqlStr)
	if err != nil {
		panic(err)
	}
}

func (this *MysqlDataContext) CreateTableSQL(entity tiny.Entity) string {
	sql := this.DropTableSQL(entity.TableName())
	etype := reflect.TypeOf(entity).Elem()

	columnSqlList := make([]string, 0)
	for i := 0; i < etype.NumField(); i++ {
		sField := etype.Field(i)

		defineStr, isTableColumn := this.interpreter.GetFieldDefineStr(sField)
		if !isTableColumn {
			continue
		}

		defineMap := this.interpreter.FormatDefine(defineStr)
		_, isMapping := defineMap[tagDefine.Mapping]
		if isMapping {
			continue
		}

		columnSqlList = append(columnSqlList, this.interpreter.GetColumnSqls(defineMap, sField.Name, "init", false))
	}
	sql += fmt.Sprintf("CREATE TABLE `%s` ( %s );", entity.TableName(), strings.Join(columnSqlList, ","))
	return sql
}

func (this *MysqlDataContext) DropTableSQL(tableName string) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS `%s`; \n", tableName)
}

func (this *MysqlDataContext) Commit() {
	if this.tx != nil {
		if this.tranCount > 1 {
			this.tranCount--
		} else if this.tranCount == 1 {
			this.tranCount = 0
			tiny.Log(strings.Join(this.querySentence, ""))
			_, err := this.tx.Query(strings.Join(this.querySentence, ""))
			if err != nil {
				panic(err)
			}
			err = this.tx.Commit()
			this.cleanTransactionStatus()
			if err != nil {
				panic(err)
			}
		}
	} else {
		tiny.Log(strings.Join(this.querySentence, "\n"))
		_, err := this.db.Query(strings.Join(this.querySentence, "\n"))
		this.db.Close()
		if err != nil {
			panic(err)
		}
	}
}

func (this *MysqlDataContext) submit(sqlStr string) {
	// this.querySentence = append(this.querySentence, sqlStr)
	// strings.Join(this.querySentence, "\n")
	_, err := this.db.Query(sqlStr)
	this.db.Close()
	if err != nil {
		panic(err)
	}
	tiny.Log(sqlStr)
}

func (this *MysqlDataContext) cleanTransactionStatus() {
	this.querySentence = make([]string, 0)
	this.tranCount = 0
}

func (this *MysqlDataContext) BeginTranscation() {
	if this.tx == nil {
		tx, err := this.db.Begin()
		if err != nil {
			panic(err)
		}
		this.tx = tx
	}

	this.tranCount++
}

func (this *MysqlDataContext) RollBack() {
	if this.tx != nil {
		this.tranCount = 0
		this.tx.Rollback()
		this.cleanTransactionStatus()
	}
}

func (this *MysqlDataContext) Query(sqlStr string, noCommit bool) map[int]map[string]string {
	if noCommit {
		this.querySentence = append(this.querySentence, sqlStr)
		return nil
	} else {
		db, err := sql.Open("mysql", this.conStr)
		if err != nil {
			panic(err)
		}

		tiny.Log(sqlStr)
		rows, err := db.Query(sqlStr)
		if err != nil {
			panic(err)
		}

		//返回所有列
		cols, _ := rows.Columns()
		//这里表示一行所有列的值，用[]byte表示
		vals := make([][]byte, len(cols))
		//这里表示一行填充数据
		scans := make([]interface{}, len(cols))
		//这里scans引用vals，把数据填充到[]byte里
		for k := range vals {
			scans[k] = &vals[k]
		}
		i := 0
		result := make(map[int]map[string]string)

		for rows.Next() {
			//填充数据
			rows.Scan(scans...)
			//每行数据
			row := make(map[string]string)
			//把vals中的数据复制到row中
			for k, v := range vals {
				key := cols[k]
				//这里把[]byte数据转成string
				row[key] = string(v)
			}
			//放入结果集
			result[i] = row
			i++
		}

		return result
	}
}

func (this *MysqlDataContext) RegistModel(entity tiny.Entity) {
	t := reflect.TypeOf(entity).Elem()
	this.entityRefMap[t.Name()] = t
}

func (this *MysqlDataContext) GetEntityInstance(entityName string) interface{} {
	entityType, ok := this.entityRefMap[entityName]
	if !ok {
		return nil
	}

	return reflect.New(entityType).Elem().Interface()
}

func (this *MysqlDataContext) getKeyValueList(entity tiny.Entity, includeNilValue bool) ([]string, []string, map[string]string) {
	etype := reflect.TypeOf(entity).Elem()
	evalue := reflect.ValueOf(entity).Elem()
	fields := make([]string, 0)
	values := make([]string, 0)

	kvMap := make(map[string]string)

	for i := 0; i < etype.NumField(); i++ {
		sField := etype.Field(i)
		defineStr, has := this.interpreter.GetFieldDefineStr(sField)
		if !has {
			continue
		}

		vi := evalue.Field(i)
		value := vi.Interface()
		if evalue.Field(i).Kind() == reflect.Ptr && !vi.IsNil() {
			value = evalue.Field(i).Elem().Interface()
		}
		vStr := this.interpreter.TransValueToStr(value)

		defineMap := this.interpreter.FormatDefine(defineStr)
		dataType := defineMap[tagDefine.Type]
		if dataType != nil && strings.Index(dataType.(string), "varchar") > 0 {
			vStr = fmt.Sprintf("'%s'", vStr)
		}

		columnName, has := defineMap[tagDefine.Column]
		if !has {
			columnName = sField.Name
		}

		_, isMapping := defineMap[tagDefine.Mapping]
		if isMapping {
			continue
		}

		if includeNilValue {
			values = append(values, vStr)
			fields = append(fields, fmt.Sprintf("`%s`", columnName))
		} else {
			if vStr != "NULL" {
				values = append(values, vStr)
				fields = append(fields, fmt.Sprintf("`%s`", columnName))
			}
		}

		kvMap[fmt.Sprintf("%s", columnName)] = vStr
	}

	return fields, values, kvMap
}

type MysqlDataOption struct {
	Host            string
	Port            string
	Username        string
	Password        string
	DataBaseName    string
	CharSet         string
	ConnectionLimit int
}
