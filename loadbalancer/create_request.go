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

package loadbalancer

import (
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/packages-go/validate"
)

type CreateRequest struct {
	Zone string `service:"-" validate:"required"`

	Name               string `validate:"required"`
	Description        string `validate:"min=0,max=512"`
	Tags               types.Tags
	IconID             types.ID
	SwitchID           types.ID `validate:"required"`
	PlanID             types.ID `validate:"required"`
	VRID               int
	IPAddresses        []string `validate:"required,min=1,max=2,dive,ipv4"`
	NetworkMaskLen     int      `validate:"required"`
	DefaultRoute       string   `validate:"omitempty,ipv4"`
	VirtualIPAddresses iaas.LoadBalancerVirtualIPAddresses

	NoWait bool
}

func (req *CreateRequest) Validate() error {
	return validate.New().Struct(req)
}

func (req *CreateRequest) ApplyRequest() *ApplyRequest {
	return &ApplyRequest{
		Zone:               req.Zone,
		Name:               req.Name,
		Description:        req.Description,
		Tags:               req.Tags,
		IconID:             req.IconID,
		SwitchID:           req.SwitchID,
		PlanID:             req.PlanID,
		VRID:               req.VRID,
		IPAddresses:        req.IPAddresses,
		NetworkMaskLen:     req.NetworkMaskLen,
		DefaultRoute:       req.DefaultRoute,
		VirtualIPAddresses: req.VirtualIPAddresses,
		SettingsHash:       "",
		NoWait:             req.NoWait,
	}
}
