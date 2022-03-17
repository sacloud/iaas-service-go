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

package internet

import (
	"context"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	internetBuilder "github.com/sacloud/sacloud-go/service/iaas/internet/builder"
	"github.com/sacloud/sacloud-go/service/iaas/serviceutil"
	"github.com/sacloud/sacloud-go/service/validate"
)

type UpdateRequest struct {
	Zone string   `request:"-" validate:"required"`
	ID   types.ID `request:"-" validate:"required"`

	Name          *string     `request:",omitempty" validate:"omitempty,min=1"`
	Description   *string     `request:",omitempty" validate:"omitempty,min=1,max=512"`
	Tags          *types.Tags `request:",omitempty"`
	IconID        *types.ID   `request:",omitempty"`
	BandWidthMbps *int        `request:",omitempty"`
	EnableIPv6    *bool       `request:",omitempty"`
}

func (req *UpdateRequest) Validate() error {
	return validate.Struct(req)
}

func (req *UpdateRequest) Builder(ctx context.Context, caller iaas.APICaller) (*internetBuilder.Builder, error) {
	current, err := iaas.NewInternetOp(caller).Read(ctx, req.Zone, req.ID)
	if err != nil {
		return nil, err
	}

	builder := &internetBuilder.Builder{
		Name:           current.Name,
		Description:    current.Description,
		Tags:           current.Tags,
		IconID:         current.IconID,
		NetworkMaskLen: current.NetworkMaskLen,
		BandWidthMbps:  current.BandWidthMbps,
		EnableIPv6:     len(current.Switch.IPv6Nets) > 0,
		Client:         internetBuilder.NewAPIClient(caller),
	}

	if err := serviceutil.RequestConvertTo(req, builder); err != nil {
		return nil, err
	}
	return builder, nil
}
