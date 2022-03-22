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

package localrouter

import (
	"context"

	"github.com/sacloud/iaas-api-go"
)

func (s *Service) Find(req *FindRequest) ([]*iaas.LocalRouter, error) {
	return s.FindWithContext(context.Background(), req)
}

func (s *Service) FindWithContext(ctx context.Context, req *FindRequest) ([]*iaas.LocalRouter, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	params, err := req.ToRequestParameter()
	if err != nil {
		return nil, err
	}

	client := iaas.NewLocalRouterOp(s.caller)
	found, err := client.Find(ctx, params)
	if err != nil {
		return nil, err
	}
	return found.LocalRouters, nil
}