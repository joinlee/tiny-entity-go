package tiny

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/joinlee/tiny-entity-go/tagDefine"
)

type DataContextOptions struct {
	Host            string
	Port            string
	Username        string
	Password        string
	DataBaseName    string
	CharSet         string
	ConnectionLimit int
}

type DataContextBase struct {
	Db     *sql.DB
	Option DataContextOptions
	AESKey string

	selectStrs   []string
	WhereStrs    []string
	orderByStrs  []string
	groupByStrs  []string
	limt         map[string]int
	joinOnPart   []string
	tranCount    int
	tx           *sql.Tx
	entityRefMap map[string]reflect.Type
}

func NewDataContextBase(opt DataContextOptions) *DataContextBase {
	obj := new(DataContextBase)
	obj.Option = opt
	obj.AESKey = "53798C8E9F68B02F82E892F64F5DEF8B"
	obj.tranCount = 0

	obj.WhereStrs = make([]string, 0)
	obj.selectStrs = make([]string, 0)
	obj.orderByStrs = make([]string, 0)
	obj.groupByStrs = make([]string, 0)
	obj.limt = make(map[string]int, 0)
	obj.joinOnPart = make([]string, 0)
	obj.entityRefMap = make(map[string]reflect.Type)
	return obj
}

func (this *DataContextBase) AddToSelect(fields []string) {
	this.selectStrs = append(this.selectStrs, strings.Join(fields, ","))
}

// brackets 是否添加括弧
func (this *DataContextBase) AddToWhere(sql string, brackets bool) {
	if brackets {
		this.WhereStrs = append(this.WhereStrs, fmt.Sprintf("(%s)", sql))
	} else {
		this.WhereStrs = append(this.WhereStrs, sql)
	}
}

func (this *DataContextBase) AddToOrdereBy(field string, isDesc bool, tableName string) {
	field = fmt.Sprintf("`%s`.`%s`", tableName, field)
	if isDesc {
		field += " DESC"
	}
	this.orderByStrs = append(this.orderByStrs, field)
}

func (this *DataContextBase) AddToGroupBy(field string, tableName string) {
	field = fmt.Sprintf("`%s`.`%s`", tableName, field)
	this.groupByStrs = append(this.groupByStrs, field)
}

func (this *DataContextBase) AddToLimt(key string, value int) {
	this.limt[key] = value
}

func (this *DataContextBase) AddToJoinOn(sqlStr string) {
	this.joinOnPart = append(this.joinOnPart, sqlStr)
}

func (this *DataContextBase) GetSelectFieldList(entity Entity, tableName string) []string {
	list := make([]string, 0)
	etype := reflect.TypeOf(entity)
	if etype.Kind() == reflect.Ptr {
		etype = etype.Elem()
	}
	for i := 0; i < etype.NumField(); i++ {
		fd := etype.Field(i)
		cName := fd.Name
		defineStr, has := this.GetFieldDefineStr(fd)
		if !has {
			continue
		}

		defineMap := this.FormatDefine(defineStr)
		_, has = defineMap[tagDefine.Mapping]
		if has {
			continue
		}

		list = append(list, fmt.Sprintf("%s AS `%s_%s`", fmt.Sprintf("`%s`.`%s`", tableName, cName), tableName, cName))
	}

	return list
}

func (this *DataContextBase) GetFinalSql(tableName string, entity Entity) string {
	var sql string
	sql += "SELECT "
	if len(this.selectStrs) == 0 {
		fields := this.GetSelectFieldList(entity, tableName)
		this.AddToSelect(fields)
	}
	sql += strings.Join(this.selectStrs, ",")
	sql += " FROM `" + tableName + "`"
	if len(this.joinOnPart) > 0 {
		sql += strings.Join(this.joinOnPart, " ")
	}

	if len(this.WhereStrs) > 0 {
		tmp := make([]string, 0)
		for index, wsql := range this.WhereStrs {
			tmp = append(tmp, wsql)
			if wsql == "AND" || wsql == "OR" {
				continue
			}
			if index >= len(this.WhereStrs)-1 {
				break
			}
			nextSql := this.WhereStrs[index+1]
			if nextSql == "AND" || nextSql == "OR" {
				continue
			}
			tmp = append(tmp, "AND")
		}

		sql += " WHERE " + strings.Join(tmp, " ")
	}
	if len(this.groupByStrs) > 0 {
		sql += fmt.Sprintf(" GROUP BY %s", strings.Join(this.groupByStrs, ","))
	}
	if len(this.orderByStrs) > 0 {
		sql += fmt.Sprintf(" ORDER BY %s", strings.Join(this.orderByStrs, ","))
	}
	if len(this.limt) > 0 {
		if len(this.limt) > 1 {
			sql += fmt.Sprintf(" LIMIT %d,%d", this.limt["skip"], this.limt["take"])
		} else {
			sql += fmt.Sprintf(" LIMIT %d", this.limt["take"])
		}
	}

	sql += ";"

	return sql
}

