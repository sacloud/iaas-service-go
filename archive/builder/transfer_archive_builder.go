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

package builder

import (
	"context"
	"fmt"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/query"
	"github.com/sacloud/iaas-api-go/types"
)

// TransferArchiveBuilder 共有アーカイブからアーカイブの作成を行う
type TransferArchiveBuilder struct {
	Name        string
	Description string
	Tags        types.Tags
	IconID      types.ID

	SourceArchiveID   types.ID
	SourceArchiveZone string

	NoWait bool
	Client *APIClient
}

// Validate 設定値の検証
func (b *TransferArchiveBuilder) Validate(ctx context.Context, zone string) error {
	requiredValues := map[string]bool{
		"Name":              b.Name == "",
		"SourceArchiveID":   b.SourceArchiveID.IsEmpty(),
		"SourceArchiveZone": b.SourceArchiveZone == "",
	}
	for key, empty := range requiredValues {
		if empty {
			return fmt.Errorf("%s is required", key)
		}
	}
	return nil
}

// Build 他ゾーンのアーカイブからアーカイブの作成を行う
func (b *TransferArchiveBuilder) Build(ctx context.Context, zone string) (*iaas.Archive, error) {
	if err := b.Validate(ctx, zone); err != nil {
		return nil, err
	}

	zoneID, err := query.ZoneIDFromName(ctx, b.Client.Zone, zone)
	if err != nil {
		return nil, err
	}

	sourceInfo, err := b.Client.Archive.Read(ctx, b.SourceArchiveZone, b.SourceArchiveID)
	if err != nil {
		return nil, err
	}

	archive, err := b.Client.Archive.Transfer(ctx, b.SourceArchiveZone, b.SourceArchiveID, zoneID,
		&iaas.ArchiveTransferRequest{
			Name:        b.Name,
			Description: b.Description,
			Tags:        b.Tags,
			IconID:      b.IconID,
			SizeMB:      sourceInfo.SizeMB,
		})
	if err != nil {
		return nil, err
	}
	if b.NoWait {
		return archive, nil
	}

	lastState, err := iaas.WaiterForReady(func() (interface{}, error) {
		return b.Client.Archive.Read(ctx, zone, archive.ID)
	}).WaitForState(ctx)

	var ret *iaas.Archive
	if lastState != nil {
		ret = lastState.(*iaas.Archive)
	}
	return ret, err
}
