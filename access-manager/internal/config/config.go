// internal/config/config.go
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
    Server struct {
        Port int    `mapstructure:"port"`
        Host string `mapstructure:"host"`
    } `mapstructure:"server"`

    Vault struct {
        Address string `mapstructure:"address"`
        Token   string `mapstructure:"token"`
    } `mapstructure:"vault"`

    Redis struct {
        Address  string `mapstructure:"address"`
        Password string `mapstructure:"password"`
        DB       int    `mapstructure:"db"`
    } `mapstructure:"redis"`

    Logging struct {
        Level string `mapstructure:"level"`
    } `mapstructure:"logging"`
}

func LoadConfig(configPath string) (*Config, error) {
    config := &Config{}

    viper.SetDefault("server.port", 8080)
    viper.SetDefault("server.host", "0.0.0.0")
    viper.SetDefault("vault.address", "http://localhost:8200")
    viper.SetDefault("redis.address", "localhost:6379")
    viper.SetDefault("redis.db", 0)
    viper.SetDefault("logging.level", "info")

    if configPath != "" {
        viper.SetConfigFile(configPath)
    } else {
        viper.AddConfigPath(".")
        viper.AddConfigPath("./config")
        viper.SetConfigName("config")
    }

    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, fmt.Errorf("error reading config file: %w", err)
        }
    }

    if err := viper.Unmarshal(config); err != nil {
        return nil, fmt.Errorf("error unmarshaling config: %w", err)
    }

    return config, nil
}
