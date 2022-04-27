package tiny

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shishisongsong/tiny-entity-go/tagDefine"
	"github.com/shishisongsong/tiny-entity-go/utils"
)

type CodeGenerator[T IDataContextInterpreter] struct {
	options CodeGeneratorOptions
	ctx     T
}

func NewCodeGenerator[T IDataContextInterpreter](ctx T, opt CodeGeneratorOptions) *CodeGenerator[T] {
	obj := &CodeGenerator[T]{}
	if opt.Host == "" {
		opt.Host = "localhost"
	}
	if opt.CharSet == "" {
		opt.CharSet = "utf8"
	}
	if opt.Port == "" {
		opt.Port = "3306"
	}
	if opt.ConnectionLimit == 0 {
		opt.ConnectionLimit = 50
	}
	if opt.Driver == "" {
		opt.Driver = "mysql"
	}

	if opt.Driver == "mysql" {
		opt.Port = "3306"
	} else if opt.Driver == "kingbase" {
		opt.Port = "54321"
	}

	obj.options = opt
	obj.ctx = ctx

	return obj
}

func (this *CodeGenerator[T]) getCtxFileName() string {
	fnames := strings.Split(this.options.CtxFileName, "/")
	fName := fnames[len(fnames)-1]

	ctxStructName := utils.Capitalize(fName)

	return ctxStructName
}

func (this *CodeGenerator[T]) GenerateCtxFile() {
	rootPath := utils.GetRootPath()
	ctxStructName := this.getCtxFileName()
	modelNames := this.LoadEntityModes()

	content := fmt.Sprintf("package %s \n", this.options.PackageName)
	content += "import ( \n"
	content += fmt.Sprintf(" \"%s\" \n", this.options.ModulePkgName)
	content += "\"github.com/shishisongsong/tiny-entity-go\" \n"

	if this.options.Driver == "mysql" {
		content += "tinyMysql \"github.com/shishisongsong/tiny-entity-go/mysql\" \n"
	} else if this.options.Driver == "kingbase" {
		content += "tinyKing \"github.com/shishisongsong/tiny-entity-go/kingbase\" \n"
	}

	content += ") \n"
	content += fmt.Sprintf("type %s struct { \n", ctxStructName)

	if this.options.Driver == "mysql" {
		content += "*tinyMysql.MysqlDataContext \n"
	} else if this.options.Driver == "kingbase" {
		content += "*tinyKing.KingDataContext \n"
	}

	for _, modelName := range modelNames {
		content += fmt.Sprintf("%s *models.%s \n", modelName, modelName)
	}

	content += "} \n"
	content += fmt.Sprintf("func New%s() *%s { \n", ctxStructName, ctxStructName)
	content += fmt.Sprintf("ctx := &%s{} \n", ctxStructName)

	if this.options.Driver == "mysql" {
		content += "ctx.MysqlDataContext = tinyMysql.NewMysqlDataContext(tiny.DataContextOptions{ \n"
	} else if this.options.Driver == "kingbase" {
		content += "ctx.KingDataContext = tinyKing.NewKingDataContext(tiny.DataContextOptions{ \n"
	}

	content += fmt.Sprintf("Host:            \"%s\", \n", this.options.Host)
	content += fmt.Sprintf("Port:            \"%s\", \n", this.options.Port)
	content += fmt.Sprintf("Username:            \"%s\", \n", this.options.Username)
	content += fmt.Sprintf("Password:            \"%s\", \n", this.options.Password)
	content += fmt.Sprintf("DataBaseName:            \"%s\", \n", this.options.DataBaseName)
	content += fmt.Sprintf("CharSet:            \"%s\", \n", this.options.CharSet)
	content += fmt.Sprintf("ConnectionLimit:            %d, \n", this.options.ConnectionLimit)
	content += "}) \n\n"

	for _, modelName := range modelNames {
		content += fmt.Sprintf("ctx.%s = &models.%s{ \n", modelName, modelName)

		if this.options.Driver == "mysql" {
			content += fmt.Sprintf("IEntityObject: tinyMysql.NewEntityObjectMysql[models.%s](ctx.MysqlDataContext, \"%s\"),}\n", modelName, modelName)
		} else if this.options.Driver == "kingbase" {
			content += fmt.Sprintf("IEntityObject: tinyKing.NewEntityObjectKing[models.%s](ctx.KingDataContext, \"%s\"),}\n", modelName, modelName)
		}
		content += fmt.Sprintf("ctx.RegistModel(ctx.%s)\n", modelName)
	}

	content += "return ctx } \n"
	content += fmt.Sprintf("func (this *%s) CreateDatabase() { \n", ctxStructName)

	if this.options.Driver == "mysql" {
		content += "this.MysqlDataContext.CreateDatabase() \n"
	} else if this.options.Driver == "kingbase" {
		content += "this.KingDataContext.CreateDatabase() \n"
	}

	for _, modelName := range modelNames {
		content += fmt.Sprintf("this.CreateTable(this.%s) \n", modelName)
	}

	content += "} \n"

	content += fmt.Sprintf("func (this *%s) GetEntityList() map[string]tiny.Entity { \n", ctxStructName)
	content += "list := make(map[string]tiny.Entity) \n"
	for _, modelName := range modelNames {
		content += fmt.Sprintf("list[\"%s\"] = this.%s \n", modelName, modelName)
	}
	content += "return list } \n"

	utils.WriteFile(content, rootPath+"/"+this.options.CtxFileName+".go")
}

