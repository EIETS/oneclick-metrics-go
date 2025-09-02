package db

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"oneclick-metrics-go/global"
	"time"
)

// 全局连接
var DB *sql.DB

// 初始化数据库连接
func InitDb() error {
	var connStr string
	config := global.ServerConfig
	connStr = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		//config.DataBaseInfo.Driver,
		config.DataBaseInfo.User,
		config.DataBaseInfo.Password,
		config.DataBaseInfo.Host,
		config.DataBaseInfo.Port,
		config.DataBaseInfo.Dbname,
	)

	db, err := sql.Open(config.DataBaseInfo.Driver, connStr)
	if err != nil {
		zap.S().Errorf("打开数据库失败: %w", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		zap.S().Errorf("数据库连接失败: %w", err)
		return err
	}

	DB = db

	return nil
}
