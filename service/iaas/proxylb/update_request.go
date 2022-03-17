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

package proxylb

import (
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/sacloud-go/service/iaas/serviceutil"
	"github.com/sacloud/sacloud-go/service/validate"
)

type UpdateRequest struct {
	ID types.ID `request:"-" validate:"required"`

	Name          *string                    `request:",omitempty" validate:"omitempty,min=1"`
	Description   *string                    `request:",omitempty" validate:"omitempty,min=1,max=512"`
	Tags          *types.Tags                `request:",omitempty"`
	IconID        *types.ID                  `request:",omitempty"`
	Plan          *types.EProxyLBPlan        `request:",omitempty" validate:"omitempty,oneof=100 500 1000 5000 10000 50000 100000 400000"`
	HealthCheck   *iaas.ProxyLBHealthCheck   `request:",omitempty"`
	SorryServer   *iaas.ProxyLBSorryServer   `request:",omitempty"`
	BindPorts     *[]*iaas.ProxyLBBindPort   `request:",omitempty"`
	Servers       *[]*iaas.ProxyLBServer     `request:",omitempty"`
	Rules         *[]*iaas.ProxyLBRule       `request:",omitempty"`
	LetsEncrypt   *iaas.ProxyLBACMESetting   `request:",omitempty"`
	StickySession *iaas.ProxyLBStickySession `request:",omitempty"`
	Gzip          *iaas.ProxyLBGzip          `request:",omitempty"`
	ProxyProtocol *iaas.ProxyLBProxyProtocol `request:",omitempty"`
	Syslog        *iaas.ProxyLBSyslog        `request:",omitempty"`
	Timeout       *iaas.ProxyLBTimeout       `request:",omitempty"`
	SettingsHash  string
}

func (req *UpdateRequest) Validate() error {
	return validate.Struct(req)
}

func (req *UpdateRequest) ToRequestParameter(current *iaas.ProxyLB) (*iaas.ProxyLBUpdateRequest, error) {
	r := &iaas.ProxyLBUpdateRequest{}
	if err := serviceutil.RequestConvertTo(current, r); err != nil {
		return nil, err
	}
	if err := serviceutil.RequestConvertTo(req, r); err != nil {
		return nil, err
	}
	return r, nil
}
