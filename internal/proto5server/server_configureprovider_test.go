// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerConfigureProvider(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test": tftypes.String,
		},
	}

	testValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test": tftypes.NewValue(tftypes.String, "test-value"),
	})

	testDynamicValue, err := tfprotov5.NewDynamicValue(testType, testValue)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov5.NewDynamicValue(): %s", err)
	}

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test": schema.StringAttribute{
				Required: true,
			},
		},
	}

	type testCase struct {
		server           *Server
		request          *tfprotov5.ConfigureProviderRequest
		expectedError    error
		expectedResponse *tfprotov5.ConfigureProviderResponse
	}

	testCases := map[string]testCase{
		"nil": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{},
				},
			},
			request:          nil,
			expectedResponse: &tfprotov5.ConfigureProviderResponse{},
		},
		"no-schema": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						SchemaMethod: func(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {},
						ConfigureMethod: func(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
							// Intentially empty, test is passing if it makes it this far
						},
					},
				},
			},
			request:          &tfprotov5.ConfigureProviderRequest{},
			expectedResponse: &tfprotov5.ConfigureProviderResponse{},
		},
		"request-config": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						SchemaMethod: func(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
							resp.Schema = testSchema
						},
						ConfigureMethod: func(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
							var got types.String

							resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("test"), &got)...)

							if resp.Diagnostics.HasError() {
								return
							}

							if got.ValueString() != "test-value" {
								resp.Diagnostics.AddError("Incorrect req.Config", "expected test-value, got "+got.ValueString())
							}
						},
					},
				},
			},
			request: &tfprotov5.ConfigureProviderRequest{
				Config: &testDynamicValue,
			},
			expectedResponse: &tfprotov5.ConfigureProviderResponse{},
		},
		"request-terraformversion": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						SchemaMethod: func(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {},
						ConfigureMethod: func(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
							if req.TerraformVersion != "1.0.0" {
								resp.Diagnostics.AddError("Incorrect req.TerraformVersion", "expected 1.0.0, got "+req.TerraformVersion)
							}
						},
					},
				},
			},
			request: &tfprotov5.ConfigureProviderRequest{
				TerraformVersion: "1.0.0",
			},
			expectedResponse: &tfprotov5.ConfigureProviderResponse{},
		},
		"response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						SchemaMethod: func(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {},
						ConfigureMethod: func(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
							resp.Diagnostics.AddWarning("warning summary", "warning detail")
							resp.Diagnostics.AddError("error summary", "error detail")
						},
					},
				},
			},
			request: &tfprotov5.ConfigureProviderRequest{},
			expectedResponse: &tfprotov5.ConfigureProviderResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityWarning,
						Summary:  "warning summary",
						Detail:   "warning detail",
					},
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "error summary",
						Detail:   "error detail",
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.ConfigureProvider(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
