package test

import (
	"encoding/json"
	"fmt"

	"github.com/joinlee/tiny-entity-go"
	"github.com/joinlee/tiny-entity-go/utils"
)

func GetCodeGenerator[T tiny.IDataContextInterpreter](ctx T) *tiny.CodeGenerator[T] {
	filePath := fmt.Sprintf("%s/domain/dbConfig.json", utils.GetRootPath())
	fileContent := utils.ReadFile(filePath)
	opt := &tiny.CodeGeneratorOptions{}
	json.Unmarshal([]byte(fileContent), opt)

	codeGenerator := tiny.NewCodeGenerator(ctx, tiny.CodeGeneratorOptions{
		CtxFileName:     "domain/tinyDataContext",
		ModelFilePath:   "domain/models",
		ModulePkgName:   "github.com/joinlee/tiny-entity-go/test/domain/models",
		PackageName:     "domain",
		Username:        opt.Username,
		Password:        opt.Password,
		DataBaseName:    opt.DataBaseName,
		CharSet:         opt.CharSet,
		ConnectionLimit: opt.ConnectionLimit,
		Host:            opt.Host,
		Driver:          "mysql", //kingbase
	})
	return codeGenerator
}
