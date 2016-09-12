// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package i18n

// These are currently manually kept in sync with the message file and would be better to use 'go generate'
// to produce this on the fly.

// ErrorAuthorizationMissingHeader and the other constants defined are used to avoid proliferation of strings
// in the rest of the code.
const (
	ErrorAuthorizationMissingHeader         = "error_auth_header_missing"
	ErrorAuthorizationMalformedHeader       = "error_auth_header_malformed"
	ErrorAuthorizationTokenValidationFailed = "error_auth_failed_validation"
	ErrorAuthorizationNotAuthorized         = "error_auth_not_authorized"
	ErrorEncoding                           = "error_encoding_generic"
	ErrorFilterBadFields                    = "error_filter_bad_fields"
	ErrorFilterSelectionCriteria            = "error_filter_selection_criteria"
	ErrorFilterGeneric                      = "error_filter_generic"
	ErrorInstanceEndpointMissing            = "error_instance_endpoint_missing"
	ErrorInstanceEndpontMalformed           = "error_instance_endpoint_missing_type_or_value"
	ErrorInstanceEndpointInvalidType        = "error_instance_endpoint_type_invalid"
	ErrorInstanceIdentifierMissing          = "error_instance_identifier_missing"
	ErrorInstanceEnumeration                = "error_instance_enumeration"
	ErrorInstanceMetadataInvalid            = "error_instance_metadata_invalid"
	ErrorInstanceStatusInvalid              = "error_instance_status_invalid"
	ErrorInstanceNotFound                   = "error_instance_not_found"
	ErrorInstanceDeletionFailed             = "error_instance_deletion_failure"
	ErrorInstanceHeartbeatFailed            = "error_instance_heartbeat_failure"
	ErrorInstanceRegistrationFailed         = "error_instance_registration_failure"
	ErrorInternalServer                     = "error_internal"
	ErrorInternalWithString                 = "error_internal_generic_with_string"
	ErrorNilObject                          = "error_nil_object"
	ErrorNamespaceNotFound                  = "error_namespace_undefined"
	ErrorServiceNameMissing                 = "error_service_name_missing"
	ErrorServiceEnumeration                 = "error_service_enumeration"
	ErrorServiceNotFound                    = "error_service_not_found"
)

// EurekaErrorApplicationEnumeration and other constants denote Eureka specific errors. In addition, Eureka API may
// use above, generic, error message identifiers.
const (
	EurekaErrorApplicationEnumeration       = "eureka_error_application_enumeration"
	EurekaErrorApplicationIdentifierMissing = "eureka_error_application_id_missing"
	EurekaErrorApplicationNotFound          = "eureka_error_application_not_found"
	EurekaErrorApplicationMismatch          = "eureka_error_application_mismatch"
	EurekaErrorDashboardGeneration          = "eureka_error_dashboard_generation"
	EurekaErrorRequiredFieldsMissing        = "eureka_error_required_fields_missing"
	EurekaErrorStatusMissing                = "eureka_error_status_missing"
	EurekaErrorVIPEnumeration               = "eureka_error_vip_enumeration"
	EurekaErrorVIPRequired                  = "eureka_error_vip_required"
)
