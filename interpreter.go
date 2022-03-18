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

var AESKey = "53798C8E9F68B02F82E892F64F5DEF8B"

type Interpreter struct {
	tableName   string
	selectStrs  []string
	whereStrs   []string
	orderByStrs []string
	groupByStrs []string
	limt        map[string]int
	joinOnPart  []string
	AESKey      string
}

func NewInterpreter(tableName string) *Interpreter {
	interpreter := &Interpreter{}
	interpreter.whereStrs = make([]string, 0)
	interpreter.selectStrs = make([]string, 0)
	interpreter.orderByStrs = make([]string, 0)
	interpreter.groupByStrs = make([]string, 0)
	interpreter.limt = make(map[string]int)
	interpreter.joinOnPart = make([]string, 0)
	interpreter.tableName = tableName
	interpreter.AESKey = AESKey
	return interpreter
}

func (t *Interpreter) AddToSelect(fields []string) {
	t.selectStrs = append(t.selectStrs, strings.Join(fields, ","))
}

func (t *Interpreter) CleanSelectPart() {
	t.selectStrs = make([]string, 0)
}
func (t *Interpreter) Clean() {
	t.whereStrs = make([]string, 0)
	t.selectStrs = make([]string, 0)
	t.groupByStrs = make([]string, 0)
	t.orderByStrs = make([]string, 0)
	t.limt = make(map[string]int)
	t.joinOnPart = make([]string, 0)
}

func (t *Interpreter) GetSelectFieldList(entity Entity, tableName string) []string {
	list := make([]string, 0)
	etype := reflect.TypeOf(entity)
	if etype.Kind() == reflect.Ptr {
		etype = etype.Elem()
	}
	for i := 0; i < etype.NumField(); i++ {
		fd := etype.Field(i)
		cName := fd.Name
		defineStr, has := t.GetFieldDefineStr(fd)
		if !has {
			continue
		}

		defineMap := t.FormatDefine(defineStr)
		_, has = defineMap[tagDefine.Mapping]
		if has {
			continue
		}

		list = append(list, fmt.Sprintf("%s AS %s_%s", t.AddFieldTableName(cName, tableName), tableName, cName))
	}

	return list
}

func (t *Interpreter) AddToWhere(sql string, brackets bool) {
	if brackets {
		t.whereStrs = append(t.whereStrs, fmt.Sprintf("(%s)", sql))
	} else {
		t.whereStrs = append(t.whereStrs, sql)
	}

}

func (t *Interpreter) AddToOrdereBy(field string, isDesc bool) {
	field = t.AddFieldTableName(field, t.tableName)
	if isDesc {
		field += " DESC"
	}
	t.orderByStrs = append(t.orderByStrs, field)
}

func (t *Interpreter) AddToGroupBy(field string) {
	field = t.AddFieldTableName(field, t.tableName)
	t.groupByStrs = append(t.groupByStrs, field)
}

func (t *Interpreter) AddToLimt(key string, value int) {
	t.limt[key] = value
}

func (t *Interpreter) AddToJoinOn(sqlStr string) {
	t.joinOnPart = append(t.joinOnPart, sqlStr)
}

func (t *Interpreter) AddFieldTableName(field interface{}, tableName string) string {
	return fmt.Sprintf("`%s`.`%s`", tableName, field)
}

func (t *Interpreter) GetFinalSql(tableName string, entity Entity) string {
	var sql string
	sql += "SELECT "
	if len(t.selectStrs) == 0 {
		fields := t.GetSelectFieldList(entity, tableName)
		t.AddToSelect(fields)
	}
	sql += strings.Join(t.selectStrs, ",")
	sql += " FROM `" + tableName + "`"
	if len(t.joinOnPart) > 0 {
		sql += strings.Join(t.joinOnPart, " ")
	}

	if len(t.whereStrs) > 0 {
		// sql += " WHERE " + strings.Join(t.whereStrs, " AND ")
		tmp := make([]string, 0)
		for index, wsql := range t.whereStrs {
			tmp = append(tmp, wsql)
			if wsql == "AND" || wsql == "OR" {
				continue
			}
			if index >= len(t.whereStrs)-1 {
				break
			}
			nextSql := t.whereStrs[index+1]
			if nextSql == "AND" || nextSql == "OR" {
				continue
			}
			tmp = append(tmp, "AND")
		}

		sql += " WHERE " + strings.Join(tmp, " ")
	}
	if len(t.groupByStrs) > 0 {
		sql += fmt.Sprintf(" GROUP BY %s", strings.Join(t.groupByStrs, ""))
	}
	if len(t.orderByStrs) > 0 {
		sql += fmt.Sprintf(" ORDER BY %s", strings.Join(t.orderByStrs, ","))
	}
	if len(t.limt) > 0 {
		if len(t.limt) > 1 {
			sql += fmt.Sprintf(" LIMIT %d,%d", t.limt["skip"], t.limt["take"])
		} else {
			sql += fmt.Sprintf(" LIMIT %d", t.limt["take"])
		}
	}

	sql += ";"

	return sql
}

