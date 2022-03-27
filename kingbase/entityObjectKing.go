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

type EntityObjectKing[T tiny.Entity] struct {
	ctx *KingDataContext

	tableName    string
	joinEntities map[string]tinyMysql.JoinEntityItem
}

func NewEntityObjectKing[T tiny.Entity](ctx *KingDataContext, tableName string) *EntityObjectKing[T] {
	entity := &EntityObjectKing[T]{}
	entity.ctx = ctx
	entity.tableName = tableName
	entity.InitEntityObj(tableName)

	return entity
}

func (this *EntityObjectKing[T]) InitEntityObj(tableName string) {
	// this.ctx = tiny.NewctxKing(tableName)
	this.joinEntities = make(map[string]tinyMysql.JoinEntityItem)
}

func (this *EntityObjectKing[T]) TableName() string {
	return this.tableName
}

func (this *EntityObjectKing[T]) GetIQueryObject() tiny.IQueryObject[T] {
	return this
}

func (this *EntityObjectKing[T]) And() tiny.IQueryObject[T] {
	this.ctx.AddToWhere("AND", false)
	return this
}

func (this *EntityObjectKing[T]) Or() tiny.IQueryObject[T] {
	this.ctx.AddToWhere("OR", false)
	return this
}

// 添加查询条件
/* queryStr 查询语句， args 条件参数
ex： ctx.User.Where("Id = ?", user.Id).Any() */
func (this *EntityObjectKing[T]) Where(queryStr interface{}, args ...interface{}) tiny.IQueryObject[T] {
	return this.wherePartHandle(this.tableName, queryStr, args)
}

//添加指定表的查询条件
/* entity 需要查询的实体 queryStr 查询语句， args 条件参数
entity 表示查询外键表的条件
ex： ctx.User.WhereWith(ctx.Account, "Id = ?", user.Id).Any() */
func (this *EntityObjectKing[T]) WhereWith(entity tiny.Entity, queryStr interface{}, args ...interface{}) tiny.IQueryObject[T] {
	tableName := reflect.TypeOf(entity).Elem().Name()
	return this.wherePartHandle(tableName, queryStr, args)
}

func (this *EntityObjectKing[T]) Contains(felid string, values interface{}) tiny.IQueryObject[T] {
	return this.inPartHandle(this.tableName, felid, values)
}

func (this *EntityObjectKing[T]) ContainsWith(entity tiny.Entity, felid string, values interface{}) tiny.IQueryObject[T] {
	tableName := reflect.TypeOf(entity).Elem().Name()
	return this.inPartHandle(tableName, felid, values)
}
func (this *EntityObjectKing[T]) inPartHandle(tableName string, felid string, values interface{}) tiny.IQueryObject[T] {
	qs := "\"" + tableName + "\".\"" + felid + "\" IN"
	vs := make([]string, 0)

	if reflect.TypeOf(values).Kind() == reflect.Slice {
		s := reflect.ValueOf(values)
		for i := 0; i < s.Len(); i++ {
			value := s.Index(i)
			vs = append(vs, this.ctx.TransValueToStrByType(value, value.Kind().String()))
		}

		qs = qs + " ( " + strings.Join(vs, ",") + " )"
		this.ctx.AddToWhere(qs, true)
	}
	return this
}

func (this *EntityObjectKing[T]) wherePartHandle(tableName string, queryStr interface{}, args []interface{}) tiny.IQueryObject[T] {
	if queryStr == nil || queryStr == "" {
		return this
	}
	qs := queryStr.(string)
	for _, value := range args {
		qs = strings.Replace(qs, "?", this.ctx.TransValueToStr(value), 1)
	}
	qs = strings.ReplaceAll(qs, "`", "")
	qs = this.ctx.AddFieldTableName(qs, tableName)
	this.ctx.AddToWhere(qs, true)
	return this
}

func (this *EntityObjectKing[T]) OrderBy(field interface{}) tiny.IQueryObject[T] {
	this.ctx.AddToOrdereBy(field.(string), false, this.tableName)
	return this
}

