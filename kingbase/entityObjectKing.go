package tinyKing

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/joinlee/tiny-entity-go"
	tinyMysql "github.com/joinlee/tiny-entity-go/mysql"
	"github.com/joinlee/tiny-entity-go/tagDefine"
)

type EntityObjectKing struct {
	ctx         *KingDataContext
	interpreter *tiny.InterpreterKing

	tableName    string
	joinEntities map[string]tinyMysql.JoinEntityItem
}

func NewEntityObjectKing(ctx *KingDataContext, tableName string) *EntityObjectKing {
	entity := &EntityObjectKing{}
	entity.ctx = ctx
	entity.tableName = tableName
	entity.InitEntityObj(tableName)

	return entity
}

func (this *EntityObjectKing) InitEntityObj(tableName string) {
	this.interpreter = tiny.NewInterpreterKing(tableName)
	this.joinEntities = make(map[string]tinyMysql.JoinEntityItem)
}

func (this *EntityObjectKing) TableName() string {
	return this.tableName
}

func (this *EntityObjectKing) GetIQueryObject() tiny.IQueryObject {
	return this
}

func (this *EntityObjectKing) And() tiny.IQueryObject {
	this.interpreter.AddToWhere("AND", false)
	return this
}

func (this *EntityObjectKing) Or() tiny.IQueryObject {
	this.interpreter.AddToWhere("OR", false)
	return this
}

// 添加查询条件
/* queryStr 查询语句， args 条件参数
ex： ctx.User.Where("Id = ?", user.Id).Any() */
func (this *EntityObjectKing) Where(queryStr interface{}, args ...interface{}) tiny.IQueryObject {
	return this.wherePartHandle(this.tableName, queryStr, args)
}

//添加指定表的查询条件
/* entity 需要查询的实体 queryStr 查询语句， args 条件参数
entity 表示查询外键表的条件
ex： ctx.User.WhereWith(ctx.Account, "Id = ?", user.Id).Any() */
func (this *EntityObjectKing) WhereWith(entity tiny.Entity, queryStr interface{}, args ...interface{}) tiny.IQueryObject {
	tableName := reflect.TypeOf(entity).Elem().Name()
	return this.wherePartHandle(tableName, queryStr, args)
}

func (this *EntityObjectKing) Contains(felid string, values interface{}) tiny.IQueryObject {
	return this.inPartHandle(this.tableName, felid, values)
}

func (this *EntityObjectKing) ContainsWith(entity tiny.Entity, felid string, values interface{}) tiny.IQueryObject {
	tableName := reflect.TypeOf(entity).Elem().Name()
	return this.inPartHandle(tableName, felid, values)
}
func (this *EntityObjectKing) inPartHandle(tableName string, felid string, values interface{}) tiny.IQueryObject {
	qs := "\"" + tableName + "\".\"" + felid + "\" IN"
	vs := make([]string, 0)

	if reflect.TypeOf(values).Kind() == reflect.Slice {
		s := reflect.ValueOf(values)
		for i := 0; i < s.Len(); i++ {
			value := s.Index(i)
			vs = append(vs, this.interpreter.TransValueToStrByType(value, value.Kind().String()))
		}

		qs = qs + " ( " + strings.Join(vs, ",") + " )"
		this.interpreter.AddToWhere(qs, true)
	}
	return this
}

func (this *EntityObjectKing) wherePartHandle(tableName string, queryStr interface{}, args []interface{}) tiny.IQueryObject {
	if queryStr == nil || queryStr == "" {
		return this
	}
	qs := queryStr.(string)
	for _, value := range args {
		qs = strings.Replace(qs, "?", this.interpreter.TransValueToStr(value), 1)
	}
	qs = strings.ReplaceAll(qs, "`", "")
	qs = this.interpreter.FormatQuerySetence(qs, tableName)
	this.interpreter.AddToWhere(qs, true)
	return this
}

