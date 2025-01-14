package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	MySQL     MySQLConfig     `mapstructure:"mysql"`
	Detection DetectionConfig `mapstructure:"detection"`
}

type ServerConfig struct {
	Address   string `mapstructure:"address"`
	UploadDir string `mapstructure:"upload_dir"`
	FrameDir  string `mapstructure:"frame_dir"`
}

type DetectionConfig struct {
	SupportedVideoFormats []string `mapstructure:"supported_video_formats"`
	MaxFileSize           int64    `mapstructure:"max_file_size"`
	DefaultThreshold      float64  `mapstructure:"default_threshold"`
	MaxFrames             int      `mapstructure:"max_frames"`
}

type MySQLConfig struct {
	DSN             string `mapstructure:"dsn"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../")
	viper.AddConfigPath("../../")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath("../../config")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
