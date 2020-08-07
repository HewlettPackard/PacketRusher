//+build !debug

/*
 * Namf_Communication
 *
 * AMF Communication Service
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package Namf_Communication

import (
	"crypto/tls"
	"net/http"

	"golang.org/x/net/http2"
)

// APIClient manages communication with the Namf_Communication API v1.0.0
// In most cases there should be only one, shared, APIClient.
type APIClient struct {
	cfg    *Configuration
	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// API Services
	IndividualSubscriptionDocumentApi                             *IndividualSubscriptionDocumentApiService
	IndividualUeContextDocumentApi                                *IndividualUeContextDocumentApiService
	N1N2IndividualSubscriptionDocumentApi                         *N1N2IndividualSubscriptionDocumentApiService
	N1N2MessageCollectionDocumentApi                              *N1N2MessageCollectionDocumentApiService
	N1N2SubscriptionsCollectionForIndividualUEContextsDocumentApi *N1N2SubscriptionsCollectionForIndividualUEContextsDocumentApiService
	N1N2MessageTransferStatusNotificationCallbackDocumentApi      *N1N2MessageTransferStatusNotificationCallbackDocumentApiService
	NonUEN2MessageNotificationIndividualSubscriptionDocumentApi   *NonUEN2MessageNotificationIndividualSubscriptionDocumentApiService
	NonUEN2MessagesCollectionDocumentApi                          *NonUEN2MessagesCollectionDocumentApiService
	NonUEN2MessagesSubscriptionsCollectionDocumentApi             *NonUEN2MessagesSubscriptionsCollectionDocumentApiService
	SubscriptionsCollectionDocumentApi                            *SubscriptionsCollectionDocumentApiService
	N1MessageNotifyCallbackDocumentApiServiceCallbackDocumentApi  *N1MessageNotifyCallbackDocumentApiService
	N2InfoNotifyCallbackDocumentApiServiceCallbackDocumentApi     *N2InfoNotifyCallbackDocumentApiService
	N2MessageNotifyCallbackDocumentApiServiceCallbackDocumentApi  *N2MessageNotifyCallbackDocumentApiService
	AmfStatusChangeCallbackDocumentApiServiceCallbackDocumentApi  *AmfStatusChangeCallbackDocumentApiService
}

type service struct {
	client *APIClient
}

// NewAPIClient creates a new API client. Requires a userAgent string describing your application.
// optionally a custom http.Client to allow for advanced features such as caching.
func NewAPIClient(cfg *Configuration) *APIClient {
	if cfg.httpClient == nil {
		cfg.httpClient = http.DefaultClient
		cfg.httpClient.Transport = &http2.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	c := &APIClient{}
	c.cfg = cfg
	c.common.client = c

	// API Services
	c.IndividualSubscriptionDocumentApi = (*IndividualSubscriptionDocumentApiService)(&c.common)
	c.IndividualUeContextDocumentApi = (*IndividualUeContextDocumentApiService)(&c.common)
	c.N1N2IndividualSubscriptionDocumentApi = (*N1N2IndividualSubscriptionDocumentApiService)(&c.common)
	c.N1N2MessageCollectionDocumentApi = (*N1N2MessageCollectionDocumentApiService)(&c.common)
	c.N1N2SubscriptionsCollectionForIndividualUEContextsDocumentApi = (*N1N2SubscriptionsCollectionForIndividualUEContextsDocumentApiService)(&c.common)
	c.N1N2MessageTransferStatusNotificationCallbackDocumentApi = (*N1N2MessageTransferStatusNotificationCallbackDocumentApiService)(&c.common)
	c.NonUEN2MessageNotificationIndividualSubscriptionDocumentApi = (*NonUEN2MessageNotificationIndividualSubscriptionDocumentApiService)(&c.common)
	c.NonUEN2MessagesCollectionDocumentApi = (*NonUEN2MessagesCollectionDocumentApiService)(&c.common)
	c.NonUEN2MessagesSubscriptionsCollectionDocumentApi = (*NonUEN2MessagesSubscriptionsCollectionDocumentApiService)(&c.common)
	c.SubscriptionsCollectionDocumentApi = (*SubscriptionsCollectionDocumentApiService)(&c.common)
	c.N1MessageNotifyCallbackDocumentApiServiceCallbackDocumentApi = (*N1MessageNotifyCallbackDocumentApiService)(&c.common)
	c.N2InfoNotifyCallbackDocumentApiServiceCallbackDocumentApi = (*N2InfoNotifyCallbackDocumentApiService)(&c.common)
	c.N2MessageNotifyCallbackDocumentApiServiceCallbackDocumentApi = (*N2MessageNotifyCallbackDocumentApiService)(&c.common)
	c.AmfStatusChangeCallbackDocumentApiServiceCallbackDocumentApi = (*AmfStatusChangeCallbackDocumentApiService)(&c.common)
	return c
}
