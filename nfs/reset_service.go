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

package nfs

import (
	"context"

	"github.com/sacloud/iaas-api-go"
)

func (s *Service) Reset(req *ResetRequest) error {
	return s.ResetWithContext(context.Background(), req)
}

func (s *Service) ResetWithContext(ctx context.Context, req *ResetRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	client := iaas.NewNFSOp(s.caller)
	return client.Reset(ctx, req.Zone, req.ID)
}
