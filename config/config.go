package config

type DataBase struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Dbname   string `mapstructure:"dbname"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type ServerConfig struct {
	Host         string    `mapstructure:"host"`
	Port         string    `mapstructure:"port"`
	DataBaseInfo DataBase  `mapstructure:"database"`
	Timezone     string    `mapstructure:"timezone"`
	Interval     int       `mapstructure:"interval"`
	LogInfo      LogConfig `mapstructure:"log"`
}
