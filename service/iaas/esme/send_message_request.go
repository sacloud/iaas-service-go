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

package esme

import (
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/sacloud-go/service/validate"
)

type SendMessageRequest struct {
	ID types.ID `request:"-" validate:"required"`

	Destination string `validate:"required"`
	Sender      string `validate:"required"`
	OTP         string
	DomainName  string
}

func (req *SendMessageRequest) Validate() error {
	return validate.Struct(req)
}

func (req *SendMessageRequest) ToRequestParameter() interface{} {
	if req.OTP == "" {
		return &iaas.ESMESendMessageWithGeneratedOTPRequest{
			Destination: req.Destination,
			Sender:      req.Sender,
			DomainName:  req.DomainName,
		}
	}
	return &iaas.ESMESendMessageWithInputtedOTPRequest{
		Destination: req.Destination,
		Sender:      req.Sender,
		DomainName:  req.DomainName,
		OTP:         req.OTP,
	}
}
