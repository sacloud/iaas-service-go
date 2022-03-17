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

package packetfilter

import (
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/sacloud-go/service/iaas/serviceutil"
	"github.com/sacloud/sacloud-go/service/validate"
)

type UpdateRequest struct {
	Zone string   `service:"-" validate:"required"`
	ID   types.ID `service:"-" validate:"required"`

	Name                   *string                         `service:",omitempty" validate:"omitempty,min=1"`
	Description            *string                         `service:",omitempty" validate:"omitempty,min=1,max=512"`
	Expression             *[]*iaas.PacketFilterExpression `service:",omitempty"`
	OriginalExpressionHash string
}

func (req *UpdateRequest) Validate() error {
	return validate.Struct(req)
}

func (req *UpdateRequest) ToRequestParameter(current *iaas.PacketFilter) (*iaas.PacketFilterUpdateRequest, error) {
	r := &iaas.PacketFilterUpdateRequest{}
	if err := serviceutil.RequestConvertTo(current, r); err != nil {
		return nil, err
	}
	if err := serviceutil.RequestConvertTo(req, r); err != nil {
		return nil, err
	}
	return r, nil
}
