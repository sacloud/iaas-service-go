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

package server

import (
	"context"

	"github.com/sacloud/iaas-api-go"
	serverBuilder "github.com/sacloud/iaas-service-go/server/builder"
)

func (s *Service) Apply(req *ApplyRequest) (*iaas.Server, error) {
	return s.ApplyWithContext(context.Background(), req)
}

func (s *Service) ApplyWithContext(ctx context.Context, req *ApplyRequest) (*iaas.Server, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	builder, err := req.Builder(s.caller)
	if err != nil {
		return nil, err
	}

	var result *serverBuilder.BuildResult

	if req.ID.IsEmpty() {
		created, err := builder.Build(ctx, req.Zone)
		if err != nil {
			return nil, err
		}
		result = created
	} else {
		updated, err := builder.Update(ctx, req.Zone)
		if err != nil {
			return nil, err
		}
		result = updated
	}

	serverOp := iaas.NewServerOp(s.caller)
	server, err := serverOp.Read(ctx, req.Zone, result.ServerID)
	if err != nil {
		return nil, err
	}
	return server, nil
}