func (this *EntityObjectKing[T]) OrderByDesc(field interface{}) tiny.IQueryObject[T] {
	this.ctx.AddToOrdereBy(field.(string), true, this.tableName)
	return this
}

func (this *EntityObjectKing[T]) IndexOf() tiny.IQueryObject[T] {
	return this
}

func (this *EntityObjectKing[T]) GroupBy(field interface{}) tiny.IResultQueryObject[T] {
	fStr := field.(string)

	if len(this.joinEntities) > 0 {
		for k := range this.joinEntities {
			// 将所有外键表的Id加入到GroupBy 子句
			this.ctx.AddToGroupBy("Id", k)
		}
	}
	// 将主键表加入到GroupBy 子句
	this.ctx.AddToGroupBy("Id", this.tableName)

	if fStr != "Id" {
		this.ctx.AddToGroupBy(fStr, this.tableName)
	}

	return this
}

func (this *EntityObjectKing[T]) Select(fields ...interface{}) tiny.IResultQueryObject[T] {
	list := make([]string, 0)
	for _, item := range fields {
		list = append(list, fmt.Sprintf("%s AS %s_%s", this.ctx.AddFieldTableName(item.(string), this.tableName), this.tableName, item.(string)))
	}
	this.ctx.CleanSelectPart()
	this.ctx.AddToSelect(list)
	return this
}

func (this *EntityObjectKing[T]) Take(count int) tiny.ITakeChildQueryObject[T] {
	this.ctx.AddToLimt("take", count)
	return this
}
func (this *EntityObjectKing[T]) Skip(count int) tiny.IAssembleResultQuery[T] {
	this.ctx.AddToLimt("skip", count)
	return this
}

// 添加外联引用
/*
	fEntity 需要连接的实体， mField 主表的连接字段， fField 外联表的字段
*/
func (this *EntityObjectKing[T]) JoinOn(fEntity tiny.Entity, mField string, fField string) tiny.IQueryObject[T] {
	return this.joinHandle(this.tableName, fEntity, mField, fField)
}

func (this *EntityObjectKing[T]) JoinOnWith(mEntity tiny.Entity, fEntity tiny.Entity, mField string, fField string) tiny.IQueryObject[T] {
	mTableName := mEntity.TableName()
	return this.joinHandle(mTableName, fEntity, mField, fField)
}
func (this *EntityObjectKing[T]) joinHandle(mTableName string, fEntity tiny.Entity, mField string, fField string) tiny.IQueryObject[T] {
	if len(this.joinEntities) == 0 {
		mEntity := this.ctx.GetEntityInstance(mTableName)
		mainTableFields := this.ctx.GetSelectFieldList(mEntity.(tiny.Entity), mTableName)
		this.ctx.AddToSelect(mainTableFields)
	}

	fTableFields := this.ctx.GetSelectFieldList(fEntity, fEntity.TableName())
	this.ctx.AddToSelect(fTableFields)

	this.joinEntities[fEntity.TableName()] = tinyMysql.JoinEntityItem{
		Mkey:   mField,
		Fkey:   fField,
		Entity: fEntity,
	}
	sqlStr := fmt.Sprintf(" LEFT JOIN \"%s\" ON \"%s\".\"%s\" = \"%s\".\"%s\"", fEntity.TableName(), mTableName, mField, fEntity.TableName(), fField)
	this.ctx.AddToJoinOn(sqlStr)
	return this
}

func (this *EntityObjectKing[T]) Max() float64 {
	this.ctx.Clean()
	return 0
}

func (this *EntityObjectKing[T]) Min() float64 {
	this.ctx.Clean()
	return 0
}
func (this *EntityObjectKing[T]) Count() int {
	return this.CountArgs(fmt.Sprintf("\"%s\".\"Id\"", this.tableName))
}

