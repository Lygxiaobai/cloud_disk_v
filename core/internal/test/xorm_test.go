package test

import (
	"bytes"
	"cloud_disk/core/internal/models"

	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"testing"
	"xorm.io/xorm"
)

func TestCRUDXorm(t *testing.T) {
	var engine *xorm.Engine
	var err error
	engine, err = xorm.NewEngine("mysql", "root:123456@(127.0.0.1:3306)/cloud_disk?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		t.Fatal(err)
	}
	//find查询需要传入slice或map类型
	data := make([]*models.UserBasic, 0)
	err = engine.Find(&data)
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	dst := new(bytes.Buffer)
	err = json.Indent(dst, b, "", " ")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(dst)
}
