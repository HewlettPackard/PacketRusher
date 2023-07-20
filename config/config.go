package config

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/nas/nasType"
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
		RoutingIndicator string `yaml: "routingindicator"`
		Hplmn struct {
			Mcc string `yaml: "mcc"`
			Mnc string `yaml: "mnc"`
		} `yaml: "hplmn"`
		Snssai struct {
			Sst int    `yaml: "sst"`
			Sd  string `yaml: "sd"`
		} `yaml: "snssai"`
		Integrity struct {
			Nia0 bool `yaml: "nia0"`
			Nia1 bool `yaml: "nia1"`
			Nia2 bool `yaml: "nia2"`
		} `yaml: "integrity"`
		Ciphering struct {
			Nea0 bool `yaml: "nea0"`
			Nea1 bool `yaml: "nea1"`
			Nea2 bool `yaml: "nea2"`
		} `yaml: "ciphering"`
		TunnelEnabled bool `yaml: "tunnelenabled"`
	} `yaml:"ue"`
	AMF struct {
		Ip   string `yaml: "ip"`
		Port int    `yaml: "port"`
	} `yaml:"amfif"`
	Logs struct {
		Level int `yaml: "level"`
	} `yaml:"logs"`
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

func (config *Config) GetSockPath() string {
	return "/tmp/" + config.GNodeB.PlmnList.Mcc + config.GNodeB.PlmnList.Mnc + config.GNodeB.PlmnList.GnbId + ".sock"
}

func (config *Config) GetUESecurityCapability() *nasType.UESecurityCapability {
	UESecurityCapability := &nasType.UESecurityCapability{
		Iei:    nasMessage.RegistrationRequestUESecurityCapabilityType,
		Len:    8,
		Buffer: []uint8{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}

	// Ciphering algorithms
	UESecurityCapability.SetEA0_5G(boolToUint8(config.Ue.Ciphering.Nea0))
	UESecurityCapability.SetEA1_128_5G(boolToUint8(config.Ue.Ciphering.Nea1))
	UESecurityCapability.SetEA2_128_5G(boolToUint8(config.Ue.Ciphering.Nea2))
	UESecurityCapability.SetEA3_128_5G(0)

	// Integrity algorithms
	UESecurityCapability.SetIA0_5G(boolToUint8(config.Ue.Integrity.Nia0))
	UESecurityCapability.SetIA1_128_5G(boolToUint8(config.Ue.Integrity.Nia1))
	UESecurityCapability.SetIA2_128_5G(boolToUint8(config.Ue.Integrity.Nia2))
	UESecurityCapability.SetIA3_128_5G(0)

	return UESecurityCapability
}

func boolToUint8(boolean bool) uint8 {
	if boolean {
		return 1
	} else {
		return 0
	}
}