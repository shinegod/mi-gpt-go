package database

import (
	"mi-gpt-go/internal/models"
	"mi-gpt-go/pkg/logger"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 初始化数据库
func Init(dbPath string) (*gorm.DB, error) {
	var err error
	
	// 配置 GORM 日志级别
	config := &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	}

	// 连接数据库
	DB, err = gorm.Open(sqlite.Open(dbPath), config)
	if err != nil {
		return nil, err
	}

	// 自动迁移数据库结构
	err = DB.AutoMigrate(
		&models.Config{},           // 配置表
		&models.User{},
		&models.Room{},
		&models.Message{},
		&models.Memory{},
		&models.ShortTermMemory{},
		&models.LongTermMemory{},
	)
	if err != nil {
		return nil, err
	}

	logger.Info("数据库初始化成功")
	return DB, nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
} 