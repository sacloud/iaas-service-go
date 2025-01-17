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

package vpcrouter

import (
	"testing"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/testutil"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/iaas-service-go/setup"
	"github.com/sacloud/iaas-service-go/vpcrouter/builder"
	"github.com/stretchr/testify/require"
)

func TestVPCRouterService_convertApplyRequest(t *testing.T) {
	caller := testutil.SingletonAPICaller()
	cases := []struct {
		in     *ApplyRequest
		expect *builder.Builder
	}{
		{
			in: &ApplyRequest{
				Zone:        "tk1a",
				ID:          101,
				Name:        "name",
				Description: "desc",
				Tags:        types.Tags{"tag1", "tag2"},
				IconID:      102,
				PlanID:      types.VPCRouterPlans.Standard,
				NICSetting:  &builder.StandardNICSetting{},
				AdditionalNICSettings: []builder.AdditionalNICSettingHolder{
					&builder.AdditionalStandardNICSetting{
						SwitchID:       103,
						IPAddress:      "192.168.0.1",
						NetworkMaskLen: 24,
						Index:          1,
					},
				},
				RouterSetting: &RouterSetting{
					VRID:                      1,
					InternetConnectionEnabled: true,
					L2TPIPsecServer: &iaas.VPCRouterL2TPIPsecServer{
						RangeStart:      "192.168.0.250",
						RangeStop:       "192.168.0.254",
						PreSharedSecret: "presharedsecret",
					},
					RemoteAccessUsers: []*iaas.VPCRouterRemoteAccessUser{
						{
							UserName: "username",
							Password: "password",
						},
					},
				},
				NoWait:          true,
				BootAfterCreate: true,
			},
			expect: &builder.Builder{
				Zone: "tk1a",
				ID:   101,

				Name:        "name",
				Description: "desc",
				Tags:        types.Tags{"tag1", "tag2"},
				IconID:      102,
				PlanID:      types.VPCRouterPlans.Standard,
				NICSetting:  &builder.StandardNICSetting{},
				AdditionalNICSettings: []builder.AdditionalNICSettingHolder{
					&builder.AdditionalStandardNICSetting{
						SwitchID:       103,
						IPAddress:      "192.168.0.1",
						NetworkMaskLen: 24,
						Index:          1,
					},
				},
				RouterSetting: &builder.RouterSetting{
					VRID:                      1,
					InternetConnectionEnabled: true,
					L2TPIPsecServer: &iaas.VPCRouterL2TPIPsecServer{
						RangeStart:      "192.168.0.250",
						RangeStop:       "192.168.0.254",
						PreSharedSecret: "presharedsecret",
					},
					RemoteAccessUsers: []*iaas.VPCRouterRemoteAccessUser{
						{
							UserName: "username",
							Password: "password",
						},
					},
				},
				SetupOptions: &setup.Options{
					BootAfterBuild: true,
				},
				Client: iaas.NewVPCRouterOp(caller),
				NoWait: true,
			},
		},
	}

	for _, tc := range cases {
		require.EqualValues(t, tc.expect, tc.in.Builder(caller))
	}
}
