package db

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Dbname   string `yaml:"dbname"`
	} `yaml:"database"`
}

// 全局连接
var DB *sql.DB

// 加载配置类
func LoadConfig(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(data, &cfg)
	return cfg, err
}

// 初始化数据库连接
func InitDb() error {
	var connStr string
	if envUrl := os.Getenv("DATABASE_URL"); envUrl != "" {
		connStr = envUrl
	} else {
		config, err := LoadConfig("config/config.yaml")
		if err != nil {
			return fmt.Errorf("读取配置失败: %w", err)
		}
		connStr = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s",
			config.Database.User,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.Dbname,
		)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	DB = db

	return nil
}