func (this *EntityObjectKing) OrderBy(field interface{}) tiny.IQueryObject {
	this.interpreter.AddToOrdereBy(field.(string), false)
	return this
}

func (this *EntityObjectKing) OrderByDesc(field interface{}) tiny.IQueryObject {
	this.interpreter.AddToOrdereBy(field.(string), true)
	return this
}

func (this *EntityObjectKing) IndexOf() tiny.IQueryObject {
	return this
}

func (this *EntityObjectKing) GroupBy(field interface{}) tiny.IResultQueryObject {
	fStr := field.(string)

	if len(this.joinEntities) > 0 {
		for k := range this.joinEntities {
			// 将所有外键表的Id加入到GroupBy 子句
			this.interpreter.AddToGroupBy(this.interpreter.AddFieldTableName("Id", k))
		}
	}
	// 将主键表加入到GroupBy 子句
	this.interpreter.AddToGroupBy(this.interpreter.AddFieldTableName("Id", this.tableName))

	if fStr != "Id" {
		this.interpreter.AddToGroupBy(this.interpreter.AddFieldTableName(fStr, this.tableName))
	}

	return this
}

func (this *EntityObjectKing) Select(fields ...interface{}) tiny.IResultQueryObject {
	list := make([]string, 0)
	for _, item := range fields {
		list = append(list, fmt.Sprintf("%s AS %s_%s", this.interpreter.AddFieldTableName(item.(string), this.tableName), this.tableName, item.(string)))
	}
	this.interpreter.CleanSelectPart()
	this.interpreter.AddToSelect(list)
	return this
}

func (this *EntityObjectKing) Take(count int) tiny.ITakeChildQueryObject {
	this.interpreter.AddToLimt("take", count)
	return this
}
func (this *EntityObjectKing) Skip(count int) tiny.IAssembleResultQuery {
	this.interpreter.AddToLimt("skip", count)
	return this
}

// 添加外联引用
/*
	fEntity 需要连接的实体， mField 主表的连接字段， fField 外联表的字段
*/
func (this *EntityObjectKing) JoinOn(fEntity tiny.Entity, mField string, fField string) tiny.IQueryObject {
	return this.joinHandle(this.tableName, fEntity, mField, fField)
}

func (this *EntityObjectKing) JoinOnWith(mEntity tiny.Entity, fEntity tiny.Entity, mField string, fField string) tiny.IQueryObject {
	mTableName := mEntity.TableName()
	return this.joinHandle(mTableName, fEntity, mField, fField)
}
func (this *EntityObjectKing) joinHandle(mTableName string, fEntity tiny.Entity, mField string, fField string) tiny.IQueryObject {
	if len(this.joinEntities) == 0 {
		mEntity := this.ctx.GetEntityInstance(mTableName)
		mainTableFields := this.interpreter.GetSelectFieldList(mEntity.(tiny.Entity), mTableName)
		this.interpreter.AddToSelect(mainTableFields)
	}

	fTableFields := this.interpreter.GetSelectFieldList(fEntity, fEntity.TableName())
	this.interpreter.AddToSelect(fTableFields)

	this.joinEntities[fEntity.TableName()] = tinyMysql.JoinEntityItem{
		Mkey:   mField,
		Fkey:   fField,
		Entity: fEntity,
	}
	sqlStr := fmt.Sprintf(" LEFT JOIN \"%s\" ON \"%s\".\"%s\" = \"%s\".\"%s\"", fEntity.TableName(), mTableName, mField, fEntity.TableName(), fField)
	this.interpreter.AddToJoinOn(sqlStr)
	return this
}

func (this *EntityObjectKing) Max() float64 {
	this.interpreter.Clean()
	return 0
}

func (this *EntityObjectKing) Min() float64 {
	this.interpreter.Clean()
	return 0
}
func (this *EntityObjectKing) Count() int {
	return this.CountArgs(fmt.Sprintf("\"%s\".\"Id\"", this.tableName))
}

