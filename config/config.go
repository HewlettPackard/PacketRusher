/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package config

import (
	"io/ioutil"
	"path"
	"path/filepath"
	"runtime"

	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/nas/nasType"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Conf: Used for access to configuration
var Data = getConfig()

type Config struct {
	GNodeB GNodeB `yaml:"gnodeb"`
	Ue     Ue     `yaml:"ue"`
	AMF    AMF    `yaml:"amfif"`
	Logs   Logs   `yaml:"logs"`
}

type GNodeB struct {
	ControlIF        ControlIF        `yaml:"controlif"`
	DataIF           DataIF           `yaml:"dataif"`
	PlmnList         PlmnList         `yaml:"plmnlist"`
	SliceSupportList SliceSupportList `yaml:"slicesupportlist"`
}

type ControlIF struct {
	Ip   string `yaml:"ip"`
	Port int    `yaml:"port"`
}
type DataIF struct {
	Ip   string `yaml:"ip"`
	Port int    `yaml:"port"`
}
type PlmnList struct {
	Mcc   string `yaml:"mcc"`
	Mnc   string `yaml:"mnc"`
	Tac   string `yaml:"tac"`
	GnbId string `yaml:"gnbid"`
}
type SliceSupportList struct {
	Sst string `yaml:"sst"`
	Sd  string `yaml:"sd"`
}

type Ue struct {
	Msin             string    `yaml:"msin"`
	Key              string    `yaml:"key"`
	Opc              string    `yaml:"opc"`
	Amf              string    `yaml:"amf"`
	Sqn              string    `yaml:"sqn"`
	Dnn              string    `yaml:"dnn"`
	RoutingIndicator string    `yaml:"routingindicator"`
	Hplmn            Hplmn     `yaml:"hplmn"`
	Snssai           Snssai    `yaml:"snssai"`
	Integrity        Integrity `yaml:"integrity"`
	Ciphering        Ciphering `yaml:"ciphering"`
	TunnelEnabled    bool      `yaml:"tunnelenabled"`
}

type Hplmn struct {
	Mcc string `yaml:"mcc"`
	Mnc string `yaml:"mnc"`
}
type Snssai struct {
	Sst int    `yaml:"sst"`
	Sd  string `yaml:"sd"`
}
type Integrity struct {
	Nia0 bool `yaml:"nia0"`
	Nia1 bool `yaml:"nia1"`
	Nia2 bool `yaml:"nia2"`
	Nia3 bool `yaml:"nia3"`
}
type Ciphering struct {
	Nea0 bool `yaml:"nea0"`
	Nea1 bool `yaml:"nea1"`
	Nea2 bool `yaml:"nea2"`
	Nea3 bool `yaml:"nea3"`
}

type AMF struct {
	Ip   string `yaml:"ip"`
	Port int    `yaml:"port"`
}

type Logs struct {
	Level int `yaml:"level"`
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
	UESecurityCapability.SetEA3_128_5G(boolToUint8(config.Ue.Ciphering.Nea3))

	// Integrity algorithms
	UESecurityCapability.SetIA0_5G(boolToUint8(config.Ue.Integrity.Nia0))
	UESecurityCapability.SetIA1_128_5G(boolToUint8(config.Ue.Integrity.Nia1))
	UESecurityCapability.SetIA2_128_5G(boolToUint8(config.Ue.Integrity.Nia2))
	UESecurityCapability.SetIA3_128_5G(boolToUint8(config.Ue.Integrity.Nia3))

	return UESecurityCapability
}

func boolToUint8(boolean bool) uint8 {
	if boolean {
		return 1
	} else {
		return 0
	}
}
