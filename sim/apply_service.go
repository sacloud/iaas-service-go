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

package sim

import (
	"context"

	"github.com/sacloud/iaas-api-go"
	builder2 "github.com/sacloud/iaas-service-go/sim/builder"
)

func (s *Service) Apply(req *ApplyRequest) (*iaas.SIM, error) {
	return s.ApplyWithContext(context.Background(), req)
}

func (s *Service) ApplyWithContext(ctx context.Context, req *ApplyRequest) (*iaas.SIM, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	builder := &builder2.Builder{
		Name:        req.Name,
		Description: req.Description,
		Tags:        req.Tags,
		IconID:      req.IconID,
		ICCID:       req.ICCID,
		PassCode:    req.PassCode,
		Activate:    req.Activate,
		IMEI:        req.IMEI,
		Carrier:     req.Carriers,
		Client:      builder2.NewAPIClient(s.caller),
	}
	if err := builder.Validate(ctx); err != nil {
		return nil, err
	}

	if req.ID.IsEmpty() {
		return builder.Build(ctx)
	}
	return builder.Update(ctx, req.ID)
}