func (this *CodeGenerator[T]) LoadEntityModes() []string {
	rootPath := utils.GetRootPath()
	modelNames := make([]string, 0)

	d, _ := os.Open(rootPath + "/" + this.options.ModelFilePath + "/")
	fi, _ := d.Readdir(-1)
	for _, fileItem := range fi {
		if fileItem.Mode().IsRegular() {
			tmp := strings.Split(utils.Capitalize(fileItem.Name()), ".")
			modelNames = append(modelNames, tmp[0])
		}
	}

	return modelNames
}

func (this *CodeGenerator[T]) AutoMigration() {
	var logReport MigrationLog
	ctxFileName := this.getCtxFileName()
	fileStr := utils.ReadFile(fmt.Sprintf("%s/%s_migrationLogs.json", utils.GetRootPath(), ctxFileName))
	if fileStr != "" {
		json.Unmarshal([]byte(fileStr), &logReport)
		// 已经有历史的迁移记录
		r := this.ComparisonTable(logReport)
		if len(r) > 0 {
			logReport = MigrationLog{
				Version: utils.GetTimeSpan(),
				Logs:    r,
			}
		}
	} else {
		// 第一次初始化的时候，生成迁移记录
		logs := make([]MigrationLogInfo, 0)
		for _, entity := range this.ctx.GetEntityList() {
			log := MigrationLogInfo{
				Content: MigrationLogContent{
					TableName:    entity.TableName(),
					Version:      utils.GetTimeSpan(),
					ColumnDefine: this.ctx.GetEntityFieldsDefineInfo(entity),
				},
				Action: "init",
			}

			logs = append(logs, log)
		}

		logReport = MigrationLog{
			Version: utils.GetTimeSpan(),
			Logs:    logs,
		}
	}

	sqlStrs := this.TransLogToSqls(logReport)
	sqlReports := make([]MigrationSqlItem, 0)
	sqlItem := MigrationSqlItem{
		Version: time.Now().UnixNano() / 1e6,
		SqlStrs: sqlStrs,
		Done:    true,
	}

	sqlReports = append(sqlReports, sqlItem)

	if len(sqlStrs) > 0 {
		Transaction(this.ctx, func(ctx T) {
			ctx.Query(strings.Join(sqlStrs, ""))
		})

		utils.WriteFile(utils.JsonStringify(logReport), fmt.Sprintf("%s_migrationLogs.json", ctxFileName))
		utils.WriteFile(utils.JsonStringify(sqlReports), fmt.Sprintf("%s_migrationSqls.json", ctxFileName))
	}

	fmt.Println("AutoMigration Finish!!!")
}

