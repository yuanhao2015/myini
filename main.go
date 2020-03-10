package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

type MysqlConfig struct {
	Adress   string `ini:"adress"`
	Port     int    `ini:"port"`
	Username string `ini:"username"`
	Password string `ini:"password"`
}

type RedisConfig struct {
	Host     string `ini:"host"`
	Port     int    `ini:"port"`
	Passwrod string `ini:"password"`
	Database int    `ini:"database"`
}

type Config struct {
	MysqlConfig `ini:"mysql"`
	RedisConfig `ini:"redis"`
}

func loadIni(f string, data interface{}) error {
	file, err := ioutil.ReadFile(f)
	if err != nil {
		fmt.Printf("open file faied, error:%v", err)
		return err
	}
	tObj := reflect.TypeOf(data)
	if tObj.Kind() != reflect.Ptr {
		err = errors.New("data param should be a point")
		return err
	}
	if tObj.Elem().Kind() != reflect.Struct {
		err = errors.New("data param should be a struct")
		return err
	}
	linSlice := strings.Split(string(file), "\n")
	//fmt.Println(linSlice)
	var sectionName string
	for k, line := range linSlice {
		line = strings.TrimSpace(line)
		// 空行 分号 井号均为注释行
		if strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") || len(line) == 0 {
			continue
		}
		// 遇到[开头表示节(section)，如果不是[开头的=分隔的键值对
		if strings.HasPrefix(line, "[") {
			if line[0] != '[' || line[len(line)-1] != ']' {
				err = fmt.Errorf("line %d,syntax error", k+1)
				return err
			}

			if line[0] == '[' && line[len(line)-1] == ']' {
				l := strings.TrimSpace(line[1 : len(line)-1])
				if len(l) == 0 {
					err = fmt.Errorf("line %d,syntax error", k+1)
					return err
				}
				for i := 0; i < tObj.Elem().NumField(); i++ {
					t := tObj.Elem().Field(i)
					if t.Tag.Get("ini") == l {
						sectionName = t.Name
					}
				}
			}

		} else {
			if strings.Index(line, "=") == -1 || strings.HasPrefix(line, "=") {
				err = fmt.Errorf("line %d,syntax error", k+1)
				return err
			}
			vObj := reflect.ValueOf(data)
			sValue := vObj.Elem().FieldByName(sectionName)
			sType := sValue.Type()
			if sType.Kind() != reflect.Struct {
				err = fmt.Errorf("data中的%s 应该是个结构体", sectionName)
				return err
			}
			index := strings.Index(line, "=")
			key := line[:index]
			value := line[index+1:]
			var filedName string
			//便利循环结构体中的每个字段，判断tag是不是等于key
			for i := 0; i < sValue.NumField(); i++ {
				t := sType.Field(i)
				if t.Tag.Get("ini") == key {
					filedName = t.Name
					break
				}
			}
			//根据filename去取出字段
			fileObj := sValue.FieldByName(filedName)
			switch fileObj.Type().Kind() {
			case reflect.String:
				fileObj.SetString(value)
			case reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8, reflect.Int:
				vv, _ := strconv.ParseInt(value, 10, 64)
				fileObj.SetInt(vv)
			case reflect.Float32, reflect.Float64:
				vv, _ := strconv.ParseFloat(value, 64)
				fileObj.SetFloat(vv)
			case reflect.Bool:
				vv, _ := strconv.ParseBool(value)
				fileObj.SetBool(vv)
			}
		}

	}
	return nil
}

func main() {
	var cfg Config
	err := loadIni("./conf.ini", &cfg)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%#v\n", cfg)
}
