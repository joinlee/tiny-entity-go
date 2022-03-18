package tiny

type IInterpreter interface {
	GetEntityFieldsDefineInfo(entity interface{}) map[string]map[string]interface{}
	GetColumnSqls(defineMap map[string]interface{}, fieldName string, action string, delIndexSql bool, tableName string) (columnSql string, indexSql string)
	AlterTableDropColumn(tableName string, columnName string) string
	AlterTableAddColumn(tableName string, columnName string) string
	AlterTableAlterColumn(tableName string, oldColumnName string, newColumnName string, changeSql string) string
}
