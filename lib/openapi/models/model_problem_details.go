/*
 * Nsmf_EventExposure
 *
 * Session Management Event Exposure Service API
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package models

type ProblemDetails struct {
	Type          string         `json:"type,omitempty" yaml:"type" bson:"type" mapstructure:"Type"`
	Title         string         `json:"title,omitempty" yaml:"title" bson:"title" mapstructure:"Title"`
	Status        int32          `json:"status,omitempty" yaml:"status" bson:"status" mapstructure:"Status"`
	Detail        string         `json:"detail,omitempty" yaml:"detail" bson:"detail" mapstructure:"Detail"`
	Instance      string         `json:"instance,omitempty" yaml:"instance" bson:"instance" mapstructure:"Instance"`
	Cause         string         `json:"cause,omitempty" yaml:"cause" bson:"cause" mapstructure:"Cause"`
	InvalidParams []InvalidParam `json:"invalidParams,omitempty" yaml:"invalidParams" bson:"invalidParams" mapstructure:"InvalidParams"`
}
