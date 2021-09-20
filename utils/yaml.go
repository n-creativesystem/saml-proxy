package utils

import (
	"errors"
	"fmt"
	"os"

	"github.com/ghodss/yaml"
	"github.com/spf13/viper"
)

func LoadYAML(configFilename string, into interface{}) error {
	v := viper.New()
	v.SetConfigFile(configFilename)
	v.SetConfigType("yaml")
	v.SetTypeByDefaultValue(true)
	if configFilename == "" {
		return errors.New("no configuration file")
	}
	buf, err := os.ReadFile(configFilename)
	if err != nil {
		return fmt.Errorf("unable to load config file: %w", err)
	}
	if err := yaml.UnmarshalStrict(buf, into, yaml.DisallowUnknownFields); err != nil {
		return fmt.Errorf("error unmarshal config: %w", err)
	}
	return nil
}