func (t *Interpreter) TransValueToStr(v interface{}) string {
	valueType := reflect.TypeOf(v)
	vtStr := valueType.Name()

	return t.TransValueToStrByType(v, vtStr)
}
func (t *Interpreter) TransValueToStrByType(v interface{}, typeName string) string {
	result := "NULL"
	if typeName == "string" {
		result = "'" + t.FormatSQL(fmt.Sprintf("%s", v)) + "'"
	} else if typeName == "int" || typeName == "int64" {
		result = fmt.Sprintf("%d", v)
	} else if typeName == "bool" {
		result = strconv.FormatBool(v.(bool))
	}
	return result
}

func (t *Interpreter) ConverNilValue(fieldType string, value string) interface{} {
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

func (t *Interpreter) FormatDefine(defineStr string) map[string]interface{} {
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

func (t *Interpreter) GetFieldDefineStr(field reflect.StructField) (string, bool) {
	defineStr, has := field.Tag.Lookup("tiny")
	return defineStr, has
}

func (t *Interpreter) GetEntityFieldsDefineInfo(entity interface{}) map[string]map[string]interface{} {
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

func (t *Interpreter) GetColumnSqls(defineMap map[string]interface{}, fieldName string, action string, delIndexSql bool, tableName string) (columnSql string, indexSql string) {
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
func (t *Interpreter) FormatQuerySetence(qs string, tableName string) string {
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

func (t *Interpreter) FormatSQL(sql string) string {
	sql = strings.ReplaceAll(sql, "\\", "")
	sql = strings.ReplaceAll(sql, "'", "\\'")
	return sql
}

func (t *Interpreter) AlterTableDropColumn(tableName string, columnName string) string {
	return fmt.Sprintf("ALTER TABLE `%s` Drop `%s`; ", tableName, columnName)
}

func (t *Interpreter) AlterTableAddColumn(tableName string, columnName string) string {
	return fmt.Sprintf("ALTER TABLE `%s` Add %s; ", tableName, columnName)
}

func (t *Interpreter) AlterTableAlterColumn(tableName string, oldColumnName string, newColumnName string, changeSql string) string {
	return fmt.Sprintf("ALTER TABLE `%s` CHANGE `%s` `%s` %s; ", tableName, oldColumnName, newColumnName, changeSql)
}

func (t *Interpreter) AesEncrypt(orig string, key string) string {
	// 转成字节数组
	origData := []byte(orig)
	k := []byte(key)
	// 分组秘钥
	// NewCipher该函数限制了输入k的长度必须为16, 24或者32
	block, _ := aes.NewCipher(k)
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 补全码
	origData = t.PKCS7Padding(origData, blockSize)
	// 加密模式
	blockMode := cipher.NewCBCEncrypter(block, k[:blockSize])
	// 创建数组
	cryted := make([]byte, len(origData))
	// 加密
	blockMode.CryptBlocks(cryted, origData)
	return base64.StdEncoding.EncodeToString(cryted)
}
func (t *Interpreter) AesDecrypt(cryted string, key string) string {
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
	orig = t.PKCS7UnPadding(orig)
	return string(orig)
}

func (t *Interpreter) PKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//去码
func (t *Interpreter) PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func IntToPtr(value int) *int {
	return &value
}

func Int64ToPtr(v int64) *int64 {
	return &v
}

func BoolToPtr(v bool) *bool {
	return &v
}
