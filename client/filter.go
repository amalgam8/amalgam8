package client

import (
	"net/url"
	"strings"
)

// InstanceFilter is used to filter service instances returned from lookup calls.
type InstanceFilter struct {

	// ServiceName is used to filter service instances based on their service name.
	// When set to a non-empty string, registered service instances will be returned
	// only if their service name matches the specified service name.
	ServiceName string

	// Status is used to filter service instances based on their status.
	// When set to a non-empty string, registered service instances will be returned
	// only if their status matches the specified status.
	Status string

	// Tags is used to filter service instances based on their tags.
	// When set to a non-empty array, registered service instances will be returned
	// only if they are tagged with each of the specified tags.
	Tags []string

	// Fields is used to filter the fields returned for each service instance.
	// When set to a non-empty array, returned service instances will have their corresponding fields set,
	// while other fields will remain at their zero-value.
	// When set to an empty or nil array, returned service intances will have all of their fields set.
	Fields []string
}

// Enumerates available values for InstanceField.
const (
	FieldID            = "id"
	FieldServiceName   = "service_name"
	FieldEndpoint      = "endpoint"
	FieldStatus        = "status"
	FieldTags          = "tags"
	FieldMetadata      = "metadata"
	FieldTTL           = "ttl"
	FieldLastHeartbeat = "last_heartbeat"
)

// asQueryParams convert the filter into a set of query parameters that can be added to a lookup request.
func (filter *InstanceFilter) asQueryParams() url.Values {
	queryParams := make(url.Values)

	if filter.ServiceName != "" {
		queryParams.Add("service_name", filter.ServiceName)
	}

	if filter.Status != "" {
		queryParams.Add("status", filter.Status)
	}

	if len(filter.Tags) > 0 {
		queryParams.Add("tags", strings.Join(filter.Tags, ","))
	}

	if len(filter.Fields) > 0 {
		queryParams.Add("fields", strings.Join(filter.Fields, ","))
	}

	return queryParams
}