func (this *DataContextBase) CleanSelectPart() {
	this.selectStrs = make([]string, 0)
}
func (this *DataContextBase) Clean() {
	this.WhereStrs = make([]string, 0)
	this.selectStrs = make([]string, 0)
	this.groupByStrs = make([]string, 0)
	this.orderByStrs = make([]string, 0)
	this.limt = make(map[string]int)
	this.joinOnPart = make([]string, 0)
}

//插入数据到数据库
func (this *DataContextBase) CreateSql(entity Entity) string {
	tableName := this.GetEntityName(entity)
	fields, values, _ := this.getKeyValueList(entity, true)

	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s);", tableName, strings.Join(fields, ","), strings.Join(values, ","))

	return sql
}

func (this *DataContextBase) CreateBatchSql(entities []Entity) string {
	if len(entities) > 0 {
		tableName := this.GetEntityName(entities[0])
		fieldPart := ""
		valueStrs := make([]string, 0)

		for _, entity := range entities {
			tfields, values, _ := this.getKeyValueList(entity, true)
			if fieldPart == "" {
				fieldPart = strings.Join(tfields, ",")
			}
			valueStrs = append(valueStrs, fmt.Sprintf("(%s)", strings.Join(values, ",")))
		}

		sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s;", tableName, fieldPart, strings.Join(valueStrs, ","))

		return sql
	}

	return ""
}

func (this *DataContextBase) UpdateSql(entity Entity) string {
	tableName := this.GetEntityName(entity)
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

	sql := fmt.Sprintf("UPDATE `%s` SET %s WHERE `Id` = %s;", tableName, strings.Join(vList, ","), idValue)

	return sql
}

//批量更新数据表中的数据
//entity 实体对象
//fields 需要更新的字段列表，传入参数例子：[ Username = 'lkc', age = 18 ]
//queryStr 条件参数 例子：gender = 'male'
func (this *DataContextBase) UpdateWithSql(entity Entity, fields interface{}, queryStr interface{}) string {
	tableName := this.GetEntityName(entity)
	fds := fields.([]string)
	fdsAfter := make([]string, 0)
	for _, v := range fds {
		fdsAfter = append(fdsAfter, this.AddFieldTableName(v, tableName))
	}
	qs := queryStr.(string)
	qs = this.AddFieldTableName(qs, tableName)

	sql := fmt.Sprintf("UPDATE `%s` SET %s WHERE %s ;", tableName, strings.Join(fdsAfter, ","), qs)

	return sql
}

//通过实体Id 删除数据
func (this *DataContextBase) DeleteSql(entity Entity) string {
	tableName := this.GetEntityName(entity)
	_, _, kvMap := this.getKeyValueList(entity, false)

	sql := fmt.Sprintf("DELETE FROM `%s` WHERE `%s`.`Id`= %s ;", tableName, tableName, kvMap["Id"])

	return sql
}

//通过指定条件删除数据
func (this *DataContextBase) DeleteWithSql(entity Entity, queryStr interface{}, args ...interface{}) string {
	qs := queryStr.(string)
	tableName := this.GetEntityName(entity)
	if qs != "" {
		for _, value := range args {
			qs = strings.Replace(qs, "?", this.TransValueToStr(value), 1)
		}
		qs = this.AddFieldTableName(qs, tableName)
		qs = "WHERE " + qs
	}

	sql := fmt.Sprintf("DELETE FROM `%s` %s;", tableName, qs)

	return sql
}

func (this *DataContextBase) AddFieldTableName(qs string, tableName string) string {
	qsList := strings.Split(qs, " ")
	for i, s := range qsList {
		if s == "=" || s == ">" || s == "<" || s == ">=" || s == "<=" || s == "LIKE" || s == "IS" || s == "!=" || s == "IN" {
			index := i - 1
			if qsList[index] == "NOT" {
				index -= 1
			}
			qsList[index] = fmt.Sprintf("`%s`.`%s`", tableName, qsList[index])
		}
	}

	qs = strings.Join(qsList, " ")
	return qs
}

// 将值转化成拼装Sql 需要的字符串
func (this *DataContextBase) TransValueToStr(v interface{}) string {
	valueType := reflect.TypeOf(v)
	vtStr := valueType.Name()

	return this.TransValueToStrByType(v, vtStr)
}

// 将值转化成拼装Sql 需要的字符串
func (t *DataContextBase) TransValueToStrByType(v interface{}, typeName string) string {
	result := "NULL"
	if typeName == "string" {
		result = "'" + t.FilterSpecialChar(fmt.Sprintf("%s", v)) + "'"
	} else if typeName == "int" || typeName == "int64" || typeName == "float" || typeName == "float32" || typeName == "float64" {
		result = fmt.Sprintf("%d", v)
	} else if typeName == "bool" {
		result = strconv.FormatBool(v.(bool))
	}
	return result
}