func (this *EntityObjectKing) CountArgs(field string) int {
	this.interpreter.CleanSelectPart()
	field = strings.ReplaceAll(field, "`", "\"")
	this.interpreter.AddToSelect([]string{fmt.Sprintf("COUNT(%s)", field)})
	sqlStr := this.interpreter.GetFinalSql(this.tableName, nil)
	rows := this.ctx.Query(sqlStr, false)

	result := 0
	for _, rowData := range rows {
		for _, cellData := range rowData {
			result, _ = strconv.Atoi(cellData)
		}
	}
	this.interpreter.Clean()
	return result
}

func (this *EntityObjectKing) Any() bool {
	count := this.Count()
	return count > 0
}

func (this *EntityObjectKing) First(entity interface{}) (bool, *tiny.Empty) {
	mEntity := this.ctx.GetEntityInstance(this.tableName)
	sqlStr := this.interpreter.GetFinalSql(this.tableName, mEntity.(tiny.Entity))
	rows := this.ctx.Query(sqlStr, false)

	dataList := this.queryToDatas2(this.tableName, rows)

	isNull := false

	if len(dataList) > 0 {
		jsonStr := tiny.JsonStringify(dataList[0])
		json.Unmarshal([]byte(jsonStr), entity)
	} else {
		entity = nil
		isNull = true
	}

	this.InitEntityObj(this.tableName)
	this.interpreter.Clean()

	if isNull {
		return isNull, &tiny.Empty{}
	} else {
		return isNull, nil
	}
}

func (this *EntityObjectKing) ToList(list interface{}) {
	mEntity := this.ctx.GetEntityInstance(this.tableName)
	sqlStr := this.interpreter.GetFinalSql(this.tableName, mEntity.(tiny.Entity))
	rows := this.ctx.Query(sqlStr, false)
	dataList := this.queryToDatas2(this.tableName, rows)

	jsonStr := tiny.JsonStringify(dataList)
	json.Unmarshal([]byte(jsonStr), list)
	this.InitEntityObj(this.tableName)
	this.interpreter.Clean()
}

func (this *EntityObjectKing) queryToDatas2(tableName string, rows map[int]map[string]string) []map[string]interface{} {
	mEntity := this.ctx.GetEntityInstance(tableName)
	mappingList := this.getEntityMappingFields(mEntity)
	aesList := this.getEntityAESFields(mEntity)

	dataList := this.formatToData(tableName, rows)

	mappingDatasTmp := make(map[string][]map[string]interface{})

	if len(mappingList) > 0 {
		for _, dataItem := range dataList {
			for mappingTable, mtype := range mappingList {
				mappingDatas, has := mappingDatasTmp[mappingTable]
				if !has {
					mappingDatas = this.queryToDatas2(mappingTable, rows)
					mappingDatasTmp[mappingTable] = mappingDatas
				}

				joinObj := this.joinEntities[mappingTable]

				mkeyValue := reflect.ValueOf(dataItem[joinObj.Mkey])
				mkeyValueType := reflect.TypeOf(dataItem[joinObj.Mkey])
				if mkeyValueType == nil {
					continue
				}
				if mkeyValueType.Kind() == reflect.Ptr {
					mkeyValue = mkeyValue.Elem()
				}
				objs := this.joinDataFilter(mappingDatas, mkeyValue, joinObj.Fkey)
				if mtype == "one" {
					if len(mappingDatas) > 0 && len(objs) > 0 {
						dataItem[mappingTable] = objs[0]
					}
				} else if mtype == "many" {
					dataItem[mappingTable] = objs
				}
			}
		}
	}

	if len(aesList) > 0 {
		for _, dataItem := range dataList {
			for _, item := range aesList {
				if dataItem[item] == nil {
					continue
				}

				vType := reflect.TypeOf(dataItem[item])
				if vType.Kind() == reflect.Ptr {
					// 如果是指针类型
					v := reflect.ValueOf(dataItem[item])
					vv := v.Elem().Interface()
					if vv == nil {
						continue
					}
					dataItem[item] = this.interpreter.AesDecrypt(vv.(string), this.interpreter.AESKey)

				} else {
					dataItem[item] = this.interpreter.AesDecrypt(dataItem[item].(string), this.interpreter.AESKey)
				}
			}
		}
	}

	return dataList
}

