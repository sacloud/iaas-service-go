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
	"github.com/sacloud/iaas-api-go/ostype"
	"github.com/sacloud/iaas-api-go/types"
)

// Director パラメータに応じて適切なDiskBuilderを構築する
type Director struct {
	OSType ostype.ArchiveOSType

	Name                string
	SizeGB              int
	DistantFrom         []types.ID
	PlanID              types.ID
	Connection          types.EDiskConnection
	EncryptionAlgorithm types.EDiskEncryptionAlgorithm
	Description         string
	Tags                types.Tags
	IconID              types.ID

	DiskID          types.ID
	SourceDiskID    types.ID
	SourceArchiveID types.ID

	EditParameter *EditRequest

	NoWait bool
	Client *APIClient
}

// Builder パラメータに応じて適切なDiskBuilderを返す
func (d *Director) Builder() Builder {
	switch {
	case d.OSType == ostype.Custom:
		switch {
		case !d.DiskID.IsEmpty():
			return &ConnectedDiskBuilder{
				ID:                  d.DiskID,
				Name:                d.Name,
				Description:         d.Description,
				Tags:                d.Tags,
				IconID:              d.IconID,
				Connection:          d.Connection,
				EncryptionAlgorithm: d.EncryptionAlgorithm,
				EditParameter:       d.EditParameter.ToUnixDiskEditRequest(),
				NoWait:              d.NoWait,
				Client:              d.Client,
			}
		case !d.SourceDiskID.IsEmpty(), !d.SourceArchiveID.IsEmpty():
			return &FromDiskOrArchiveBuilder{
				SourceDiskID:        d.SourceDiskID,
				SourceArchiveID:     d.SourceArchiveID,
				Name:                d.Name,
				SizeGB:              d.SizeGB,
				DistantFrom:         d.DistantFrom,
				PlanID:              d.PlanID,
				Connection:          d.Connection,
				EncryptionAlgorithm: d.EncryptionAlgorithm,
				Description:         d.Description,
				Tags:                d.Tags,
				IconID:              d.IconID,
				EditParameter:       d.EditParameter.ToUnixDiskEditRequest(),
				NoWait:              d.NoWait,
				Client:              d.Client,
			}
		default:
			return &BlankBuilder{
				Name:                d.Name,
				SizeGB:              d.SizeGB,
				DistantFrom:         d.DistantFrom,
				PlanID:              d.PlanID,
				Connection:          d.Connection,
				EncryptionAlgorithm: d.EncryptionAlgorithm,
				Description:         d.Description,
				Tags:                d.Tags,
				IconID:              d.IconID,
				NoWait:              d.NoWait,
				Client:              d.Client,
			}
		}
	case d.OSType.IsSupportDiskEdit():
		return &FromUnixBuilder{
			OSType:              d.OSType,
			Name:                d.Name,
			SizeGB:              d.SizeGB,
			DistantFrom:         d.DistantFrom,
			PlanID:              d.PlanID,
			Connection:          d.Connection,
			EncryptionAlgorithm: d.EncryptionAlgorithm,
			Description:         d.Description,
			Tags:                d.Tags,
			IconID:              d.IconID,
			EditParameter:       d.EditParameter.ToUnixDiskEditRequest(),
			NoWait:              d.NoWait,
			Client:              d.Client,
		}
	default:
		// 現在はOSTypeにディスクの修正不可のアーカイブはないためここには到達しない
		return &FromFixedArchiveBuilder{
			OSType:              d.OSType,
			Name:                d.Name,
			SizeGB:              d.SizeGB,
			DistantFrom:         d.DistantFrom,
			PlanID:              d.PlanID,
			Connection:          d.Connection,
			EncryptionAlgorithm: d.EncryptionAlgorithm,
			Description:         d.Description,
			Tags:                d.Tags,
			IconID:              d.IconID,
			NoWait:              d.NoWait,
			Client:              d.Client,
		}
	}
}
