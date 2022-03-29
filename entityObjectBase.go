package tiny

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/joinlee/tiny-entity-go/tagDefine"
	"github.com/joinlee/tiny-entity-go/utils"
)

type JoinEntityItem struct {
	Mkey   string
	Fkey   string
	Entity interface{}
}

type EntityObjectBase[T Entity] struct {
	Ctx          *DataContextBase
	TableName    string
	JoinEntities map[string]JoinEntityItem
}

func (this *EntityObjectBase[T]) SetCtx(ctx interface{}) {
	this.Ctx = ctx.(*DataContextBase)
}

func (this *EntityObjectBase[T]) First() *T {
	entity := new(T)
	sqlStr := this.Ctx.GetFinalSql(this.TableName, *entity)
	rows := this.Ctx.Query(sqlStr)
	dataList := this.QueryToDatas2(this.TableName, rows)

	if len(dataList) > 0 {
		jsonStr := utils.JsonStringify(dataList[0])
		json.Unmarshal([]byte(jsonStr), entity)
	} else {
		entity = nil
	}

	this.InitEntityObj(this.TableName)
	this.Ctx.Clean()

	return entity
}

func (this *EntityObjectBase[T]) ToList() []T {
	list := make([]T, 0)
	mEntity := new(T)
	sqlStr := this.Ctx.GetFinalSql(this.TableName, *mEntity)
	rows := this.Ctx.Query(sqlStr)
	dataList := this.QueryToDatas2(this.TableName, rows)

	jsonStr := utils.JsonStringify(dataList)
	json.Unmarshal([]byte(jsonStr), &list)
	this.InitEntityObj(this.TableName)
	this.Ctx.Clean()

	return list
}

func (this *EntityObjectBase[T]) InitEntityObj(tableName string) {
	this.TableName = tableName
	this.JoinEntities = make(map[string]JoinEntityItem)
}

func (this *EntityObjectBase[T]) JoinHandle(mTableName string, fEntity Entity, mField string, fField string) {
	if len(this.JoinEntities) == 0 {
		mEntity := new(T)
		mainTableFields := this.Ctx.GetSelectFieldList(*mEntity, mTableName)
		this.Ctx.AddToSelect(mainTableFields)
	}

	fTableFields := this.Ctx.GetSelectFieldList(fEntity, fEntity.TableName())
	this.Ctx.AddToSelect(fTableFields)

	this.JoinEntities[fEntity.TableName()] = JoinEntityItem{
		Mkey:   mField,
		Fkey:   fField,
		Entity: fEntity,
	}
	sqlStr := fmt.Sprintf(" LEFT JOIN `%s` ON `%s`.`%s` = `%s`.`%s`", fEntity.TableName(), mTableName, mField, fEntity.TableName(), fField)
	this.Ctx.AddToJoinOn(sqlStr)
}

func (this *EntityObjectBase[T]) WherePartHandle(tableName string, queryStr interface{}, args []interface{}) {
	if queryStr == nil || queryStr == "" {
		return
	}
	qs := queryStr.(string)
	for _, value := range args {
		qs = strings.Replace(qs, "?", this.Ctx.TransValueToStr(value), 1)
	}
	qs = this.Ctx.AddFieldTableName(qs, tableName)
	this.Ctx.AddToWhere(qs, true)
}

func (this *EntityObjectBase[T]) GroupByHandle(field interface{}) {
	fStr := field.(string)

	if len(this.JoinEntities) > 0 {
		for k := range this.JoinEntities {
			// 将所有外键表的Id加入到GroupBy 子句
			this.Ctx.AddToGroupBy("Id", k)
		}
	}
	// 将主键表加入到GroupBy 子句
	this.Ctx.AddToGroupBy("Id", this.TableName)

	if fStr != "Id" {
		this.Ctx.AddToGroupBy(fStr, this.TableName)
	}
}

func (this *EntityObjectBase[T]) InPartHandle(tableName string, felid string, values interface{}) {
	qs := "`" + tableName + "`.`" + felid + "` IN"
	vs := make([]string, 0)

	if reflect.TypeOf(values).Kind() == reflect.Slice {
		s := reflect.ValueOf(values)
		for i := 0; i < s.Len(); i++ {
			value := s.Index(i)
			vs = append(vs, this.Ctx.TransValueToStrByType(value, value.Kind().String()))
		}

		qs = qs + " ( " + strings.Join(vs, ",") + " )"
		this.Ctx.AddToWhere(qs, true)
	}
}

