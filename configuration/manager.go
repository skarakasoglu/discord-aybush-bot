package configuration

import (
	"github.com/spf13/viper"
	"log"
)

var (
	Manager manager
)

type manager struct{
	Credentials credentials
}

func ReadConfigurationFile(path string, fileName string) {
	viper.SetConfigName(fileName)
	viper.AddConfigPath(path)

	if err := viper.ReadInConfig(); err !=  nil {
		log.Fatalf("Error reading configuration file: %v", err)
	}

	err := viper.Unmarshal(&Manager)
	if err != nil {
		log.Fatalf("Error decoding configuration file into struct: %v", err)
	}
}

type credentials struct{
	Token string
}

func (c credentials) GetToken() string {
	return c.Token
}