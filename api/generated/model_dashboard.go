// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

/*
 * Electricity Maps app API
 *
 * API to access and configure the Electricity Maps app
 *
 * API version: 1.0.0
 */

package apiserver

// Dashboard - A frontend dashboard
type Dashboard struct {

	// The internal Id of dashboard
	Id *int32 `json:"id,omitempty"`

	// The name for this dashboard
	Name string `json:"name"`

	// ID of the project to which the dashboard belongs
	ProjectId string `json:"projectId"`

	// ID of the user who owns the dashboard
	UserId string `json:"userId"`

	// The sequence of the dashboard
	// Deprecated
	Sequence *int32 `json:"sequence,omitempty"`

	// List of widgets on this dashboard (order matches the order of widgets on the dashboard)
	Widgets *[]Widget `json:"widgets,omitempty"`

	// Is the dashboard public and not bound to a dedicated user
	Public *bool `json:"public,omitempty"`
}

// AssertDashboardRequired checks if the required fields are not zero-ed
func AssertDashboardRequired(obj Dashboard) error {
	elements := map[string]interface{}{
		"name":      obj.Name,
		"projectId": obj.ProjectId,
		"userId":    obj.UserId,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	if obj.Widgets != nil {
		for _, el := range *obj.Widgets {
			if err := AssertWidgetRequired(el); err != nil {
				return err
			}
		}
	}
	return nil
}

// AssertDashboardConstraints checks if the values respects the defined constraints
func AssertDashboardConstraints(obj Dashboard) error {
	if obj.Widgets != nil {
		for _, el := range *obj.Widgets {
			if err := AssertWidgetConstraints(el); err != nil {
				return err
			}
		}
	}
	return nil
}
