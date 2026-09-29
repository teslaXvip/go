package xorm

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	xormpkg "xorm.io/xorm"
)

type Login struct {
	Username string
	Password string
}

func XormQuickStart() {
	host := "localhost"
	port := 3306
	dbname := "test"
	user := "tester"
	pass := "123456"

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, port, dbname)
	engine, err := xormpkg.NewEngine("mysql", connStr)
	if err != nil {
		panic(err)
	}

	err = engine.Ping()
	if err != nil {
		panic(err)
	}

	instance1 := Login{Username: "xc", Password: "666"}
	engine.Insert(instance1)

	var instance2 Login
	engine.Get(&instance2)
	fmt.Printf("%#v\n", instance2)
}
