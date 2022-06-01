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
	"github.com/sacloud/packages-go/validate"
)

type CreateRequest struct {
	Name        string `validate:"required"`
	Description string `validate:"min=0,max=512"`
	Tags        types.Tags
	IconID      types.ID

	Zones        []string `validate:"required"`
	Config       string   `validate:"required"`
	ServerPrefix string   `validate:"required"`
	Up           int      `validate:"required"`
	Down         int      `validate:"required"`
	APIKeyID     string   `validate:"required"`
}

func (req *CreateRequest) Validate() error {
	return validate.New().Struct(req)
}

func (req *CreateRequest) ToRequestParameter() (*iaas.AutoScaleCreateRequest, error) {
	return &iaas.AutoScaleCreateRequest{
		Name:         req.Name,
		Description:  req.Description,
		Tags:         req.Tags,
		IconID:       req.IconID,
		Zones:        req.Zones,
		Config:       req.Config,
		ServerPrefix: req.ServerPrefix,
		Up:           req.Up,
		Down:         req.Down,
		APIKeyID:     req.APIKeyID,
	}, nil
}
