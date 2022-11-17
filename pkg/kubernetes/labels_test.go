// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertResourceTypeToLabelValue(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		want         string
	}{
		{
			name:         "valid-resource-type",
			resourceType: "Applications.Core/containers",
			want:         "Applications.Core-containers",
		},
		{
			name:         "invalid-resource-type",
			resourceType: "Applications.Core.containers",
			want:         "Applications.Core.containers",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertResourceTypeToLabelValue(tt.resourceType); got != tt.want {
				t.Errorf("ConvertResourceTypeToLabelValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertLabelToResourceType(t *testing.T) {
	tests := []struct {
		name       string
		labelValue string
		want       string
	}{
		{
			name:       "valid-label-value",
			labelValue: "applications.core-containers",
			want:       "applications.core/containers",
		},
		{
			name:       "invalid-label-value",
			labelValue: "Applications.Core.containers",
			want:       "Applications.Core.containers",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertLabelToResourceType(tt.labelValue); got != tt.want {
				t.Errorf("ConvertLabelToResourceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeResoureName(t *testing.T) {
	nameTests := []struct {
		in  string
		out string
	}{
		{
			"resource",
			"resource",
		},
		{
			"Resource",
			"resource",
		},
		{
			"",
			"panic",
		},
	}

	for _, tt := range nameTests {
		t.Run(tt.in, func(t *testing.T) {
			if tt.in == "" {
				require.Panics(t, func() {
					NormalizeResourceName(tt.in)
				})
			} else {
				require.Equal(t, tt.out, NormalizeResourceName(tt.in))
			}
		})
	}
}
