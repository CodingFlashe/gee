package dao

import (
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
	"time"
)

var _db *gorm.DB

func Database(connRead, connWrite string) {
	var ormLogger logger.Interface
	if gin.Mode() == "debug" {
		//打印日志信息
		ormLogger = logger.Default.LogMode(logger.Info)
	} else {
		//线下环境就默认配置
		ormLogger = logger.Default
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       connRead, //数据源名称字段
		DefaultStringSize:         256,      //string类型字段默认长度
		DisableDatetimePrecision:  true,     //禁止datetime精度，mysql 5.6之前的数据库不支持
		DontSupportRenameIndex:    true,     //重命名索引，就要把索引先删掉再重建， mysql 5.7 之前不支持
		DontSupportRenameColumn:   true,     //用 `change` 重命名列，mysql 8之前的数据库不支持
		SkipInitializeWithVersion: false,    //根据当前MySQL版本自动配置
	}), &gorm.Config{
		Logger: ormLogger, //日志文件
		//系统默认数据库表名加s
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, //love表将是love,不再是loves
		},
	})
	if err != nil {
		return
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(20)                  //设置连接池最大连接数,空闲
	sqlDB.SetMaxOpenConns(100)                 //设置与数据库建立连接的最大数目
	sqlDB.SetConnMaxLifetime(time.Second * 30) //连接池里面的连接最大存活时长
	_db = db

	// 主从配置
	_ = db.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{mysql.Open(connWrite)},                      //写操作
		Replicas: []gorm.Dialector{mysql.Open(connRead), mysql.Open(connRead)}, //读操作
		Policy:   dbresolver.RandomPolicy{},
	}))
	migration()
}
func NewDBClient(ctx context.Context) *gorm.DB {
	db := _db
	return db.WithContext(ctx)
}