func (this *EntityObjectKing[T]) CountArgs(field string) int {
	this.ctx.CleanSelectPart()
	field = strings.ReplaceAll(field, "`", "\"")
	this.ctx.AddToSelect([]string{fmt.Sprintf("COUNT(%s)", field)})
	sqlStr := this.ctx.GetFinalSql(this.tableName, nil)
	rows := this.ctx.Query(sqlStr)

	result := 0
	for _, rowData := range rows {
		for _, cellData := range rowData {
			result, _ = strconv.Atoi(cellData)
		}
	}
	this.ctx.Clean()
	return result
}

func (this *EntityObjectKing[T]) Any() bool {
	count := this.Count()
	return count > 0
}

func (this *EntityObjectKing[T]) First() *T {
	entity := new(T)

	sqlStr := this.ctx.GetFinalSql(this.tableName, *entity)
	rows := this.ctx.Query(sqlStr)
	dataList := this.queryToDatas2(this.tableName, rows)

	if len(dataList) > 0 {
		jsonStr := tiny.JsonStringify(dataList[0])
		json.Unmarshal([]byte(jsonStr), entity)
	} else {
		entity = nil
	}

	this.InitEntityObj(this.tableName)
	this.ctx.Clean()

	return entity
}

func (this *EntityObjectKing[T]) ToList() []T {
	list := make([]T, 0)
	mEntity := new(T)
	sqlStr := this.ctx.GetFinalSql(this.tableName, *mEntity)
	rows := this.ctx.Query(sqlStr)
	dataList := this.queryToDatas2(this.tableName, rows)

	jsonStr := tiny.JsonStringify(dataList)
	json.Unmarshal([]byte(jsonStr), &list)
	this.InitEntityObj(this.tableName)
	this.ctx.Clean()

	return list
}

func (this *EntityObjectKing[T]) queryToDatas2(tableName string, rows map[int]map[string]string) []map[string]interface{} {
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
					dataItem[item] = this.ctx.AesDecrypt(vv.(string), this.ctx.AESKey)

				} else {
					dataItem[item] = this.ctx.AesDecrypt(dataItem[item].(string), this.ctx.AESKey)
				}
			}
		}
	}

	return dataList
}

func (this *EntityObjectKing[T]) formatToData(tableName string, rows map[int]map[string]string) []map[string]interface{} {
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

			dataMap[fieldName] = this.ctx.ConverNilValue(fmt.Sprintf("%s", fdType.Type), value)
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

func (this *EntityObjectKing[T]) getEntityFieldInfo(tableName string) map[string]reflect.StructField {
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

func (this *EntityObjectKing[T]) joinDataFilter(arr []map[string]interface{}, mKeyValue interface{}, fKey string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	for _, item := range arr {
		if fmt.Sprintf("%s", item[fKey]) == fmt.Sprintf("%s", mKeyValue) {
			result = append(result, item)
		}
	}
	return result
}

func (this *EntityObjectKing[T]) getEntityAESFields(entity interface{}) []string {
	result := make([]string, 0)
	eType := reflect.TypeOf(entity)
	entityPtrValueElem := reflect.ValueOf(reflect.New(eType).Interface()).Elem()
	for i := 0; i < entityPtrValueElem.NumField(); i++ {
		fdType := eType.Field(i)
		defineStr, has := this.ctx.GetFieldDefineStr(fdType)
		if !has {
			continue
		}
		defineMap := this.ctx.FormatDefine(defineStr)
		_, hasAES := defineMap[tagDefine.AES]
		if hasAES {
			result = append(result, fdType.Name)
		}
	}
	return result
}

func (this *EntityObjectKing[T]) getEntityMappingFields(entity interface{}) map[string]string {
	result := make(map[string]string)
	eType := reflect.TypeOf(entity)
	entityPtrValueElem := reflect.ValueOf(reflect.New(eType).Interface()).Elem()
	for i := 0; i < entityPtrValueElem.NumField(); i++ {
		fdType := eType.Field(i)
		defineStr, has := this.ctx.GetFieldDefineStr(fdType)
		if !has {
			continue
		}
		defineMap := this.ctx.FormatDefine(defineStr)
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
