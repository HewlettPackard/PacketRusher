/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
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

// TunnelBackend is the datapath carrying the user plane of the UEs' tunnels.
type TunnelBackend string

const (
	TunnelBackendGtp5g     TunnelBackend = "gtp5g"
	TunnelBackendUserspace TunnelBackend = "userspace"
)

// PDUSessionType is the IP version of the UEs' PDU sessions, from ue.pdusessiontype.
type PDUSessionType string

const (
	PDUSessionIPv4   PDUSessionType = "IPv4"
	PDUSessionIPv6   PDUSessionType = "IPv6"
	PDUSessionIPv4v6 PDUSessionType = "IPv4v6"
)

// pduSessionTypes gives the NAS value of each type (TS 24.501 §9.11.4.11); IPv4 is
// the default.
var pduSessionTypes = map[PDUSessionType]uint8{
	"": nasType.PDUSessType_IPv4, PDUSessionIPv4: nasType.PDUSessType_IPv4,
	PDUSessionIPv6: nasType.PDUSessType_IPv6, PDUSessionIPv4v6: nasType.PDUSessType_IPv4v6,
}

func (t PDUSessionType) NAS() uint8 {
	return pduSessionTypes[t]
}

// tunnelBackends lists the backends by order of preference for "auto".
var tunnelBackends = []struct {
	name      TunnelBackend
	ipv6      bool
	available func() bool
}{
	{TunnelBackendGtp5g, false, func() bool { _, err := os.Stat("/sys/module/gtp5g"); return err == nil }},
	{TunnelBackendUserspace, true, func() bool { return true }},
}

// ResolveTunnelBackend returns the backend that --tunnel-backend designates: with
// "auto" the first available one, otherwise the named one, which must be available.
// A backend without IPv6 is not available to IPv6 PDU sessions; it carries the IPv4
// half of IPv4v6 ones.
func ResolveTunnelBackend(name string, sessionType PDUSessionType) (TunnelBackend, error) {
	for _, backend := range tunnelBackends {
		ipv6 := backend.ipv6 || sessionType != PDUSessionIPv6
		if (name == "auto" || name == string(backend.name)) && ipv6 && backend.available() {
			log.Info("[TESTER] Using the ", backend.name, " tunnel backend")
			return backend.name, nil
		}
		if name == string(backend.name) && !ipv6 {
			return "", fmt.Errorf("the %s tunnel backend cannot carry IPv6 PDU sessions: use the userspace one", name)
		}
		if name == string(backend.name) {
			return "", fmt.Errorf("the %s tunnel backend was requested but is not available on this host", name)
		}
	}
	return "", fmt.Errorf("unknown tunnel backend %q: use auto, gtp5g or userspace", name)
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
	Msin                   string     `yaml:"msin"`
	Key                    string     `yaml:"key"`
	Opc                    string     `yaml:"opc"`
	Amf                    string     `yaml:"amf"`
	Sqn                    string     `yaml:"sqn"`
	Dnn                    string     `yaml:"dnn"`
	ProtectionScheme       int        `yaml:"protectionScheme"`
	HomeNetworkPublicKey   string     `yaml:"homeNetworkPublicKey"`
	HomeNetworkPublicKeyID uint8      `yaml:"homeNetworkPublicKeyID"`
	RoutingIndicator       string     `yaml:"routingindicator"`
	Hplmn                  Hplmn      `yaml:"hplmn"`
	Snssai                 Snssai     `yaml:"snssai"`
	Integrity              Integrity  `yaml:"integrity"`
	Ciphering              Ciphering  `yaml:"ciphering"`
	TunnelMode             TunnelMode `yaml:"-"`
	TunnelMTU              int        `yaml:"tunnelmtu"`

	// PDUSessionType is IPv4 (the default), IPv6 or IPv4v6.
	PDUSessionType PDUSessionType `yaml:"pdusessiontype"`

	// TunnelBackend is resolved from --tunnel-backend when a tunnel is requested.
	TunnelBackend TunnelBackend `yaml:"-"`
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

	if _, known := pduSessionTypes[cfg.Ue.PDUSessionType]; !known {
		log.Fatal("ue.pdusessiontype must be IPv4, IPv6 or IPv4v6")
	}

	if cfg.Ue.TunnelMTU < 0 {
		log.Fatal("ue.tunnelmtu must be zero (automatic) or a positive IPv4 MTU")
	}

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
