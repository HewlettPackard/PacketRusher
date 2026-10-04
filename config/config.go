/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package config

import (
	"crypto/ecdh"
	"encoding/hex"
	"fmt"
	"my5G-RANTester/internal/common/sidf"
	"os"
	"path"
	"path/filepath"
	"strconv"

	nasType "github.com/free5gc/nas/ie"
	"github.com/goccy/go-yaml"
	log "github.com/sirupsen/logrus"
)

// TunnelMode indicates how to create a GTP-U tunnel interface in an UE.
type TunnelMode int

const (
	// TunnelDisabled disables the GTP-U tunnel.
	TunnelDisabled TunnelMode = iota
	// TunnelPlain creates a TUN device only.
	TunnelTun
	// TunnelPlain creates a TUN device and a VRF device.
	TunnelVrf
	// TunnelShared creates one TUN device per gNB rather than per UE, so that many
	// UEs can carry traffic through a single gNB. Routing is policy-rule based: a
	// device can only have one master, so per-UE VRF enslavement is not available
	// in this mode.
	TunnelShared
)

// TunnelBackend selects the datapath independently of UE routing mode.
type TunnelBackend string

const (
	// Auto tries eBPF and permits a setup-only userspace fallback after rollback.
	TunnelBackendAuto      TunnelBackend = "auto"
	TunnelBackendKernel    TunnelBackend = "gtp5g"
	TunnelBackendUserspace TunnelBackend = "userspace"
	TunnelBackendEBPF      TunnelBackend = "ebpf"
)

func ParseTunnelBackend(value string) (TunnelBackend, error) {
	switch value {
	case "", string(TunnelBackendAuto):
		return TunnelBackendAuto, nil
	case string(TunnelBackendKernel):
		return TunnelBackendKernel, nil
	case string(TunnelBackendUserspace):
		return TunnelBackendUserspace, nil
	case string(TunnelBackendEBPF):
		return TunnelBackendEBPF, nil
	default:
		return "", fmt.Errorf("tunnel backend %q must be auto, gtp5g, userspace, or ebpf", value)
	}
}

// OverrideTunnelBackend applies a CLI override to the loaded configuration.
func OverrideTunnelBackend(backend TunnelBackend) Config {
	config.Ue.TunnelBackend = backend
	return *config
}

var config *Config

type Config struct {
	GNodeB GNodeB `yaml:"gnodeb"`
	Ue     Ue     `yaml:"ue"`
	AMFs   []*AMF `yaml:"amfif"`
	Logs   Logs   `yaml:"logs"`
}

type GNodeB struct {
	ControlIF        IPv4Port         `yaml:"controlif"`
	DataIF           IPv4Port         `yaml:"dataif"`
	PlmnList         PlmnList         `yaml:"plmnlist"`
	SliceSupportList SliceSupportList `yaml:"slicesupportlist"`
}

type PlmnList struct {
	Mcc         string `yaml:"mcc"`
	Mnc         string `yaml:"mnc"`
	Tac         string `yaml:"tac"`
	GnbId       string `yaml:"gnbid"`
	GnbIDLength uint8  `yaml:"gnbidlength"`
	CellID      uint16 `yaml:"cellid"`
}
type SliceSupportList struct {
	Sst string `yaml:"sst"`
	Sd  string `yaml:"sd"`
}

type Ue struct {
	Msin                   string         `yaml:"msin"`
	Key                    string         `yaml:"key"`
	Opc                    string         `yaml:"opc"`
	Amf                    string         `yaml:"amf"`
	Sqn                    string         `yaml:"sqn"`
	Dnn                    string         `yaml:"dnn"`
	ProtectionScheme       int            `yaml:"protectionScheme"`
	HomeNetworkPublicKey   string         `yaml:"homeNetworkPublicKey"`
	HomeNetworkPublicKeyID uint8          `yaml:"homeNetworkPublicKeyID"`
	RoutingIndicator       string         `yaml:"routingindicator"`
	Hplmn                  Hplmn          `yaml:"hplmn"`
	Snssai                 Snssai         `yaml:"snssai"`
	Integrity              Integrity      `yaml:"integrity"`
	Ciphering              Ciphering      `yaml:"ciphering"`
	TunnelMode             TunnelMode     `yaml:"-"`
	TunnelMTU              int            `yaml:"tunnelmtu"`
	PDUSessionType         PDUSessionType `yaml:"pdusessiontype"`
	TunnelBackend          TunnelBackend  `yaml:"tunnelbackend"`
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
	IPv4Port
}

type Logs struct {
	Level int `yaml:"level"`
}

func GetConfig() Config {
	if config == nil {
		LoadDefaultConfig()
	}
	return *config
}

func LoadDefaultConfig() Config {
	return Load(getDefautlConfigPath())
}

