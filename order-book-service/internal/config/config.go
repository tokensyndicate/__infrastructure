package config

import (
	"time"

	"github.com/spf13/viper"
)

type ExchangeSettings struct {
    Name      string `mapstructure:"name"`
    Enabled   bool   `mapstructure:"enabled"`
    APIKey    string `mapstructure:"api_key"`    // fallback
    APISecret string `mapstructure:"api_secret"` // fallback
}

type APISettings struct {
    Port            int           `mapstructure:"port"`
    ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type AccessManagerSettings struct {
    Enabled  bool          `mapstructure:"enabled"`
    URL      string        `mapstructure:"url"`
    Timeout  time.Duration `mapstructure:"timeout"`
}

type RedisSettings struct {
    Address  string `mapstructure:"address"`
    Password string `mapstructure:"password"`
    DB       int    `mapstructure:"db"`
}

type Config struct {
    RepositoryType string                      `mapstructure:"repository_type"`
    Database       DatabaseSettings            `mapstructure:"database"`
    InfluxDB       InfluxDBSettings            `mapstructure:"influxdb"`
    API            APISettings                 `mapstructure:"api"`
    Exchanges      map[string]ExchangeSettings `mapstructure:"exchanges"`
    AccessManager  AccessManagerSettings       `mapstructure:"access_manager"`
    Redis          RedisSettings               `mapstructure:"redis"`
    InstanceID     string                      `mapstructure:"instance_id"`
}

type DatabaseSettings struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type InfluxDBSettings struct {
	URL          string `mapstructure:"url"`
	Token        string `mapstructure:"token"`
	Organization string `mapstructure:"organization"`
	Bucket       string `mapstructure:"bucket"`
}

func Load() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("./config")

    setDefaults()
    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }

    if instanceID := viper.GetString("INSTANCE_ID"); instanceID != "" {
        config.InstanceID = instanceID
    }

    return &config, nil
}

func setDefaults() {
    viper.SetDefault("access_manager.enabled", false)
    viper.SetDefault("access_manager.url", "http://access-manager:8080")
    viper.SetDefault("access_manager.timeout", time.Second*10)
    viper.SetDefault("redis.address", "redis:6379")
    viper.SetDefault("redis.db", 0)
    viper.SetDefault("api.shutdown_timeout", time.Second*15)
}
