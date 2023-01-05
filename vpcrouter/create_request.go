// Copyright 2022-2023 The sacloud/iaas-service-go Authors
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
	"github.com/sacloud/iaas-api-go/types"
	vpcRouterBuilder "github.com/sacloud/iaas-service-go/vpcrouter/builder"
	"github.com/sacloud/packages-go/validate"
)

type CreateRequest struct {
	Zone string `service:"-" validate:"required"`

	Name        string `validate:"required"`
	Description string `validate:"min=0,max=512"`
	Tags        types.Tags
	IconID      types.ID

	PlanID                types.ID `validate:"required"`
	Version               int
	NICSetting            *vpcRouterBuilder.PremiumNICSetting
	AdditionalNICSettings []*vpcRouterBuilder.AdditionalPremiumNICSetting
	RouterSetting         *RouterSetting
	NoWait                bool
	BootAfterCreate       bool
}

func (req *CreateRequest) Validate() error {
	return validate.New().Struct(req)
}

func (req *CreateRequest) ApplyRequest() *ApplyRequest {
	var additionalNICs []vpcRouterBuilder.AdditionalNICSettingHolder
	for _, nic := range req.AdditionalNICSettings {
		additionalNICs = append(additionalNICs, nic)
	}
	return &ApplyRequest{
		Zone:                  req.Zone,
		Name:                  req.Name,
		Description:           req.Description,
		Tags:                  req.Tags,
		IconID:                req.IconID,
		PlanID:                req.PlanID,
		Version:               req.Version,
		NICSetting:            req.NICSetting,
		AdditionalNICSettings: additionalNICs,
		RouterSetting:         req.RouterSetting,
		NoWait:                req.NoWait,
		BootAfterCreate:       req.BootAfterCreate,
	}
}
