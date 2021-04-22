package tiny

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
	"github.com/joinlee/tiny-entity-go/tagDefine"
)

type CodeGenerator struct {
	options CodeGeneratorOptions
	ctx     IDataContext
}

func NewCodeGenerator(opt CodeGeneratorOptions) *CodeGenerator {
	obj := &CodeGenerator{}
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
	obj.options = opt

	return obj
}

func (this *CodeGenerator) GenerateCtxFile() {
	rootPath := this.getRootPath()
	ctxStructName := Capitalize(this.options.CtxFileName)
	modelNames := this.LoadEntityModes()

	content := fmt.Sprintf("package %s \n", this.options.PackageName)
	content += "import ( \n"
	content += fmt.Sprintf("\"tinyGo/%s\" \n", this.options.ModelFilePath)
	content += "\"github.com/joinlee/tiny-entity-go\" \n"
	content += "tinyMysql \"github.com/joinlee/tiny-entity-go/mysql\" \n"
	content += ") \n"
	content += fmt.Sprintf("type %s struct { \n", ctxStructName)
	content += "*tinyMysql.MysqlDataContext \n"

	for _, modelName := range modelNames {
		content += fmt.Sprintf("%s *models.%s \n", modelName, modelName)
	}

	content += "} \n"
	content += fmt.Sprintf("func New%s() *%s { \n", ctxStructName, ctxStructName)
	content += fmt.Sprintf("ctx := &%s{} \n", ctxStructName)
	content += "ctx.MysqlDataContext = tinyMysql.NewMysqlDataContext(tinyMysql.MysqlDataOption{ \n"
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
		content += fmt.Sprintf("EntityObjectMysql: tinyMysql.NewEntityObjectMysql(ctx.MysqlDataContext, \"%s\"),}\n", modelName)
		content += fmt.Sprintf("ctx.RegistModel(ctx.%s)\n", modelName)
	}

	content += "return ctx } \n"
	content += fmt.Sprintf("func (this *%s) CreateDatabase() { \n", ctxStructName)
	content += "this.MysqlDataContext.CreateDatabase() \n"
	for _, modelName := range modelNames {
		content += fmt.Sprintf("this.CreateTable(this.%s) \n", modelName)
	}

	content += "} \n"

	content += "func (this *TinyDataContext) GetEntityList() map[string]tiny.IEntityObject { \n"
	content += "list := make(map[string]tiny.IEntityObject) \n"
	for _, modelName := range modelNames {
		content += fmt.Sprintf("list[\"%s\"] = this.%s \n", modelName, modelName)
	}
	content += "return list } \n"

	WriteFile(content, rootPath+"/domain/"+this.options.CtxFileName+".go")
}

func (this *CodeGenerator) LoadEntityModes() []string {
	rootPath := this.getRootPath()
	modelNames := make([]string, 0)

	d, _ := os.Open(rootPath + "/" + this.options.ModelFilePath + "/")
	fi, _ := d.Readdir(-1)
	for _, fileItem := range fi {
		if fileItem.Mode().IsRegular() {
			tmp := strings.Split(Capitalize(fileItem.Name()), ".")
			modelNames = append(modelNames, tmp[0])
		}
	}

	return modelNames
}

