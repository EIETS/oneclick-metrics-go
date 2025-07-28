package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

//func Connect() (*pgx.Conn, error) {
//	dbUrl := os.Getenv("DATABASE_URL")
//	if dbUrl == "" {
//		dbUrl = "postgres://user:password@localhost:5432/dbname"
//	}
//	conn, err := pgx.Connect(context.Background(), dbUrl)
//	if err != nil {
//		return nil, err
//	}
//	log.Println("Connected to database")
//	return conn, nil
//}

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Dbname   string `yaml:"dbname"`
	} `yaml:"database"`
}

// 全局连接池
var Pool *pgxpool.Pool

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return fmt.Errorf("创建连接池失败:%w", err)
	}
	//defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("数据库连接失败:%w", err)
	}
	Pool = pool
	return nil
}
