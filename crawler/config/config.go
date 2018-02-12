package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"sync"
)

type Config struct {
	Database struct {
		DriverName  string `yaml:"drivername"`
		ConnectPool int    `yaml:"connectionpool"`
		Address     struct {
			Host string `yaml:host`
			Port string `yaml:port`
		}
		Dbname string `yaml:dbname`
		User   string `yaml:user`
		Passwd string `yaml:passwd`
	}
}

var instance *Config

var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
		yamlFile, err := ioutil.ReadFile("../config/config.yaml")
		if err != nil {
			fmt.Println(err.Error())
		}
		err = yaml.Unmarshal(yamlFile, instance)
		if err != nil {
			fmt.Println(err.Error())
		}
	})
	return instance
}

func GetDBConnectString() string {
	instance := GetConfig()
	paras := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8",
		instance.Database.User,
		instance.Database.Passwd,
		instance.Database.Address.Host,
		instance.Database.Address.Port,
		instance.Database.Dbname)
	return paras
}