func (this *CodeGenerator) AutoMigration(ctx IDataContext) {
	this.ctx = ctx
	var logReport MigrationLog

	fileStr := ReadFile("migrationLogs.json")
	if fileStr != "" {
		json.Unmarshal([]byte(fileStr), &logReport)
		// 已经有历史的迁移记录
		r := this.ComparisonTable(logReport)
		if len(r) > 0 {
			logReport = MigrationLog{
				Version: time.Now().UnixNano() / 1e6,
				Logs:    r,
			}
		}
	} else {
		// 第一次初始化的时候，生成迁移记录
		logs := make([]MigrationLogInfo, 0)
		for _, entity := range ctx.GetEntityList() {
			interpreter := NewInterpreter(entity.TableName())
			log := MigrationLogInfo{
				Content: MigrationLogContent{
					TableName:    entity.TableName(),
					Version:      time.Now().UnixNano() / 1e6,
					ColumnDefine: interpreter.GetEntityFieldsDefineInfo(entity),
				},
				Action: "init",
			}

			logs = append(logs, log)
		}

		logReport = MigrationLog{
			Version: time.Now().UnixNano() / 1e6,
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
		Transaction(this.ctx, func(ctx IDataContext) {
			ctx.Query(strings.Join(sqlStrs, ""), true)
		})

		WriteFile(JsonStringify(logReport), "migrationLogs.json")
		WriteFile(JsonStringify(sqlReports), "migrationSqls.json")
	}

	fmt.Println("AutoMigration Finish!!!")
}

func (this *CodeGenerator) getRootPath() string {
	dir, _ := os.Getwd()
	return dir
}

func (this *CodeGenerator) TransLogToSqls(historyLog MigrationLog) []string {
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
				interpreter := NewInterpreter(logItem.Content.TableName)

				if diffItem.OldItem != nil && diffItem.NewItem == nil {
					// 表示删除字段
					sqlStr = append(sqlStr, fmt.Sprintf("ALTER TABLE `%s` DROP `%s`; ", logItem.Content.TableName, diffItem.OldItem[tagDefine.Column]))
				}

				if diffItem.OldItem == nil && diffItem.NewItem != nil {
					// 表示新增字段
					sqlStr = append(sqlStr, fmt.Sprintf("ALTER TABLE `%s` ADD %s; ", logItem.Content.TableName, interpreter.GetColumnSqls(diffItem.NewItem, diffItem.NewItem[tagDefine.Column].(string), "add", false)))
				}

				if diffItem.OldItem != nil && diffItem.NewItem != nil {
					// 表示修改字段
					indexDefine := diffItem.OldItem[tagDefine.INDEX]
					delIndex := indexDefine != nil && indexDefine.(bool)

					sqlStr = append(sqlStr, fmt.Sprintf("ALTER TABLE `%s` CHANGE `%s` `%s` %s; ", logItem.Content.TableName, diffItem.OldItem[tagDefine.Column], diffItem.NewItem[tagDefine.Column], interpreter.GetColumnSqls(diffItem.NewItem, diffItem.NewItem[tagDefine.Column].(string), "alter", delIndex)))
				}
			}
		}
	}

	return sqlStr
}

func (this *CodeGenerator) ComparisonTable(historyLog MigrationLog) []MigrationLogInfo {
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

func (this *CodeGenerator) ComparisonColumn(oldC MigrationLogContent, newC MigrationLogContent) []MigrationLogDiff {
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
				diff = append(diff, MigrationLogDiff{NewItem: nil, OldItem: oldItem})
			}
		}
	}

	return diff
}

func (this *CodeGenerator) getAddMigrationLogInfo(tableName string, entity IEntityObject) MigrationLogInfo {
	interpreter := NewInterpreter(tableName)
	return MigrationLogInfo{
		Action: "add",
		Content: MigrationLogContent{
			TableName:    tableName,
			Version:      time.Now().UnixNano() / 1e6,
			ColumnDefine: interpreter.GetEntityFieldsDefineInfo(entity),
		},
	}
}

type CodeGeneratorOptions struct {
	CtxFileName     string
	ModelFilePath   string
	PackageName     string
	Host            string `json:"host"`
	Port            string `json:"port"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	DataBaseName    string `json:"dataBaseName"`
	CharSet         string `json:"charSet"`
	ConnectionLimit int    `json:"connectionLimit"`
}

func Capitalize(str string) string {
	var upperStr string
	vv := []rune(str) // 后文有介绍
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if vv[i] >= 97 && vv[i] <= 122 { // 后文有介绍
				vv[i] -= 32 // string的码表相差32位
				upperStr += string(vv[i])
			} else {
				fmt.Println("Not begins with lowercase letter,")
				return str
			}
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}

func WriteFile(cont string, fileName string) {
	content := []byte(cont)
	err := ioutil.WriteFile(fileName, content, 0644)
	if err != nil {
		panic(err)
	}
}

func ReadFile(fileName string) string {
	f, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println("read fail", err)
	}
	return string(f)
}

func JsonStringify(v interface{}) string {
	jsonByte, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(jsonByte)
}

func ArrayFind(list interface{}, key string) interface{} {
	return nil
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
	// log.SetPrefix("[Tiny Debug] ")
	// log.Println(v)
	fmt.Println("[Tiny Debug]", v)
}
