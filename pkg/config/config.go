package config

import (
	"github.com/spf13/viper"
	"os"
	"sync"
	"uwdiscorwb/v1/pkg/log"
	"uwdiscorwb/v1/pkg/types"
)

var (
	once sync.Once
	config *types.Properties
	vp *viper.Viper

)

// New provides a singleton for creating the configuration
// Once handles the cases where multiple routines are trying
// to initialize the config file
func GetConfig() (c *types.Properties, err error) {
	if config != nil {
		return config, nil
	}
	once.Do(func() {
		c, err = initializeConfig(os.Getenv("CONFIG_FILE_PATH"))
	})
	return
}

// initializeConfig will read the config file from home directory
// and decodes it into Configuration structure
func initializeConfig(path string) (*types.Properties, error)  {
	vp = viper.New()
	vp.SetConfigName("config")
	vp.SetConfigType("yaml")
	vp.AddConfigPath("$HOME/")
	vp.AddConfigPath(path)
	vp.AddConfigPath(".")
	vp.AutomaticEnv()

	vp.SetDefault("port", 8080)
	vp.SetDefault("discord_webhook_url", os.Getenv("discord_webhook_url"))
	vp.SetDefault("secure_server", true)

	if err := vp.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info("config file not provided, starting with default configurations")
		}
	}
	err := vp.Unmarshal(&config)
	if err != nil {
		return config, err
	}
	return config, nil
}