func (this *CodeGenerator[T]) TransLogToSqls(historyLog MigrationLog) []string {
	entityMap := this.ctx.GetEntityList()
	sqlStr := make([]string, 0)
	for _, logItem := range historyLog.Logs {
		entity := entityMap[logItem.Content.TableName]
		if logItem.Action == "init" || logItem.Action == "add" {
			sqlStr = append(sqlStr, this.ctx.CreateTableSQL(entity))
		}

		if logItem.Action == "drop" {
			sqlStr = append(sqlStr, this.ctx.DropTableSQL(logItem.Content.TableName))
		}

		if logItem.Action == "alter" {
			for _, diffItem := range logItem.DiffContent.Column {
				if diffItem.OldItem != nil && diffItem.NewItem == nil {
					// 表示删除字段
					// sqlStr = append(sqlStr, fmt.Sprintf("ALTER TABLE `%s` DROP `%s`; ", logItem.Content.TableName, diffItem.OldItem[tagDefine.Column]))
					sqlStr = append(sqlStr, this.ctx.AlterTableDropColumn(logItem.Content.TableName, diffItem.OldItem[tagDefine.Column].(string)))
				}

				if diffItem.OldItem == nil && diffItem.NewItem != nil {
					// 表示新增字段
					colStr, indexStr := this.ctx.GetColumnSqls(diffItem.NewItem, diffItem.NewItem[tagDefine.Column].(string), "add", false, logItem.Content.TableName)
					// sqlStr = append(sqlStr, fmt.Sprintf("ALTER TABLE `%s` ADD %s; ", logItem.Content.TableName, colStr))
					sqlStr = append(sqlStr, this.ctx.AlterTableAddColumn(logItem.Content.TableName, colStr))
					if indexStr != "" {
						sqlStr = append(sqlStr, indexStr)
					}
				}

				if diffItem.OldItem != nil && diffItem.NewItem != nil {
					// 表示修改字段
					indexDefine := diffItem.OldItem[tagDefine.INDEX]
					delIndex := indexDefine != nil && indexDefine.(bool)

					colStr, indexStr := this.ctx.GetColumnSqls(diffItem.NewItem, diffItem.NewItem[tagDefine.Column].(string), "alter", delIndex, logItem.Content.TableName)

					// sqlStr = append(sqlStr, fmt.Sprintf("ALTER TABLE `%s` CHANGE `%s` `%s` %s; ", logItem.Content.TableName, diffItem.OldItem[tagDefine.Column], diffItem.NewItem[tagDefine.Column], colStr))

					sqlStr = append(sqlStr, this.ctx.AlterTableAlterColumn(logItem.Content.TableName, diffItem.OldItem[tagDefine.Column].(string), diffItem.NewItem[tagDefine.Column].(string), colStr))

					if indexStr != "" {
						sqlStr = append(sqlStr, indexStr)
					}
				}
			}
		}
	}

	tmp := make(map[string]int)

	for _, item := range sqlStr {
		tmp[item] = 1
	}

	sqlStr = make([]string, 0)
	for sql := range tmp {
		sqlStr = append(sqlStr, sql)
	}

	return sqlStr
}

func (this *CodeGenerator[T]) ComparisonTable(historyLog MigrationLog) []MigrationLogInfo {
	entities := this.ctx.GetEntityList()
	diff := make([]MigrationLogInfo, 0)
	// 对比表格的变化
	for _, item := range historyLog.Logs {
		hasTable := false
		for _, entity := range entities {
			if entity.TableName() == item.Content.TableName {
				hasTable = true
				break
			}
		}

		// 删除表
		if !hasTable {
			diff = append(diff, MigrationLogInfo{
				Action:  "drop",
				Content: item.Content,
			})
		}
	}

	for _, entity := range entities {
		var lastHisItem *MigrationLogInfo = nil
		for _, item := range historyLog.Logs {
			if item.Content.TableName == entity.TableName() {
				tmp := item
				lastHisItem = &tmp
			}
		}

		cMeta := this.getAddMigrationLogInfo(entity.TableName(), entity)
		if lastHisItem != nil {
			// 如果上次是是删除了表，这次又加入了表，则添加表
			if lastHisItem.Action == "drop" {
				diff = append(diff, cMeta)
			} else {
				//对比升级字段的不同
				columnDiffList := this.ComparisonColumn(lastHisItem.Content, cMeta.Content)
				if len(columnDiffList) > 0 {
					diff = append(diff, MigrationLogInfo{
						Action:  "alter",
						Content: cMeta.Content,
						DiffContent: struct {
							TableName string
							Column    []MigrationLogDiff
						}{
							TableName: lastHisItem.Content.TableName,
							Column:    columnDiffList,
						},
					})
				} else {
					diff = append(diff, MigrationLogInfo{
						Action:  "noChange",
						Content: cMeta.Content,
					})
				}
			}
		} else {
			diff = append(diff, cMeta)
		}
	}

	return diff
}

