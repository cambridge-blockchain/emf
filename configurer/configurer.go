package configurer

import (
	"fmt"
	"os"
	"strings"

	"path/filepath"

	"github.com/spf13/viper"
)

// Config provides a method set for extracting data out of a configuration file.
type Config interface {
	GetInt(string) int
	GetBool(string) bool
	GetInt64(string) int64
	GetString(string) string

	GetStringMapString(string) map[string]string

	SetConfigName(string)
	SetConfigType(string)
	SetEnvPrefix(string)
	SetEnvKeyReplacer(*strings.Replacer)

	AutomaticEnv()
	ReadInConfig() error
	AddConfigPath(string)
	UnmarshalKey(string, interface{}, ...viper.DecoderConfigOption) error
}

// ConfigReader defines the interface for read-only config file access
type ConfigReader interface {
	GetInt(string) int
	GetBool(string) bool
	GetInt64(string) int64
	GetString(string) string
	UnmarshalKey(string, interface{}, ...viper.DecoderConfigOption) error
}

// DomainsReader defines the interface for read-only map[string]string access
type DomainsReader interface {
	GetStringMapString(string) map[string]string
}

// BuildConfig provides the logger and the info endpoint access to hard-coded build configuration
type BuildConfig struct {
	Build                 string
	Component             string
	Version               string
	EMFVersion            string
	EchoVersion           string
	ReleaseTimestamp      string
	AdditionalInformation []DataPoint
}

// DataPoint represents additional information component wants to display
type DataPoint struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// LoadConfig loads a configuration file specified by the function argument or a path provided the command line.
func LoadConfig(configFile string, prefix string) Config {
	var err error

	var configPath string
	var configName string

	var rootConfig Config

	if configFile == "" {
		configName = "config.yaml"
		configPath = os.Getenv("GOPATH") + "/src/github.com/cambridge-blockchain/emf"
	} else {
		configPath, configName = filepath.Split(configFile)
	}

	rootConfig = viper.New()
	configName = strings.TrimSuffix(configName, filepath.Ext(configName))

	rootConfig.SetConfigType("yaml")
	rootConfig.SetConfigName(configName)
	rootConfig.AddConfigPath(os.ExpandEnv(configPath))
	rootConfig.SetEnvPrefix(prefix)
	rootConfig.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	rootConfig.AutomaticEnv()

	if err = rootConfig.ReadInConfig(); err != nil {
		fmt.Printf("[error] configuration file could not be read: %v", err)
		const exitCode = 1
		os.Exit(exitCode)
	}

	return rootConfig
}
