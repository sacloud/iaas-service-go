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

package serverplan

import "github.com/sacloud/iaas-api-go"

// Service provides a high-level API of for ServerPlan
type Service struct {
	caller iaas.APICaller
}

// New returns new service instance of ServerPlan
func New(caller iaas.APICaller) *Service {
	return &Service{caller: caller}
}
