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

package simplemonitor

import (
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/sacloud-go/service/iaas/serviceutil"
	"github.com/sacloud/sacloud-go/service/validate"
)

type UpdateRequest struct {
	ID types.ID `request:"-" validate:"required"`

	Description        *string                        `request:",omitempty" validate:"omitempty,min=1,max=512"`
	Tags               *types.Tags                    `request:",omitempty"`
	IconID             *types.ID                      `request:",omitempty"`
	MaxCheckAttempts   *int                           `request:",omitempty"`
	RetryInterval      *int                           `request:",omitempty"`
	DelayLoop          *int                           `request:",omitempty"`
	Enabled            *types.StringFlag              `request:",omitempty"`
	HealthCheck        *iaas.SimpleMonitorHealthCheck `request:",omitempty"`
	NotifyEmailEnabled *types.StringFlag              `request:",omitempty"`
	NotifyEmailHTML    *types.StringFlag              `request:",omitempty"`
	NotifySlackEnabled *types.StringFlag              `request:",omitempty"`
	SlackWebhooksURL   *string                        `request:",omitempty"`
	NotifyInterval     *int                           `request:",omitempty"`
	Timeout            *int                           `request:",omitempty"`
	SettingsHash       string
}

func (req *UpdateRequest) Validate() error {
	return validate.Struct(req)
}

func (req *UpdateRequest) ToRequestParameter(current *iaas.SimpleMonitor) (*iaas.SimpleMonitorUpdateRequest, error) {
	r := &iaas.SimpleMonitorUpdateRequest{}
	if err := serviceutil.RequestConvertTo(current, r); err != nil {
		return nil, err
	}
	if err := serviceutil.RequestConvertTo(req, r); err != nil {
		return nil, err
	}
	return r, nil
}