func (this *CodeGenerator[T]) ComparisonColumn(oldC MigrationLogContent, newC MigrationLogContent) []MigrationLogDiff {
	diff := make([]MigrationLogDiff, 0)

	for columnName, newItem := range newC.ColumnDefine {
		newItem[tagDefine.Column] = columnName
		for defineKey := range newItem {
			if defineKey == "mapping" {
				// mapping 属性不参与比对
				continue
			}

			isDifferent := false
			oldItem, hasOldItem := oldC.ColumnDefine[columnName]
			if hasOldItem {
				oldItem[tagDefine.Column] = columnName
				tempColumn := make(map[string]interface{})
				tempColumn[tagDefine.Column] = columnName

				if oldItem[tagDefine.Type] != newItem[tagDefine.Type] {
					isDifferent = true
					tempColumn[tagDefine.Type] = newItem[tagDefine.Type]
				}

				if oldItem[tagDefine.NOT_NULL] != newItem[tagDefine.NOT_NULL] {
					isDifferent = true
					tempColumn[tagDefine.NOT_NULL] = newItem[tagDefine.NOT_NULL]
				}

				if oldItem[tagDefine.DEFAULT] != newItem[tagDefine.DEFAULT] {
					isDifferent = true
					tempColumn[tagDefine.DEFAULT] = newItem[tagDefine.DEFAULT]
				}

				if oldItem[tagDefine.INDEX] != newItem[tagDefine.INDEX] {
					isDifferent = true
					tempColumn[tagDefine.INDEX] = newItem[tagDefine.INDEX]
				}

				if isDifferent {
					diff = append(diff, MigrationLogDiff{NewItem: tempColumn, OldItem: oldItem})
					break
				}

			} else {
				diff = append(diff, MigrationLogDiff{NewItem: newItem, OldItem: nil})
				break
			}
		}
	}

	for columnName, oldItem := range oldC.ColumnDefine {
		for defineKey := range oldItem {
			if defineKey == "mapping" {
				// mapping 属性不参与比对
				continue
			}
			_, hasNewItem := newC.ColumnDefine[columnName]
			if !hasNewItem {
				oldItem[tagDefine.Column] = columnName
				diff = append(diff, MigrationLogDiff{NewItem: nil, OldItem: oldItem})
			}
		}
	}

	return diff
}

func (this *CodeGenerator[T]) getAddMigrationLogInfo(tableName string, entity Entity) MigrationLogInfo {
	return MigrationLogInfo{
		Action: "add",
		Content: MigrationLogContent{
			TableName:    tableName,
			Version:      time.Now().UnixNano() / 1e6,
			ColumnDefine: this.ctx.GetEntityFieldsDefineInfo(entity),
		},
	}
}

type CodeGeneratorOptions struct {
	CtxFileName     string
	ModelFilePath   string
	ModulePkgName   string
	PackageName     string
	Host            string `json:"host"`
	Port            string `json:"port"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	DataBaseName    string `json:"dataBaseName"`
	CharSet         string `json:"charSet"`
	ConnectionLimit int    `json:"connectionLimit"`
	Driver          string `json:"driver"`
}

type MigrationLogInfo struct {
	Action      string
	Content     MigrationLogContent
	DiffContent struct {
		TableName string
		Column    []MigrationLogDiff
	}
}

type MigrationLogContent struct {
	TableName    string
	Version      int64
	ColumnDefine map[string]map[string]interface{}
}

type MigrationLog struct {
	Version int64
	Logs    []MigrationLogInfo
}

type MigrationSqlItem struct {
	Version int64
	SqlStrs []string
	Done    bool
}

type MigrationLogDiff struct {
	NewItem map[string]interface{}
	OldItem map[string]interface{}
}

func Log(v interface{}) {
	value := os.Getenv("TINY_LOG")
	if value == "ON" {
		fmt.Println("[Tiny Debug]", v)
	}
}
