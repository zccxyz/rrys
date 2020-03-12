package model

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
)

var db *gorm.DB

func init() {
	d, err := gorm.Open("mysql", "root:root@/lianxin_test?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal(err, "数据库连接失败")
	}
	db = d
}

func GetDb() *gorm.DB {
	return db
}
