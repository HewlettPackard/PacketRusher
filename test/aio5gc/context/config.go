// SPDX-License-Identifier: Apache-2.0
package context

import "my5G-RANTester/config"

// CloneConfig keeps endpoint edits in one fixture from mutating another.
func CloneConfig(cfg config.Config) config.Config {
	out := cfg
	if cfg.AMFs == nil {
		return out
	}
	out.AMFs = make([]*config.AMF, len(cfg.AMFs))
	for i, amf := range cfg.AMFs {
		if amf != nil {
			copy := *amf
			out.AMFs[i] = &copy
		}
	}
	return out
}

// Config returns the effective configuration, including ephemeral bound ports.
// Builder fills endpoints before publishing the core or starting associations.
func (a *Aio5gc) Config() config.Config {
	a.configMu.RLock()
	defer a.configMu.RUnlock()
	return CloneConfig(a.conf)
}
func (a *Aio5gc) SetAMFEndpoint(index int, amf config.AMF) {
	a.configMu.Lock()
	defer a.configMu.Unlock()
	copy := amf
	a.conf.AMFs[index] = &copy
}
