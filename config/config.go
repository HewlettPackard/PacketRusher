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
			St  string `yaml: "st"`
			Sst string `yaml: "sst"`
		} `yaml: "slicesupportlist"`
	} `yaml:"gnodeb"`
	Ue struct {
		Imsi  string `yaml: "imsi"`
		Key   string `yaml: "key"`
		Opc   string `yaml: "opc"`
		Amf   string `yaml: "amf"`
		Hplmn struct {
			Mcc string `yaml: "mcc"`
			Mnc string `yaml: "mnc"`
		} `yaml: "hplmn"`
		Snssai struct {
			Sd  int    `yaml: "sd"`
			Sst string `yaml: "sst"`
		} `yaml: "snssai"`
		Ping string `yaml: "ping"`
	} `yaml:"ue"`
	AMF struct {
		Ip   string `yaml: "ip"`
		Port int    `yaml: "port"`
		Name string `yaml: "name"`
	} `yaml:"amfif"`
	UPF struct {
		Ip   string `yaml: "ip"`
		Port int    `yaml: "port"`
	} `yaml:"upfif"`
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