// 这里是过滤特殊字符
func (t *DataContextBase) FilterSpecialChar(sql string) string {
	sql = strings.ReplaceAll(sql, "\\", "")
	sql = strings.ReplaceAll(sql, "'", "\\'")
	return sql
}

func (this *DataContextBase) CreateDatabaseSQL() string {
	sql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET %s COLLATE utf8_unicode_ci;", this.Option.DataBaseName, this.Option.CharSet)
	return sql
}

func (this *DataContextBase) DeleteDatabaseSQL() string {
	return ""
}

func (this *DataContextBase) CreateTableSQL(entity Entity) string {
	sql := this.DropTableSQL(entity.TableName())
	etype := reflect.TypeOf(entity).Elem()

	columnSqlList := make([]string, 0)
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
		colStr, _ := this.GetColumnSqls(defineMap, sField.Name, "init", false, "")
		columnSqlList = append(columnSqlList, colStr)
	}
	sql += fmt.Sprintf("CREATE TABLE `%s` ( %s );", entity.TableName(), strings.Join(columnSqlList, ","))
	return sql
}

func (this *DataContextBase) DropTableSQL(tableName string) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS `%s`; \n", tableName)
}

func (this *DataContextBase) BeginTranscation() {
	if this.tx == nil {
		tx, err := this.Db.Begin()
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		this.tx = tx
	}

	this.tranCount++
}

func (this *DataContextBase) Commit() {
	if this.tranCount > 1 {
		this.tranCount--
	} else if this.tranCount == 1 {
		err := this.tx.Commit()
		this.resetTransactionStatus()
		if err != nil {
			panic(err)
		}
	}
}

func (this *DataContextBase) RollBack() {
	if this.tx != nil {
		this.tranCount = 0
		this.tx.Rollback()
		this.resetTransactionStatus()
	}
}

