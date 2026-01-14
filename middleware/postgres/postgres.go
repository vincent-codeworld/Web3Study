package postgres

import (
	"Web3Study/config"
	"Web3Study/middleware"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Db *gorm.DB

func init() {
	getDsn := func() string {
		pgConfig := config.Config.PostgRestConfig
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
			pgConfig.Host, pgConfig.User, pgConfig.Password, pgConfig.DbName, pgConfig.Port, pgConfig.SslMode, pgConfig.TimeZone)
	}

	db, err := gorm.Open(postgres.Open(getDsn()), &gorm.Config{
		// 配置日志级别，Info 会打印所有 SQL 语句
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	//todo sqlDB 增加数据库池化配置
	sqlDB.SetConnMaxLifetime(100 * time.Second)
	Db = db
	middleware.Hook.Register(sqlDB.Close)
}
