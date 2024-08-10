package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path"
	"tablet_mapper/inputs"
)

const configFileName = ".tablet-mapper.conf"

type TabletMapperConfig map[string]inputs.InputConfig
func ReadConfigFromFile(confPath string) (TabletMapperConfig, error) {
	if file, err := os.Open(confPath); err != nil {
		log.Printf("WARN: couldn't read from config file '%s'. %s", confPath, err.Error())
	} else {
		defer file.Close()
		if buf, err := io.ReadAll(file); err != nil {
			log.Printf("WARN: read config %s", err.Error())
		} else {
			var config TabletMapperConfig
			if err = json.Unmarshal(buf, &config); err != nil {
				log.Printf("ERROR: couldn't read config file '%s'. %s", confPath, err.Error())
			} else {
				//log.Printf("INFO: read config %v", config)
				return config, nil
			}
		}
	}
	return nil, fmt.Errorf("ERROR: Couldn't read config file '%s'", confPath)

}

func GetDefaultConfpath() (string, error) {
	if user, err := user.Current(); err != nil {
		log.Printf("ERROR: couldn't read current user %s ", err.Error())
	} else {
		confPath := path.Join(user.HomeDir, configFileName)
		log.Printf("INFO: reading from file %s", confPath)
		return confPath, nil
	}
	return "", fmt.Errorf("ERROR: Couldn't get default config file %s", configFileName)

}

func WriteConfig(config TabletMapperConfig) {
	if user, err := user.Current(); err != nil {
		log.Printf("ERROR: couldn't read current user %s ", err.Error())
	} else {
		confPath := path.Join(user.HomeDir, configFileName)
		log.Printf("INFO: writing to file %s", confPath)
		if file, err := os.Create(confPath); err != nil {
			log.Printf("ERROR: couldn't write to config file %s. %s", confPath, err.Error())
		} else {
			defer file.Close()
			buf, _ := json.MarshalIndent(config, "", "  ")
			if _, err = file.Write(buf); err != nil {
				log.Printf("ERROR: couldn't write to config file %s. %s", confPath, err.Error())
			}
			//log.Printf("INFO: wrote config: %s",buf )
			file.Sync()
		}
	}
}
