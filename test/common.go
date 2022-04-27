/*
 * @Author: john lee
 * @Date: 2022-03-24 17:13:48
 * @LastEditors: john lee
 * @LastEditTime: 2022-03-30 10:51:15
 * @FilePath: \tiny-entity-go\test\common.go
 * @Description:
 *
 * Copyright (c) 2022 by john lee, All Rights Reserved.
 */
package test

import (
	"encoding/json"
	"fmt"

	"github.com/shishisongsong/tiny-entity-go"
	"github.com/shishisongsong/tiny-entity-go/utils"
)

func GetMysqlCodeGenerator[T tiny.IDataContextInterpreter](ctx T) *tiny.CodeGenerator[T] {
	filePath := fmt.Sprintf("%s/domain/dbConfig.json", utils.GetRootPath())
	fileContent := utils.ReadFile(filePath)
	opt := &tiny.CodeGeneratorOptions{}
	json.Unmarshal([]byte(fileContent), opt)

	codeGenerator := tiny.NewCodeGenerator(ctx, tiny.CodeGeneratorOptions{
		CtxFileName:     "domain/tinyDataContext",
		ModelFilePath:   "domain/models",
		ModulePkgName:   "github.com/shishisongsong/tiny-entity-go/test/domain/models",
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

func GetKingBaseCodeGenerator[T tiny.IDataContextInterpreter](ctx T) *tiny.CodeGenerator[T] {
	filePath := fmt.Sprintf("%s/domain/dbConfig.json", utils.GetRootPath())
	fileContent := utils.ReadFile(filePath)
	opt := &tiny.CodeGeneratorOptions{}
	json.Unmarshal([]byte(fileContent), opt)

	codeGenerator := tiny.NewCodeGenerator(ctx, tiny.CodeGeneratorOptions{
		CtxFileName:     "domain/tinyDataContextKingBase",
		ModelFilePath:   "domain/models",
		ModulePkgName:   "github.com/shishisongsong/tiny-entity-go/test/domain/models",
		PackageName:     "domain",
		Username:        opt.Username,
		Password:        opt.Password,
		DataBaseName:    opt.DataBaseName,
		CharSet:         opt.CharSet,
		ConnectionLimit: opt.ConnectionLimit,
		Host:            opt.Host,
		Driver:          "kingbase",
	})
	return codeGenerator
}
