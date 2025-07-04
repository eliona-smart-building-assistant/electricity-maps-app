// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

/*
 * Electricity Maps app API
 *
 * API to access and configure the Electricity Maps app
 *
 * API version: 1.0.0
 */

package apiserver

// FilterRule - Asset selection rule. Possible parameters are defined in app's README file.
type FilterRule struct {
	Parameter string `json:"parameter,omitempty"`

	Regex string `json:"regex,omitempty"`
}

// AssertFilterRuleRequired checks if the required fields are not zero-ed
func AssertFilterRuleRequired(obj FilterRule) error {
	return nil
}

// AssertFilterRuleConstraints checks if the values respects the defined constraints
func AssertFilterRuleConstraints(obj FilterRule) error {
	return nil
}
