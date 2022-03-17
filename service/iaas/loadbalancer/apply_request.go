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

package loadbalancer

import (
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/sacloud-go/service/iaas/loadbalancer/builder"
	"github.com/sacloud/sacloud-go/service/iaas/serviceutil"
	"github.com/sacloud/sacloud-go/service/validate"
)

type ApplyRequest struct {
	ID   types.ID // for update
	Zone string   `validate:"required"`

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

	SettingsHash string // for update
	NoWait       bool
}

func (req *ApplyRequest) Validate() error {
	return validate.Struct(req)
}

func (req *ApplyRequest) Builder(caller iaas.APICaller) (*builder.Builder, error) {
	b := &builder.Builder{Client: iaas.NewLoadBalancerOp(caller)}
	if err := serviceutil.RequestConvertTo(req, b); err != nil {
		return nil, err
	}
	return b, nil
}
