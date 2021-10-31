package config

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"path/filepath"
	"runtime"
)

// Conf: Used for access to configuration
var Data = getConfig()

type Config struct {
	GNodeB struct {
		ControlIF struct {
			Ip   string `yaml: "ip"`
			Port int    `yaml: "port"`
		} `yaml: "controlif"`
		DataIF struct {
			Ip   string `yaml: "ip"`
			Port int    `yaml: "port"`
		} `yaml: "dataif"`
		PlmnList struct {
			Mcc   string `yaml: "mmc"`
			Mnc   string `yaml: "mnc"`
			Tac   string `yaml: "tac"`
			GnbId string `yaml: "gnbid"`
		} `yaml: "plmnlist"`
		SliceSupportList struct {
			Sst string `yaml: "sst"`
			Sd  string `yaml: "sd"`
		} `yaml: "slicesupportlist"`
	} `yaml:"gnodeb"`
	Ue struct {
		Msin  string `yaml: "msin"`
		Key   string `yaml: "key"`
		Opc   string `yaml: "opc"`
		Amf   string `yaml: "amf"`
		Sqn   string `yaml: "sqn"`
		Dnn   string `yaml: "dnn"`
		Hplmn struct {
			Mcc string `yaml: "mcc"`
			Mnc string `yaml: "mnc"`
		} `yaml: "hplmn"`
		Snssai struct {
			Sst int    `yaml: "sst"`
			Sd  string `yaml: "sd"`
		} `yaml: "snssai"`
	} `yaml:"ue"`
	AMF struct {
		Ip   string `yaml: "ip"`
		Port int    `yaml: "port"`
		Name string `yaml: "name"`
	} `yaml:"amfif"`
}

func RootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(d)
}

func getConfig() Config {
	var cfg = Config{}
	Ddir := RootDir()
	configPath, err := filepath.Abs(Ddir + "/config/config.yml")
	log.Debug(configPath)
	if err != nil {
		log.Fatal("Could not find config in: ", configPath)
	}
	file, err := ioutil.ReadFile(configPath)
	err = yaml.Unmarshal([]byte(file), &cfg)
	if err != nil {
		log.Fatal("Could not read file in: ", configPath)
	}

	return cfg
}

func GetConfig() (Config, error) {
	var cfg = Config{}
	Ddir := RootDir()
	configPath, err := filepath.Abs(Ddir + "/config/config.yml")
	log.Debug(configPath)
	if err != nil {
		return Config{}, nil
	}
	file, err := ioutil.ReadFile(configPath)
	err = yaml.Unmarshal([]byte(file), &cfg)
	if err != nil {
		return Config{}, nil
	}

	return cfg, nil
}
