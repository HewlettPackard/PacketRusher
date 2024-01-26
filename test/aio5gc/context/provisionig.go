/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2024 Hewlett Packard Enterprise Development LP
 */
package context

import "github.com/free5gc/openapi/models"

type provisionedData struct {
	defaultSNssai   models.Snssai
	securityContext SecurityContext
}

func (p *provisionedData) GetDefaultSNssai() models.Snssai {
	return p.defaultSNssai
}

func (p *provisionedData) GetSecurityContext() SecurityContext {
	return p.securityContext
}
