package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

func GetRootPath() string {
	dir, _ := os.Getwd()
	return dir
}

func ReadFile(fileName string) string {
	f, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println("read fail", err)
	}
	return string(f)
}

func WriteFile(cont string, fileName string) {
	content := []byte(cont)
	err := ioutil.WriteFile(fileName, content, 0644)
	if err != nil {
		panic(err)
	}
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

func GetGuid() string {
	return strings.ReplaceAll(uuid.NewV4().String(), "-", "")
}

func GetTimeSpan() int64 {
	return time.Now().UnixNano() / 1e6
}

func JsonStringify(v interface{}) string {
	jsonByte, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(jsonByte)
}
