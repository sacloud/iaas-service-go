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

package database

import (
	"fmt"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/packages-go/validate"
	databaseBuilder "github.com/sacloud/sacloud-go/service/iaas/database/builder"
)

type ApplyRequest struct {
	Zone string `service:"-" validate:"required"`

	ID                    types.ID `service:"-"`
	Name                  string   `validate:"required"`
	Description           string   `validate:"min=0,max=512"`
	Tags                  types.Tags
	IconID                types.ID
	PlanID                types.ID `validate:"required"`
	SwitchID              types.ID `validate:"required"`
	IPAddresses           []string `validate:"required,min=1,max=2,dive,ipv4"`
	NetworkMaskLen        int      `validate:"required,min=1,max=32"`
	DefaultRoute          string   `validate:"omitempty,ipv4"`
	Port                  int      `validate:"omitempty,min=1,max=65535"`
	SourceNetwork         []string `validate:"dive,cidrv4"`
	DatabaseType          string   `validate:"required,oneof=mariadb postgres"`
	Username              string   `validate:"required"`
	Password              string   `validate:"required"`
	EnableReplication     bool
	ReplicaUserPassword   string `validate:"required_with=EnableReplication"`
	EnableWebUI           bool
	EnableBackup          bool
	BackupWeekdays        []types.EBackupSpanWeekday `validate:"required_with=EnableBackup,max=7"`
	BackupStartTimeHour   int                        `validate:"omitempty,min=0,max=23"`
	BackupStartTimeMinute int                        `validate:"omitempty,oneof=0 15 30 45"`
	Parameters            map[string]interface{}

	NoWait bool
}

func (req *ApplyRequest) Validate() error {
	return validate.New().Struct(req)
}

func (req *ApplyRequest) Builder(caller iaas.APICaller) (*databaseBuilder.Builder, error) {
	replicaUser := ""
	replicaPassword := ""
	if req.EnableReplication {
		replicaUser = "replica"
		replicaPassword = req.ReplicaUserPassword
	}
	builder := &databaseBuilder.Builder{
		ID:   req.ID,
		Zone: req.Zone,

		PlanID:         req.PlanID,
		SwitchID:       req.SwitchID,
		IPAddresses:    req.IPAddresses,
		NetworkMaskLen: req.NetworkMaskLen,
		DefaultRoute:   req.DefaultRoute,
		Conf: &iaas.DatabaseRemarkDBConfCommon{
			DatabaseName:     types.RDBMSVersions[types.RDBMSTypeFromString(req.DatabaseType)].Name,
			DatabaseVersion:  types.RDBMSVersions[types.RDBMSTypeFromString(req.DatabaseType)].Version,
			DatabaseRevision: types.RDBMSVersions[types.RDBMSTypeFromString(req.DatabaseType)].Revision,
		},
		CommonSetting: &iaas.DatabaseSettingCommon{
			WebUI:           types.ToWebUI(req.EnableWebUI),
			ServicePort:     req.Port,
			SourceNetwork:   req.SourceNetwork,
			DefaultUser:     req.Username,
			UserPassword:    req.Password,
			ReplicaUser:     replicaUser,
			ReplicaPassword: replicaPassword,
		},
		Name:        req.Name,
		Description: req.Description,
		Tags:        req.Tags,
		IconID:      req.IconID,
		Parameters:  req.Parameters,
		NoWait:      req.NoWait,
		Client:      databaseBuilder.NewAPIClient(caller),
	}
	if req.EnableBackup {
		builder.BackupSetting = &iaas.DatabaseSettingBackup{
			Time:      fmt.Sprintf("%02d:%02d", req.BackupStartTimeHour, req.BackupStartTimeMinute),
			DayOfWeek: req.BackupWeekdays,
		}
	}
	if req.EnableReplication {
		builder.ReplicationSetting = &iaas.DatabaseReplicationSetting{
			Model: types.DatabaseReplicationModels.MasterSlave,
		}
	}
	return builder, nil
}