func (this *EntityObjectKing) formatToData(tableName string, rows map[int]map[string]string) []map[string]interface{} {
	dataList := make([]map[string]interface{}, 0)
	fieldTypeInfos := this.getEntityFieldInfo(tableName)

	for i := 0; i < len(rows); i++ {
		rowData := rows[i]
		dataMap := make(map[string]interface{})

		for fieldKey, value := range rowData {
			tmp := strings.Split(fieldKey, "_")
			tmpTableName := tmp[0]
			if tmpTableName != tableName {
				continue
			}

			fieldName := tmp[1]
			fdType := fieldTypeInfos[fieldName]

			dataMap[fieldName] = this.interpreter.ConverNilValue(fmt.Sprintf("%s", fdType.Type), value)
		}

		exist := false
		for _, dataListItem := range dataList {
			if dataListItem["Id"] == dataMap["Id"] {
				exist = true
				break
			}
		}

		if !exist {
			dataList = append(dataList, dataMap)
		}

	}

	return dataList
}

func (this *EntityObjectKing) getEntityFieldInfo(tableName string) map[string]reflect.StructField {
	result := make(map[string]reflect.StructField)
	entity := this.ctx.GetEntityInstance(tableName)
	eType := reflect.TypeOf(entity)
	ev := reflect.ValueOf(reflect.New(eType).Interface()).Elem()
	for i := 0; i < ev.NumField(); i++ {
		fdType := eType.Field(i)
		result[fdType.Name] = fdType
	}
	return result
}

func (this *EntityObjectKing) joinDataFilter(arr []map[string]interface{}, mKeyValue interface{}, fKey string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	for _, item := range arr {
		if fmt.Sprintf("%s", item[fKey]) == fmt.Sprintf("%s", mKeyValue) {
			result = append(result, item)
		}
	}
	return result
}

func (this *EntityObjectKing) getEntityAESFields(entity interface{}) []string {
	result := make([]string, 0)
	eType := reflect.TypeOf(entity)
	entityPtrValueElem := reflect.ValueOf(reflect.New(eType).Interface()).Elem()
	for i := 0; i < entityPtrValueElem.NumField(); i++ {
		fdType := eType.Field(i)
		defineStr, has := this.interpreter.GetFieldDefineStr(fdType)
		if !has {
			continue
		}
		defineMap := this.interpreter.FormatDefine(defineStr)
		_, hasAES := defineMap[tagDefine.AES]
		if hasAES {
			result = append(result, fdType.Name)
		}
	}
	return result
}

func (this *EntityObjectKing) getEntityMappingFields(entity interface{}) map[string]string {
	result := make(map[string]string)
	eType := reflect.TypeOf(entity)
	entityPtrValueElem := reflect.ValueOf(reflect.New(eType).Interface()).Elem()
	for i := 0; i < entityPtrValueElem.NumField(); i++ {
		fdType := eType.Field(i)
		defineStr, has := this.interpreter.GetFieldDefineStr(fdType)
		if !has {
			continue
		}
		defineMap := this.interpreter.FormatDefine(defineStr)
		fd := entityPtrValueElem.Field(i)
		mapping, has := defineMap[tagDefine.Mapping]
		if has {
			_, hasJoin := this.joinEntities[mapping.(string)]
			if !hasJoin {
				continue
			}
			mappingTableName := mapping.(string)
			if fd.Kind() == reflect.Ptr {
				result[mappingTableName] = "one"
			} else if fd.Kind() == reflect.Slice {
				result[mappingTableName] = "many"
			}
		}
	}

	return result
}
