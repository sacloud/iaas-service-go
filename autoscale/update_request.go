// Copyright 2022 The sacloud/iaas-service-go Authors
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

package autoscale

import (
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/iaas-service-go/serviceutil"
	"github.com/sacloud/packages-go/validate"
)

type UpdateRequest struct {
	ID types.ID `service:"-" validate:"required"`

	Name        *string     `service:",omitempty" validate:"omitempty,min=1"`
	Description *string     `service:",omitempty" validate:"omitempty,min=1,max=512"`
	Tags        *types.Tags `service:",omitempty"`
	IconID      *types.ID   `service:",omitempty"`

	Zones               *[]string                 `service:",omitempty" validate:"omitempty,required"`
	Config              *string                   `service:",omitempty" validate:"omitempty,required"`
	CPUThresholdScaling UpdateCPUThresholdScaling `service:"-" validate:"dive"`

	SettingsHash string
}

type UpdateCPUThresholdScaling struct {
	ServerPrefix *string `service:",omitempty" validate:"omitempty,required"`
	Up           *int    `service:",omitempty" validate:"omitempty,required"`
	Down         *int    `service:",omitempty" validate:"omitempty,required"`
}

func (req *UpdateRequest) Validate() error {
	return validate.New().Struct(req)
}

func (req *UpdateRequest) ToRequestParameter(current *iaas.AutoScale) (*iaas.AutoScaleUpdateRequest, error) {
	r := &iaas.AutoScaleUpdateRequest{}
	if err := serviceutil.RequestConvertTo(current, r); err != nil {
		return nil, err
	}
	if err := serviceutil.RequestConvertTo(req, r); err != nil {
		return nil, err
	}
	if req.CPUThresholdScaling.ServerPrefix != nil {
		r.CPUThresholdScaling.ServerPrefix = *req.CPUThresholdScaling.ServerPrefix
	}
	if req.CPUThresholdScaling.Up != nil {
		r.CPUThresholdScaling.Up = *req.CPUThresholdScaling.Up
	}
	if req.CPUThresholdScaling.Down != nil {
		r.CPUThresholdScaling.Down = *req.CPUThresholdScaling.Down
	}

	return r, nil
}
