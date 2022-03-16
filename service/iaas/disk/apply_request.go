// Copyright 2022 The sacloud/sacloud-go Authors
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
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/ostype"
	"github.com/sacloud/iaas-api-go/types"
	diskBuilder "github.com/sacloud/sacloud-go/service/iaas/disk/builder"
	"github.com/sacloud/sacloud-go/service/iaas/serviceutil"
	"github.com/sacloud/sacloud-go/service/validate"
)

type ApplyRequest struct {
	Zone string   `request:"-" validate:"required"`
	ID   types.ID `request:"-"` // TODO Builderを更新対応させる(変更できない値の場合はエラーにするとか)

	Name            string `validate:"required"`
	Description     string `validate:"min=0,max=512"`
	Tags            types.Tags
	IconID          types.ID
	DiskPlanID      types.ID
	Connection      types.EDiskConnection
	SourceDiskID    types.ID
	SourceArchiveID types.ID
	ServerID        types.ID
	SizeGB          int
	DistantFrom     []types.ID

	OSType        ostype.ArchiveOSType
	EditParameter *EditParameter

	NoWait bool
}

// EditParameter ディスクの修正用パラメータ
type EditParameter struct {
	HostName string
	Password string

	DisablePWAuth       bool
	EnableDHCP          bool
	ChangePartitionUUID bool

	IPAddress      string
	NetworkMaskLen int
	DefaultRoute   string

	SSHKeys   []string
	SSHKeyIDs []types.ID

	// IsSSHKeysEphemeral trueの場合、SSHキーを生成する場合に生成したSSHキーリソースをサーバ作成後に削除する
	IsSSHKeysEphemeral bool

	IsNotesEphemeral bool
	NoteContents     []string
	Notes            []*iaas.DiskEditNote
}

func (req *ApplyRequest) Validate() error {
	return validate.Struct(req)
}

func (req *ApplyRequest) Builder(caller iaas.APICaller) (diskBuilder.Builder, error) {
	var editParameter *diskBuilder.EditRequest

	if req.EditParameter != nil {
		editParameter = &diskBuilder.EditRequest{}
		if err := serviceutil.RequestConvertTo(req.EditParameter, editParameter); err != nil {
			return nil, err
		}
	}

	director := &diskBuilder.Director{
		OSType:          req.OSType,
		Name:            req.Name,
		SizeGB:          req.SizeGB,
		DistantFrom:     req.DistantFrom,
		PlanID:          req.DiskPlanID,
		Connection:      req.Connection,
		Description:     req.Description,
		Tags:            req.Tags,
		IconID:          req.IconID,
		DiskID:          req.ID,
		SourceDiskID:    req.SourceDiskID,
		SourceArchiveID: req.SourceArchiveID,
		EditParameter:   editParameter,
		NoWait:          req.NoWait,
		Client:          diskBuilder.NewBuildersAPIClient(caller),
	}
	return director.Builder(), nil
}