func (this *EntityObjectBase[T]) SelectHandle(fields ...interface{}) {
	list := make([]string, 0)
	for _, item := range fields {
		list = append(list, fmt.Sprintf("%s AS %s_%s", this.Ctx.AddFieldTableName(item.(string), this.TableName), this.TableName, item.(string)))
	}
	this.Ctx.CleanSelectPart()
	this.Ctx.AddToSelect(list)
}

func (this *EntityObjectBase[T]) CountHandle(field string) int {
	this.Ctx.CleanSelectPart()
	this.Ctx.AddToSelect([]string{fmt.Sprintf("COUNT(%s)", field)})
	sqlStr := this.Ctx.GetFinalSql(this.TableName, nil)
	rows := this.Ctx.Query(sqlStr)

	result := 0
	for _, rowData := range rows {
		for _, cellData := range rowData {
			result, _ = strconv.Atoi(cellData)
		}
	}
	this.Ctx.Clean()
	return result
}

func (this *EntityObjectBase[T]) QueryToDatas2(tableName string, rows map[int]map[string]string) []map[string]interface{} {
	mEntity := new(T)
	mappingList := this.getEntityMappingFields(*mEntity)
	aesList := this.getEntityAESFields(*mEntity)

	dataList := this.formatToData(tableName, rows)

	mappingDatasTmp := make(map[string][]map[string]interface{})

	if len(mappingList) > 0 {
		for _, dataItem := range dataList {
			for mappingTable, mtype := range mappingList {
				mappingDatas, has := mappingDatasTmp[mappingTable]
				if !has {
					mappingDatas = this.QueryToDatas2(mappingTable, rows)
					mappingDatasTmp[mappingTable] = mappingDatas
				}

				joinObj := this.JoinEntities[mappingTable]

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
					dataItem[item] = this.Ctx.AesDecrypt(vv.(string), this.Ctx.AESKey)

				} else {
					dataItem[item] = this.Ctx.AesDecrypt(dataItem[item].(string), this.Ctx.AESKey)
				}
			}
		}
	}

	return dataList
}

func (this *EntityObjectBase[T]) joinDataFilter(arr []map[string]interface{}, mKeyValue interface{}, fKey string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	for _, item := range arr {
		if fmt.Sprintf("%s", item[fKey]) == fmt.Sprintf("%s", mKeyValue) {
			result = append(result, item)
		}
	}
	return result
}

func (this *EntityObjectBase[T]) getEntityMappingFields(entity T) map[string]string {
	result := make(map[string]string)
	eType := reflect.TypeOf(entity)
	entityPtrValueElem := reflect.ValueOf(reflect.New(eType).Interface()).Elem()
	for i := 0; i < entityPtrValueElem.NumField(); i++ {
		fdType := eType.Field(i)
		defineStr, has := this.Ctx.GetFieldDefineStr(fdType)
		if !has {
			continue
		}
		defineMap := this.Ctx.FormatDefine(defineStr)
		fd := entityPtrValueElem.Field(i)
		mapping, has := defineMap[tagDefine.Mapping]
		if has {
			_, hasJoin := this.JoinEntities[mapping.(string)]
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

func (this *EntityObjectBase[T]) getEntityAESFields(entity T) []string {
	result := make([]string, 0)
	eType := reflect.TypeOf(entity)
	entityPtrValueElem := reflect.ValueOf(reflect.New(eType).Interface()).Elem()
	for i := 0; i < entityPtrValueElem.NumField(); i++ {
		fdType := eType.Field(i)
		defineStr, has := this.Ctx.GetFieldDefineStr(fdType)
		if !has {
			continue
		}
		defineMap := this.Ctx.FormatDefine(defineStr)
		_, hasAES := defineMap[tagDefine.AES]
		if hasAES {
			result = append(result, fdType.Name)
		}
	}
	return result
}

func (this *EntityObjectBase[T]) getEntityFieldInfo(tableName string) map[string]reflect.StructField {
	result := make(map[string]reflect.StructField)
	entity := new(T)
	et := reflect.TypeOf(entity).Elem()
	ev := reflect.ValueOf(entity).Elem()
	for i := 0; i < ev.NumField(); i++ {
		fdType := et.Field(i)
		result[fdType.Name] = fdType
	}
	return result
}

func (this *EntityObjectBase[T]) formatToData(tableName string, rows map[int]map[string]string) []map[string]interface{} {
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

			dataMap[fieldName] = this.Ctx.ConverNilValue(fmt.Sprintf("%s", fdType.Type), value)
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
