package resource

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// ModifyPlanRequest represents a request for the provider to modify the
// planned new state that Terraform has generated for the resource.
type ModifyPlanRequest struct {
	// Config is the configuration the user supplied for the resource.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	Config tfsdk.Config

	// State is the current state of the resource.
	State tfsdk.State

	// Plan is the planned new state for the resource. Terraform 1.3 and later
	// supports resource destroy planning, in which this will contain a null
	// value.
	Plan tfsdk.Plan

	// ProviderMeta is metadata from the provider_meta block of the module.
	ProviderMeta tfsdk.Config
}

// ModifyPlanResponse represents a response to a
// ModifyPlanRequest. An instance of this response struct is supplied
// as an argument to the resource's ModifyPlan function, in which the provider
// should modify the Plan and populate the RequiresReplace field as appropriate.
type ModifyPlanResponse struct {
	// Plan is the planned new state for the resource.
	Plan tfsdk.Plan

	// RequiresReplace is a list of attribute paths that require the
	// resource to be replaced. They should point to the specific field
	// that changed that requires the resource to be destroyed and
	// recreated.
	RequiresReplace path.Paths

	// Diagnostics report errors or warnings related to determining the
	// planned state of the requested resource. Returning an empty slice
	// indicates a successful plan modification with no warnings or errors
	// generated.
	Diagnostics diag.Diagnostics
}