package tinyKing

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/joinlee/tiny-entity-go"
	tinyMysql "github.com/joinlee/tiny-entity-go/mysql"
	"github.com/joinlee/tiny-entity-go/tagDefine"

	// _ "kingbase.com/gokb"
	_ "github.com/joinlee/kingbase-driver"
)

type KingDataContext struct {
	db            *sql.DB
	interpreter   *tiny.InterpreterKing
	option        KingDataOption
	querySentence []string
	tx            *sql.Tx
	tranCount     int
	conStr        string
	entityRefMap  map[string]reflect.Type
}

func NewKingDataContext(opt KingDataOption) *KingDataContext {
	ctx := &KingDataContext{}
	conStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		opt.Host,
		opt.Username,
		opt.Password,
		opt.DataBaseName)
	ctx.conStr = conStr
	ctx.db = tinyMysql.GetDB(conStr, opt.ConnectionLimit, "kingbase")
	ctx.interpreter = &tiny.InterpreterKing{}
	ctx.interpreter.AESKey = tiny.AESKey
	ctx.option = opt
	ctx.querySentence = make([]string, 0)
	ctx.tranCount = 0
	ctx.entityRefMap = make(map[string]reflect.Type)
	return ctx
}

//插入数据到数据库
func (this *KingDataContext) Create(entity tiny.Entity) {
	tableName := this.getTableNameFromEntity(entity)
	sql := fmt.Sprintf("INSERT INTO \"%s\"", tableName)
	fields, values, _ := this.getKeyValueList(entity, true)
	sql += fmt.Sprintf(" (%s) VALUES (%s);", strings.Join(fields, ","), strings.Join(values, ","))

	if this.tx == nil {
		this.submit(sql, false)
	} else {
		this.querySentence = append(this.querySentence, sql)
		this.tx.Exec(sql)
	}
}

func (this *KingDataContext) CreateBatch(entities []tiny.Entity) {
	if len(entities) > 0 {
		tableName := this.getTableNameFromEntity(entities[0])

		sql := fmt.Sprintf("INSERT INTO \"%s\"", tableName)
		fieldPart := ""
		valueStrs := make([]string, 0)

		for _, entity := range entities {
			tfields, values, _ := this.getKeyValueList(entity, true)
			if fieldPart == "" {
				fieldPart = strings.Join(tfields, ",")
			}
			valueStrs = append(valueStrs, fmt.Sprintf("(%s)", strings.Join(values, ",")))
		}

		sql = fmt.Sprintf("%s (%s) VALUES %s;", sql, fieldPart, strings.Join(valueStrs, ","))

		if this.tx == nil {
			this.submit(sql, false)
		} else {
			this.querySentence = append(this.querySentence, sql)
			this.tx.Exec(sql)
		}
	}

}

//更新数据到数据库
func (this *KingDataContext) Update(entity tiny.Entity) {
	tableName := this.getTableNameFromEntity(entity)
	sql := fmt.Sprintf("UPDATE \"%s\" SET ", tableName)
	_, _, kvMap := this.getKeyValueList(entity, false)

	vList := make([]string, 0)
	idValue := ""
	for k, v := range kvMap {
		if k == "Id" {
			idValue = v
			continue
		}
		vList = append(vList, fmt.Sprintf("\"%s\"=%s", k, v))
	}
	sql += strings.Join(vList, ",") + " WHERE \"Id\" = " + idValue + ";"

	if this.tx == nil {
		this.submit(sql, false)
	} else {
		this.querySentence = append(this.querySentence, sql)
		this.tx.Exec(sql)
	}
}

//批量更新数据表中的数据
//entity 实体对象
//fields 需要更新的字段列表，传入参数例子：[ Username = 'lkc', age = 18 ]
//queryStr 条件参数 例子：gender = 'male'
func (this *KingDataContext) UpdateWith(entity tiny.Entity, fields interface{}, queryStr interface{}) {
	tableName := this.getTableNameFromEntity(entity)
	fds := fields.([]string)
	fdsAfter := make([]string, 0)
	for _, v := range fds {
		fdsAfter = append(fdsAfter, this.interpreter.FormatQuerySetence(v, tableName))
	}
	qs := queryStr.(string)
	qs = this.interpreter.FormatQuerySetence(qs, tableName)

	sql := fmt.Sprintf("UPDATE \"%s\" SET %s WHERE %s ;", tableName, strings.Join(fdsAfter, ","), qs)

	if this.tx == nil {
		this.submit(sql, false)
	} else {
		this.querySentence = append(this.querySentence, sql)
		this.tx.Exec(sql)
	}
}

//通过实体Id 删除数据
func (this *KingDataContext) Delete(entity tiny.Entity) {
	tableName := this.getTableNameFromEntity(entity)
	_, _, kvMap := this.getKeyValueList(entity, false)

	sql := fmt.Sprintf("DELETE FROM \"%s\" WHERE \"%s\".\"Id\" = %s ;", tableName, tableName, kvMap["Id"])

	if this.tx == nil {
		this.submit(sql, false)
	} else {
		this.querySentence = append(this.querySentence, sql)
		this.tx.Exec(sql)
	}
}

