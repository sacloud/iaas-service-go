// Copyright 2022-2025 The sacloud/iaas-service-go Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package containerregistry

import (
	"testing"

	"github.com/sacloud/iaas-api-go/testutil"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/iaas-service-go/containerregistry/builder"
	"github.com/stretchr/testify/require"
)

func TestContainerRegistryService_convertCreateRequest(t *testing.T) {
	name := testutil.ResourceName("container-registry-service")
	cases := []struct {
		in     *CreateRequest
		expect *ApplyRequest
	}{
		{
			in: &CreateRequest{
				Name:           name,
				Description:    "desc",
				Tags:           types.Tags{"tag1", "tag2"},
				IconID:         1,
				AccessLevel:    types.ContainerRegistryAccessLevels.ReadWrite,
				VirtualDomain:  "container-registry.test.libsacloud.com",
				SubDomainLabel: name,
				Users: []*builder.User{
					{
						UserName:   "user1",
						Password:   "password1",
						Permission: types.ContainerRegistryPermissions.ReadWrite,
					},
				},
			},
			expect: &ApplyRequest{
				ID:             0,
				Name:           name,
				Description:    "desc",
				Tags:           types.Tags{"tag1", "tag2"},
				IconID:         1,
				AccessLevel:    types.ContainerRegistryAccessLevels.ReadWrite,
				VirtualDomain:  "container-registry.test.libsacloud.com",
				SubDomainLabel: name,
				Users: []*builder.User{
					{
						UserName:   "user1",
						Password:   "password1",
						Permission: types.ContainerRegistryPermissions.ReadWrite,
					},
				},
				SettingsHash: "",
			},
		},
	}

	for _, tc := range cases {
		require.EqualValues(t, tc.expect, tc.in.ApplyRequest())
	}
}