func (this *DataContextBase) Query(sqlStr string) map[int]map[string]string {
	var rows *sql.Rows
	var err error
	Log(sqlStr)
	if this.tx != nil {
		rows, err = this.tx.Query(sqlStr)
	} else {
		rows, err = this.Db.Query(sqlStr)
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

func (this *DataContextBase) GetFieldDefineStr(field reflect.StructField) (string, bool) {
	defineStr, has := field.Tag.Lookup("tiny")
	return defineStr, has
}

func (this *DataContextBase) FormatDefine(defineStr string) map[string]interface{} {
	list := strings.Split(defineStr, ";")
	keyMap := make(map[string]interface{})
	for _, item := range list {
		value := strings.Split(item, ":")
		length := len(value)
		if length == 1 {
			keyMap[value[0]] = ""
		} else if length == 2 {
			keyMap[value[0]] = value[1]
		}
	}

	return keyMap
}

func (this *DataContextBase) GetColumnSqls(defineMap map[string]interface{}, fieldName string, action string, delIndexSql bool, tableName string) (columnSql string, indexSql string) {
	columnSqlList := make([]string, 0)
	_, isPk := defineMap[tagDefine.PRIMARY_KEY]
	column, has := defineMap[tagDefine.Column]
	if !has {
		column = fieldName
	}
	if isPk {
		columnSqlList = append(columnSqlList, fmt.Sprintf("PRIMARY KEY (`%s`), ", column))
	}
	dataType, has := defineMap[tagDefine.Type]
	if !has {
		dataType = "varchar(255)"
		if isPk {
			dataType = "varchar(32)"
		}
	}

	valueStr := ""
	notNullItem := defineMap[tagDefine.NOT_NULL]
	if notNullItem != nil {
		valueStr = "NOT NULL"
	} else {
		valueStr = "NULL"
	}

	if isPk {
		valueStr = "NOT NULL"
	}

	defaultValue, has := defineMap[tagDefine.DEFAULT]
	if has {
		valueStr = fmt.Sprintf("DEFAULT %s", defaultValue)
	}

	if action == "alter" {
		columnSqlList = append(columnSqlList, fmt.Sprintf("%s %s", dataType, valueStr))
	} else {
		columnSqlList = append(columnSqlList, fmt.Sprintf("`%s` %s %s", column, dataType, valueStr))
	}

	_, has = defineMap[tagDefine.INDEX]
	if delIndexSql {
		columnSqlList = append(columnSqlList, fmt.Sprintf(",DROP INDEX idx_%s", column))
	}
	if has {
		if action == "init" {
			columnSqlList = append(columnSqlList, fmt.Sprintf(",KEY idx_%s (`%s`) USING BTREE ", column, column))
		}

		if action == "add" {
			columnSqlList = append(columnSqlList, fmt.Sprintf(",Add INDEX idx_%s (`%s`) USING BTREE", column, column))
		}

		if action == "alter" {
			columnSqlList = append(columnSqlList, fmt.Sprintf(",Add INDEX idx_%s (`%s`) USING BTREE", column, column))
		}
	}

	return strings.Join(columnSqlList, ""), ""
}

func (this *DataContextBase) AesEncrypt(orig string, key string) string {
	// 转成字节数组
	origData := []byte(orig)
	k := []byte(key)
	// 分组秘钥
	// NewCipher该函数限制了输入k的长度必须为16, 24或者32
	block, _ := aes.NewCipher(k)
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 补全码
	origData = this.PKCS7Padding(origData, blockSize)
	// 加密模式
	blockMode := cipher.NewCBCEncrypter(block, k[:blockSize])
	// 创建数组
	cryted := make([]byte, len(origData))
	// 加密
	blockMode.CryptBlocks(cryted, origData)
	return base64.StdEncoding.EncodeToString(cryted)
}

func (this *DataContextBase) AesDecrypt(cryted string, key string) string {
	// 转成字节数组
	crytedByte, _ := base64.StdEncoding.DecodeString(cryted)
	k := []byte(key)
	// 分组秘钥
	block, _ := aes.NewCipher(k)
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 加密模式
	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
	// 创建数组
	orig := make([]byte, len(crytedByte))
	// 解密
	blockMode.CryptBlocks(orig, crytedByte)
	// 去补全码
	orig = this.PKCS7UnPadding(orig)
	return string(orig)
}

func (this *DataContextBase) PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func (this *DataContextBase) PKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (this *DataContextBase) ConverNilValue(fieldType string, value string) (v interface{}, isStr bool) {
	if fieldType == "*int" || fieldType == "*int64" {
		v, _ := strconv.Atoi(value)
		if value == "" {
			return nil, false
		}
		return &v, false
	} else if fieldType == "*bool" {
		v, _ := strconv.ParseBool(value)
		if value == "" {
			return nil, false
		}
		return &v, false
	} else if fieldType == "*string" {
		if value == "" {
			return nil, false
		}
		return &value, true
	} else if fieldType == "int" || fieldType == "int64" {
		v, _ := strconv.Atoi(value)
		return v, false
	} else if fieldType == "bool" {
		v, _ := strconv.ParseBool(value)
		return v, false
	}
	return value, true
}

func (this *DataContextBase) GetEntityName(entity Entity) string {
	tableName := ""
	if reflect.TypeOf(entity).Kind() == reflect.Ptr {
		tableName = reflect.TypeOf(entity).Elem().Name()
	} else {
		tableName = reflect.TypeOf(entity).Name()
	}
	return tableName
}

func (this *DataContextBase) getKeyValueList(entity Entity, includeNilValue bool) ([]string, []string, map[string]string) {
	etype, evalue := this.getTypeAndValueRef(entity)
	fields := make([]string, 0)
	values := make([]string, 0)

	kvMap := make(map[string]string)

	for i := 0; i < etype.NumField(); i++ {
		sField := etype.Field(i)
		defineStr, has := this.GetFieldDefineStr(sField)
		if !has {
			continue
		}

		vi := evalue.Field(i)
		value := vi.Interface()
		if evalue.Field(i).Kind() == reflect.Ptr && !vi.IsNil() {
			value = evalue.Field(i).Elem().Interface()
		}

		defineMap := this.FormatDefine(defineStr)
		_, isAES := defineMap[tagDefine.AES]
		if isAES {
			vv := this.TransValueToStr(value)
			if vv != "NULL" {
				value = this.AesEncrypt(value.(string), this.AESKey)
			}
		}

		vStr := this.TransValueToStr(value)
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

func (this *DataContextBase) getTypeAndValueRef(entity Entity) (etype reflect.Type, evalue reflect.Value) {
	if reflect.TypeOf(entity).Kind() == reflect.Ptr {
		etype = reflect.TypeOf(entity).Elem()
		evalue = reflect.ValueOf(entity).Elem()
	} else {
		etype = reflect.TypeOf(entity)
		evalue = reflect.ValueOf(entity)
	}

	return etype, evalue
}

func (this *DataContextBase) resetTransactionStatus() {
	this.tranCount = 0
}

func (this *DataContextBase) Submit(sql string) {
	Log(sql)
	var err error
	if this.tx == nil {
		_, err = this.Db.Exec(sql)

	} else {
		_, err = this.tx.Exec(sql)
	}

	if err != nil {
		panic(err)
	}
}

func (this *DataContextBase) RegistModel(entity Entity) {
	t := reflect.TypeOf(entity).Elem()
	this.entityRefMap[t.Name()] = t
}
