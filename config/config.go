package config

type Config struct {
	App             AppConfig             `yaml:"app"`
	Logger          LoggerConfig          `yaml:"logger"`
	Telegram        TelegramConfig        `yaml:"telegram"`
	Bybit           BybitConfig           `yaml:"bybit"`
	Postgres        PostgresConfig        `yaml:"postgres"`
	PostgresMetrics PostgresMetricsConfig `yaml:"postgres_metrics"`
}

type AppConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
}

type LoggerConfig struct {
	Level       string   `yaml:"level"`
	Development bool     `yaml:"development"`
	Encoding    string   `yaml:"encoding"` // json или console
	OutputPaths []string `yaml:"output_paths"`
}

type BybitConfig struct {
	APIKey    string `yaml:"api_key"`
	APISecret string `yaml:"api_secret"`
}

type TelegramConfig struct {
	Enabled bool   `yaml:"enabled"`
	Token   string `yaml:"token"`
	ChatID  int64  `yaml:"chat_id"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"db_name"`
	SSLMode  string `yaml:"ssl_mode"`
	PgDriver string `yaml:"pg_driver"`
}

type PostgresMetricsConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"db_name"`
	SSLMode  string `yaml:"ssl_mode"`
	PgDriver string `yaml:"pg_driver"`
}