//通过指定条件删除数据
func (this *KingDataContext) DeleteWith(entity tiny.Entity, queryStr interface{}, args ...interface{}) {
	tableName := this.getTableNameFromEntity(entity)
	qs := queryStr.(string)
	for _, value := range args {
		qs = strings.Replace(qs, "?", this.interpreter.TransValueToStr(value), 1)
	}
	qs = this.interpreter.FormatQuerySetence(qs, tableName)

	sql := fmt.Sprintf("DELETE FROM \"%s\" WHERE %s ;", tableName, qs)

	if this.tx == nil {
		this.submit(sql, false)
	} else {
		this.querySentence = append(this.querySentence, sql)
		this.tx.Exec(sql)
	}
}

func (this *KingDataContext) getTableNameFromEntity(entity tiny.Entity) string {
	tableName := ""
	if reflect.TypeOf(entity).Kind() == reflect.Ptr {
		tableName = reflect.TypeOf(entity).Elem().Name()
	} else {
		tableName = reflect.TypeOf(entity).Name()
	}
	return tableName
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
	tiny.Log(sqlStr)
	_, err := this.db.Exec(sqlStr)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func (this *KingDataContext) CreateTableSQL(entity tiny.Entity) string {
	sql := this.DropTableSQL(entity.TableName())
	etype := reflect.TypeOf(entity).Elem()
	tableName := entity.TableName()

	columnSqlList := make([]string, 0)
	indexSqls := make([]string, 0)
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

		colSql, indexSql := this.interpreter.GetColumnSqls(defineMap, sField.Name, "init", false, tableName)
		columnSqlList = append(columnSqlList, colSql)
		indexSqls = append(indexSqls, indexSql)
	}
	sql += fmt.Sprintf("CREATE TABLE \"%s\" ( %s );", entity.TableName(), strings.Join(columnSqlList, ","))
	sql += strings.Join(indexSqls, "")
	return sql
}

func (this *KingDataContext) DropTableSQL(tableName string) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS \"%s\"; ", tableName)
}

func (this *KingDataContext) Commit() {
	if this.tranCount > 1 {
		this.tranCount--
	} else if this.tranCount == 1 {
		tiny.Log(strings.Join(this.querySentence, ""))

		err := this.tx.Commit()
		this.cleanTransactionStatus()
		if err != nil {
			panic(err)
		}
	}
}

func (this *KingDataContext) submit(sqlStr string, isQuery bool) {
	tiny.Log(sqlStr)
	if isQuery {
		rows, err := this.db.Query(sqlStr)
		rows.Close()
		if err != nil {
			panic(err)
		}
	} else {
		_, err := this.db.Exec(sqlStr)
		if err != nil {
			panic(err)
		}
	}
}

func (this *KingDataContext) cleanTransactionStatus() {
	this.querySentence = make([]string, 0)
	this.tranCount = 0
}

func (this *KingDataContext) BeginTranscation() {
	if this.tx == nil {
		tx, err := this.db.Begin()
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		this.tx = tx
	}

	this.tranCount++
}

func (this *KingDataContext) RollBack() {
	if this.tx != nil {
		this.tranCount = 0
		this.tx.Rollback()
		this.cleanTransactionStatus()
	}
}

func (this *KingDataContext) Query(sqlStr string, noCommit bool) map[int]map[string]string {
	var rows *sql.Rows
	var err error
	tiny.Log(sqlStr)
	if this.tx != nil {
		rows, err = this.tx.Query(sqlStr)
	} else {
		rows, err = this.db.Query(sqlStr)
	}

	if err != nil {
		if rows != nil {
			rows.Close()
		}
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

	rows.Close()

	return result
}

func (this *KingDataContext) RegistModel(entity tiny.Entity) {
	t := reflect.TypeOf(entity).Elem()
	this.entityRefMap[t.Name()] = t
}

func (this *KingDataContext) GetEntityInstance(entityName string) interface{} {
	entityType, ok := this.entityRefMap[entityName]
	if !ok {
		return nil
	}

	return reflect.New(entityType).Elem().Interface()
}

func (this *KingDataContext) getTypeAndValueRef(entity tiny.Entity) (etype reflect.Type, evalue reflect.Value) {
	if reflect.TypeOf(entity).Kind() == reflect.Ptr {
		etype = reflect.TypeOf(entity).Elem()
		evalue = reflect.ValueOf(entity).Elem()
	} else {
		etype = reflect.TypeOf(entity)
		evalue = reflect.ValueOf(entity)
	}

	return etype, evalue
}

func (this *KingDataContext) getKeyValueList(entity tiny.Entity, includeNilValue bool) ([]string, []string, map[string]string) {
	etype, evalue := this.getTypeAndValueRef(entity)
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

		defineMap := this.interpreter.FormatDefine(defineStr)
		_, isAES := defineMap[tagDefine.AES]
		if isAES {
			vv := this.interpreter.TransValueToStr(value)
			if vv != "NULL" {
				value = this.interpreter.AesEncrypt(value.(string), this.interpreter.AESKey)
			}
		}

		vStr := this.interpreter.TransValueToStr(value)
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
			fields = append(fields, fmt.Sprintf("\"%s\"", columnName))
		} else {
			if vStr != "NULL" {
				values = append(values, vStr)
				fields = append(fields, fmt.Sprintf("\"%s\"", columnName))
			}
		}

		kvMap[fmt.Sprintf("%s", columnName)] = vStr
	}

	return fields, values, kvMap
}

type KingDataOption struct {
	Host            string
	Port            string
	Username        string
	Password        string
	DataBaseName    string
	CharSet         string
	ConnectionLimit int
}