func Load(configPath string) Config {
	c := readConfig(configPath)
	config = &c

	setLogLevel(*config)
	log.Info("Loaded config at: ", configPath)
	return *config
}

func readConfig(configPath string) Config {
	var cfg = Config{}
	f, err := os.Open(configPath)
	if err != nil {
		log.Fatal("Could not open config at \"", configPath, "\". ", err.Error())
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f, yaml.Strict())
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Fatal("Could not unmarshal yaml config at \"", configPath, "\". ", err.Error())
	}

	backend, err := ParseTunnelBackend(string(cfg.Ue.TunnelBackend))
	if err != nil {
		log.Fatal(err)
	}
	cfg.Ue.TunnelBackend = backend

	if cfg.Ue.TunnelMTU < 0 {
		log.Fatal("ue.tunnelmtu must be zero (automatic) or a positive IP MTU")
	}
	identity, err := cfg.GNodeB.PlmnList.Identity()
	if err != nil {
		log.Fatalf("invalid gnodeb.plmnlist: %v", err)
	}
	cfg.GNodeB.PlmnList.GnbId = identity.String()
	cfg.GNodeB.PlmnList.GnbIDLength = identity.BitLength

	sqn, err := strconv.ParseInt(cfg.Ue.Sqn, 16, 64)
	if err != nil {
		log.Fatalf("sqn[%s] is invalid: %v", cfg.Ue.Sqn, err)
	}
	cfg.Ue.Sqn = fmt.Sprintf("%012X", sqn)

	return cfg
}

func getDefautlConfigPath() string {
	b, err := os.Executable()
	if err != nil {
		log.Fatal("Failed to get executable path. ", err.Error())
	}
	dir := path.Dir(b)
	configPath, err := filepath.Abs(dir + "/config/config.yml")
	if err != nil {
		log.Fatal("Could not find defautl config at \"", configPath, "\". ", err.Error())
	}
	return configPath
}

func setLogLevel(cfg Config) {
	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)

	if cfg.Logs.Level == 0 {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.Level(cfg.Logs.Level))
	}

}

func (config *Config) GetUESecurityCapability() *nasType.UESecCapability {
	UESecurityCapability := &nasType.UESecCapability{
		Length:     2,
		EA05G:      config.Ue.Ciphering.Nea0,
		EA1_128_5G: config.Ue.Ciphering.Nea1,
		EA2_128_5G: config.Ue.Ciphering.Nea2,
		EA3_128_5G: config.Ue.Ciphering.Nea3,
		IA05G:      config.Ue.Integrity.Nia0,
		IA1_128_5G: config.Ue.Integrity.Nia1,
		IA2_128_5G: config.Ue.Integrity.Nia2,
		IA3_128_5G: config.Ue.Integrity.Nia3,
	}

	return UESecurityCapability
}

func (config *Config) GetHomeNetworkPublicKey() sidf.HomeNetworkPublicKey {
	switch config.Ue.ProtectionScheme {
	case 0:
		config.Ue.HomeNetworkPublicKey = ""
		config.Ue.HomeNetworkPublicKeyID = 0
	case 1:
		key, err := hex.DecodeString(config.Ue.HomeNetworkPublicKey)
		if err != nil {
			log.Fatalf("Invalid Home Network Public Key in configuration for Profile A: %v", err)
		}

		publicKey, err := ecdh.X25519().NewPublicKey(key)
		if err != nil {
			log.Fatalf("Invalid Home Network Public Key in configuration for Profile A: %v", err)
		}

		return sidf.HomeNetworkPublicKey{
			ProtectionScheme: strconv.Itoa(config.Ue.ProtectionScheme),
			PublicKey:        publicKey,
			PublicKeyID:      strconv.Itoa(int(config.Ue.HomeNetworkPublicKeyID)),
		}
	case 2:
		key, err := hex.DecodeString(config.Ue.HomeNetworkPublicKey)
		if err != nil {
			log.Fatalf("Invalid Home Network Public Key in configuration for Profile B: %v", err)
		}

		publicKey, err := ecdh.P256().NewPublicKey(key)
		if err != nil {
			log.Fatalf("Invalid Home Network Public Key in configuration for Profile B: %v", err)
		}

		return sidf.HomeNetworkPublicKey{
			ProtectionScheme: strconv.Itoa(config.Ue.ProtectionScheme),
			PublicKey:        publicKey,
			PublicKeyID:      strconv.Itoa(int(config.Ue.HomeNetworkPublicKeyID)),
		}
	default:
		log.Fatal("Invalid Protection Scheme for SUCI. Valid values are 0, 1 and 2")
	}

	return sidf.HomeNetworkPublicKey{
		ProtectionScheme: "0",
		PublicKey:        nil,
		PublicKeyID:      "0",
	}
}

func boolToUint8(boolean bool) uint8 {
	if boolean {
		return 1
	} else {
		return 0
	}
}
