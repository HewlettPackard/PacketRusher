package flowdesc

import (
	"errors"
	"net"
	"regexp"
	"strconv"
	"strings"
)

type IPFilterRule interface {
	// initial structure
	//Init() error
	// Set action of IPFilterRule
	SetAction(bool) error
	// Set Direction of IPFilterRule
	SetDirection(bool) error
	// Set Protocol of IPFilterRule
	// 0xfc stand for ip (any)
	SetProtocal(int) error
	// Set Source IP of IPFilterRule
	// format: IP or IP/mask or "any"
	SetSourceIp(string) error
	// Set Source port of IPFilterRule
	// format: {port/port-port}[,ports[,...]]
	SetSourcePorts(string) error
	// Set Destination IP of IPFilterRule
	// format: IP or IP/mask or "assigned"
	SetDestinationIp(string) error
	// Set Destination port of IPFilterRule
	// format: {port/port-port}[,ports[,...]]
	SetDestinationPorts(string) error
	// Encode the IPFilterRule
	Encode() (string, error)
	// Decode the IPFilterRule
	Decode() error
}

type ipFilterRule struct {
	action   bool   // true: permit, false: deny
	dir      bool   // false: in, true: out
	proto    int    // protocal number
	srcIp    string // <address/mask>
	srcPorts string // [ports]
	dstIp    string // <address/mask>
	dstPorts string // [ports]
}

func NewIPFilterRule() *ipFilterRule {
	r := &ipFilterRule{
		action:   true,
		dir:      true,
		proto:    0xfc,
		srcIp:    "",
		srcPorts: "",
		dstIp:    "",
		dstPorts: "",
	}
	return r
}

func (r *ipFilterRule) SetAction(action bool) error {
	if !action {
		return errors.New("Action cannot be deny")
	}
	r.action = action
	return nil
}

func (r *ipFilterRule) SetDirection(dir bool) error {
	if !dir {
		return errors.New("Direction cannot be in")
	}
	r.dir = dir
	return nil
}

func (r *ipFilterRule) SetProtocal(proto int) error {
	r.proto = proto
	return nil
}

func (r *ipFilterRule) SetSourceIp(networkStr string) error {
	if networkStr == "any" {
		r.srcIp = networkStr
		return nil
	}
	if networkStr[0] == '!' {
		return errors.New("Base on TS 29.212, ! expression shall not be used")
	}

	var ipStr string

	ip := net.ParseIP(networkStr)
	if ip == nil {
		_, ipNet, err := net.ParseCIDR(networkStr)
		if err != nil {
			return errors.New("Source IP format error")
		}
		ipStr = ipNet.String()
	} else {
		ipStr = ip.String()
	}

	r.srcIp = ipStr
	return nil
}

func (r *ipFilterRule) SetSourcePorts(ports string) error {

	if ports == "" {
		return nil
	}

	match, err := regexp.MatchString("^[0-9]+(-[0-9]+)?(,[0-9]+)*$", ports)

	if err != nil {
		return errors.New("Regex match error")
	}
	if !match {
		return errors.New("Ports format error")
	}

	// Check port range
	portSlice := regexp.MustCompile(`[\\,\\-]+`).Split(ports, -1)
	for _, portStr := range portSlice {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return err
		}
		if 0 > int(port) || port > 65535 {
			return errors.New("Invalid port number")
		}
	}

	r.srcPorts = ports
	return nil
}

func (r *ipFilterRule) SetDestinationIp(networkStr string) error {
	if networkStr == "assigned" {
		r.dstIp = networkStr
		return nil
	}
	if networkStr[0] == '!' {
		return errors.New("Base on TS 29.212, ! expression shall not be used")
	}

	var ipDst string

	ip := net.ParseIP(networkStr)
	if ip == nil {
		_, ipNet, err := net.ParseCIDR(networkStr)
		if err != nil {
			return errors.New("Source IP format error")
		}
		ipDst = ipNet.String()
	} else {
		ipDst = ip.String()
	}

	r.dstIp = ipDst
	return nil
}

func (r *ipFilterRule) SetDestinationPorts(ports string) error {

	if ports == "" {
		return nil
	}

	match, err := regexp.MatchString("^[0-9]+(-[0-9]+)?(,[0-9]+)*$", ports)

	if err != nil {
		return errors.New("Regex match error")
	}
	if !match {
		return errors.New("Ports format error")
	}

	// Check port range
	portSlice := regexp.MustCompile(`[\\,\\-]+`).Split(ports, -1)
	for _, portStr := range portSlice {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return err
		}
		if 0 > int(port) || port > 65535 {
			return errors.New("Invalid port number")
		}
	}

	r.dstPorts = ports
	return nil
}

func (r *ipFilterRule) Encode() (string, error) {
	var ipFilterRuleStr []string

	// action
	if r.action {
		ipFilterRuleStr = append(ipFilterRuleStr, "permit")
	} else {
		ipFilterRuleStr = append(ipFilterRuleStr, "deny")
	}

	// dir
	if r.dir {
		ipFilterRuleStr = append(ipFilterRuleStr, "out")
	} else {
		ipFilterRuleStr = append(ipFilterRuleStr, "in")
	}

	// proto
	if r.proto == 0xfc {
		ipFilterRuleStr = append(ipFilterRuleStr, "ip")
	} else {
		ipFilterRuleStr = append(ipFilterRuleStr, strconv.Itoa(r.proto))
	}

	// from
	ipFilterRuleStr = append(ipFilterRuleStr, "from")

	// src
	if r.srcIp != "" {
		ipFilterRuleStr = append(ipFilterRuleStr, r.srcIp)
	} else {
		return "", errors.New("Encode error: no source IP")
	}
	if r.srcPorts != "" {
		ipFilterRuleStr = append(ipFilterRuleStr, r.srcPorts)
	}

	// to
	ipFilterRuleStr = append(ipFilterRuleStr, "to")

	// dst
	if r.dstIp != "" {
		ipFilterRuleStr = append(ipFilterRuleStr, r.dstIp)
	} else {
		return "", errors.New("Encode error: no destination IP")
	}
	if r.dstPorts != "" {
		ipFilterRuleStr = append(ipFilterRuleStr, r.dstPorts)
	}

	// [options]

	return strings.Join(ipFilterRuleStr, " "), nil
}

func (r *ipFilterRule) Decode() error {
	return nil
}
