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

package disk

import (
	"testing"

	"github.com/sacloud/iaas-api-go/ostype"
	"github.com/sacloud/iaas-api-go/testutil"
	"github.com/sacloud/iaas-api-go/types"
	disk "github.com/sacloud/iaas-service-go/disk/builder"
	"github.com/stretchr/testify/require"
)

func TestDiskService_convertApplyRequest(t *testing.T) {
	caller := testutil.SingletonAPICaller()

	cases := []struct {
		in     *ApplyRequest
		expect disk.Builder
	}{
		// blank
		{
			in: &ApplyRequest{
				Zone:                "is1a",
				Name:                "test",
				Description:         "description",
				Tags:                types.Tags{"tag1", "tag2"},
				IconID:              types.ID(1),
				ServerID:            types.ID(2),
				DiskPlanID:          types.DiskPlans.SSD,
				Connection:          types.DiskConnections.VirtIO,
				EncryptionAlgorithm: types.DiskEncryptionAlgorithms.AES256XTS,
				SizeGB:              20,
				DistantFrom:         nil,
				OSType:              0,
				EditParameter:       nil,
				NoWait:              true,
			},
			expect: &disk.BlankBuilder{
				Name:                "test",
				Description:         "description",
				Tags:                types.Tags{"tag1", "tag2"},
				IconID:              types.ID(1),
				SizeGB:              20,
				PlanID:              types.DiskPlans.SSD,
				Connection:          types.DiskConnections.VirtIO,
				EncryptionAlgorithm: types.DiskEncryptionAlgorithms.AES256XTS,
				Client:              disk.NewBuildersAPIClient(caller),
				NoWait:              true,
			},
		},
		// linux
		{
			in: &ApplyRequest{
				Zone:                "is1a",
				Name:                "test",
				DiskPlanID:          types.DiskPlans.SSD,
				Connection:          types.DiskConnections.VirtIO,
				EncryptionAlgorithm: types.DiskEncryptionAlgorithms.AES256XTS,
				SizeGB:              20,
				OSType:              ostype.Ubuntu,
				EditParameter: &EditParameter{
					HostName: "hostname",
					Password: "password",
				},
				NoWait: true,
			},
			expect: &disk.FromUnixBuilder{
				OSType:              ostype.Ubuntu,
				Name:                "test",
				SizeGB:              20,
				PlanID:              types.DiskPlans.SSD,
				Connection:          types.DiskConnections.VirtIO,
				EncryptionAlgorithm: types.DiskEncryptionAlgorithms.AES256XTS,
				EditParameter: &disk.UnixEditRequest{
					HostName: "hostname",
					Password: "password",
				},
				Client: disk.NewBuildersAPIClient(caller),
				NoWait: true,
				ID:     0,
			},
		},
		// source disk
		{
			in: &ApplyRequest{
				Zone:                "is1a",
				Name:                "test",
				DiskPlanID:          types.DiskPlans.SSD,
				Connection:          types.DiskConnections.VirtIO,
				EncryptionAlgorithm: types.DiskEncryptionAlgorithms.AES256XTS,
				SourceDiskID:        types.ID(1),
				SizeGB:              20,
				EditParameter: &EditParameter{
					HostName: "hostname",
					Password: "password",
				},
				NoWait: true,
			},
			expect: &disk.FromDiskOrArchiveBuilder{
				Name:                "test",
				SizeGB:              20,
				PlanID:              types.DiskPlans.SSD,
				Connection:          types.DiskConnections.VirtIO,
				EncryptionAlgorithm: types.DiskEncryptionAlgorithms.AES256XTS,
				SourceDiskID:        types.ID(1),
				EditParameter: &disk.UnixEditRequest{
					HostName: "hostname",
					Password: "password",
				},
				Client: disk.NewBuildersAPIClient(caller),
				NoWait: true,
				ID:     0,
			},
		},
	}

	for _, tc := range cases {
		builder, err := tc.in.Builder(caller)
		require.NoError(t, err)
		require.EqualValues(t, tc.expect, builder)
	}
}
