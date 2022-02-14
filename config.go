package main

import (
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	ConfigName = "fabric.toml"
)

type Fabric struct {
	Name        string `toml:"name" json:"name"`
	Addr        string `toml:"addr" json:"addr"`
	OrganizationsPath        string `toml:"organizations_path" json:"organizations_path"`
	Username    string `toml:"username" json:"username"`
	CCID        string `toml:"ccid" json:"ccid"`
	ChannelId   string `mapstructure:"channel_id" toml:"channel_id" json:"channel_id"`

}

func DefaultConfig() *Fabric {
	return &Fabric{
		Addr:        "40.125.164.122:10053",
		Name:        "fabric2.3",
		Username:    "Admin",
		CCID:        "Broker-001",
		ChannelId:   "mychannel",
	}
}

func UnmarshalConfig(configPath string) (*Fabric, error) {
	viper.SetConfigFile(filepath.Join(configPath, ConfigName))
	viper.SetConfigType("toml")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("FABRIC")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := DefaultConfig()

	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
