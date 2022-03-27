package tiny

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/joinlee/tiny-entity-go/tagDefine"
)

type InterpreterKing struct {
	tableName   string
	selectStrs  []string
	whereStrs   []string
	orderByStrs []string
	groupByStrs []string
	limt        map[string]int
	joinOnPart  []string
	AESKey      string
}

func NewInterpreterKing(tableName string) *InterpreterKing {
	interpreter := &InterpreterKing{}
	interpreter.whereStrs = make([]string, 0)
	interpreter.selectStrs = make([]string, 0)
	interpreter.orderByStrs = make([]string, 0)
	interpreter.groupByStrs = make([]string, 0)
	interpreter.limt = make(map[string]int)
	interpreter.joinOnPart = make([]string, 0)
	interpreter.tableName = tableName
	// interpreter.AESKey = AESKey
	return interpreter
}

func (this *InterpreterKing) AddToSelect(fields []string) {
	this.selectStrs = append(this.selectStrs, strings.Join(fields, ","))
}

func (this *InterpreterKing) CleanSelectPart() {
	this.selectStrs = make([]string, 0)
}
func (this *InterpreterKing) Clean() {
	this.whereStrs = make([]string, 0)
	this.selectStrs = make([]string, 0)
	this.groupByStrs = make([]string, 0)
	this.orderByStrs = make([]string, 0)
	this.limt = make(map[string]int)
	this.joinOnPart = make([]string, 0)
}

func (this *InterpreterKing) GetSelectFieldList(entity Entity, tableName string) []string {
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

		list = append(list, fmt.Sprintf("%s AS \"%s_%s\"", this.AddFieldTableName(cName, tableName), tableName, cName))
	}

	return list
}

func (this *InterpreterKing) AddToWhere(sql string, brackets bool) {
	if brackets {
		this.whereStrs = append(this.whereStrs, fmt.Sprintf("(%s)", sql))
	} else {
		this.whereStrs = append(this.whereStrs, sql)
	}

}

func (this *InterpreterKing) AddToOrdereBy(field string, isDesc bool) {
	field = this.AddFieldTableName(field, this.tableName)
	if isDesc {
		field += " DESC"
	}
	this.orderByStrs = append(this.orderByStrs, field)
}

func (this *InterpreterKing) AddToGroupBy(field string) {
	this.groupByStrs = append(this.groupByStrs, field)
}

func (this *InterpreterKing) AddToLimt(key string, value int) {
	this.limt[key] = value
}

func (this *InterpreterKing) AddToJoinOn(sqlStr string) {
	this.joinOnPart = append(this.joinOnPart, sqlStr)
}

func (this *InterpreterKing) AddFieldTableName(field interface{}, tableName string) string {
	return fmt.Sprintf("\"%s\".\"%s\"", tableName, field)
}

func (this *InterpreterKing) GetFinalSql(tableName string, entity Entity) string {
	var sql string
	sql += "SELECT "
	if len(this.selectStrs) == 0 {
		fields := this.GetSelectFieldList(entity, tableName)
		this.AddToSelect(fields)
	}
	sql += strings.Join(this.selectStrs, ",")
	sql += " FROM \"" + tableName + "\""
	if len(this.joinOnPart) > 0 {
		sql += strings.Join(this.joinOnPart, " ")
	}

	if len(this.whereStrs) > 0 {
		tmp := make([]string, 0)
		for index, wsql := range this.whereStrs {
			tmp = append(tmp, wsql)
			if wsql == "AND" || wsql == "OR" {
				continue
			}
			if index >= len(this.whereStrs)-1 {
				break
			}
			nextSql := this.whereStrs[index+1]
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

func (this *InterpreterKing) TransValueToStr(v interface{}) string {
	valueType := reflect.TypeOf(v)
	vtStr := valueType.Name()

	return this.TransValueToStrByType(v, vtStr)
}
func (this *InterpreterKing) TransValueToStrByType(v interface{}, typeName string) string {
	result := "NULL"
	if typeName == "string" {
		result = "'" + this.FormatSQL(fmt.Sprintf("%s", v)) + "'"
	} else if typeName == "int" || typeName == "int64" {
		result = fmt.Sprintf("%d", v)
	} else if typeName == "bool" {
		result = strconv.FormatBool(v.(bool))
	}
	return result
}

func (this *InterpreterKing) ConverNilValue(fieldType string, value string) interface{} {
	if fieldType == "*int" || fieldType == "*int64" {
		v, _ := strconv.Atoi(value)
		if value == "" {
			return nil
		}
		return &v
	} else if fieldType == "*bool" {
		v, _ := strconv.ParseBool(value)
		if value == "" {
			return nil
		}
		return &v
	} else if fieldType == "*string" {
		if value == "" {
			return nil
		}
		return &value
	} else if fieldType == "int" || fieldType == "int64" {
		v, _ := strconv.Atoi(value)
		return v
	} else if fieldType == "bool" {
		v, _ := strconv.ParseBool(value)
		return v
	}
	return value
}

func (this *InterpreterKing) FormatDefine(defineStr string) map[string]interface{} {
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

func (this *InterpreterKing) GetFieldDefineStr(field reflect.StructField) (string, bool) {
	defineStr, has := field.Tag.Lookup("tiny")
	return defineStr, has
}

func (this *InterpreterKing) GetEntityFieldsDefineInfo(entity interface{}) map[string]map[string]interface{} {
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

func (this *InterpreterKing) GetColumnSqls(defineMap map[string]interface{}, fieldName string, action string, delIndexSql bool, tableName string) (columnSql string, indexSql string) {
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
		valueStr = "NOT NULL "
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

func (this *InterpreterKing) FormatQuerySetence(qs string, tableName string) string {
	qsList := strings.Split(qs, " ")
	for i, s := range qsList {
		if s == "=" || s == ">" || s == "<" || s == ">=" || s == "<=" || s == "LIKE" || s == "IS" || s == "!=" || s == "IN" {
			index := i - 1
			if qsList[index] == "NOT" {
				index -= 1
			}
			qsList[index] = fmt.Sprintf("\"%s\".\"%s\"", tableName, qsList[index])
		}
	}

	qs = strings.Join(qsList, " ")
	return qs
}

func (this *InterpreterKing) FormatSQL(sql string) string {
	sql = strings.ReplaceAll(sql, "\\", "")
	sql = strings.ReplaceAll(sql, "'", "\\'")
	return sql
}

func (this *InterpreterKing) AlterTableDropColumn(tableName string, columnName string) string {
	return fmt.Sprintf("ALTER TABLE \"%s\" DROP COLUMN \"%s\"; ", tableName, columnName)
}

func (this *InterpreterKing) AlterTableAddColumn(tableName string, columnName string) string {
	return fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN \"%s\"; ", tableName, columnName)
}

func (this *InterpreterKing) AlterTableAlterColumn(tableName string, oldColumnName string, newColumnName string, changeSql string) string {
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

	sql += fmt.Sprintf("ALTER TABLE \"%s\" %s", tableName, strings.Join(changeFieldItems, ","))

	return sql
}

func (this *InterpreterKing) AesEncrypt(orig string, key string) string {
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
func (this *InterpreterKing) AesDecrypt(cryted string, key string) string {
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

func (this *InterpreterKing) PKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//去码
func (this *InterpreterKing) PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
