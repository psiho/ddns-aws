package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Record struct {
	Name   string
	ZoneID string
	TTL    int64
	IP     string
}

type Config struct {
	AWS_Region            string
	AWS_Access_key_id     string
	AWS_Secret_access_key string

	Server_Port       int
	Server_Username   string
	Server_Password   string
	Server_Cert       string
	Server_PrivateKey string

	Active []Record
}

func initConfig() {
	viper.SetEnvPrefix("DDNS_AWS")
	viper.BindEnv("Server_Port")
	viper.BindEnv("Server_Username")
	viper.BindEnv("Server_Password")
	viper.BindEnv("Server_Cert")
	viper.BindEnv("Server_PrivateKey")
	viper.BindEnv("AWS_Region", "AWS_REGION")
	viper.BindEnv("AWS_Access_key_id", "AWS_ACCESS_KEY_ID")
	viper.BindEnv("AWS_Secret_access_key", "AWS_SECRET_ACCESS_KEY")

	viper.SetConfigName(".ddns-aws")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath("/etc")

	viper.SetDefault("Active", []Record{})
	viper.SetDefault("Server_Port", 8125)
	viper.SetDefault("Server_Username", "ddns")
	viper.SetDefault("Server_Password", "ddns")
	viper.SetDefault("Server_Cert", "")
	viper.SetDefault("Server_PrivateKey", "")
}

func loadConfig() {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("Config file not found!")
		} else {
			log.Fatal(err)
		}
	}
	var temp Config // workarround, for some reason unmarshalling empty config doesn't update conf
	err := viper.Unmarshal(&temp)
	conf = temp
	if err != nil {
		log.Fatal(err)
	}

	//setup env variables for AWS authentication
	os.Setenv("AWS_REGION", viper.GetString("AWS_Region"))
	os.Setenv("AWS_ACCESS_KEY_ID", viper.GetString("AWS_Access_key_id"))
	os.Setenv("AWS_SECRET_ACCESS_KEY", viper.GetString("AWS_Secret_access_key"))

	return
}

func writeConfig() {
	err := viper.WriteConfig()
	if err != nil {
		fmt.Printf("Error writing config: %+s \n", err)
	}
}

func startConfigWatcher() {
	viper.OnConfigChange(func(e fsnotify.Event) {
		if e.Op != fsnotify.Write {
			return
		}
		loadConfig()
	})
	viper.WatchConfig()
}

func updateConfigActive() {
	viper.Set("Active", conf.Active)
	writeConfig()
}
