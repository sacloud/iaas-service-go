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

package archive

import (
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/ostype"
	"github.com/sacloud/iaas-api-go/search"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/iaas-service-go/serviceutil"
	"github.com/sacloud/packages-go/objutil"
	"github.com/sacloud/packages-go/validate"
)

type FindRequest struct {
	Zone string `service:"-" validate:"required"`

	// OSType OS種別、NamesやTagsを指定した場合はそちらが優先される
	OSType ostype.ArchiveOSType `service:"-"`

	Names []string     `service:"-"`
	Tags  []string     `service:"-"`
	Scope types.EScope `service:"-"`

	Sort  search.SortKeys
	Count int
	From  int
}

func (req *FindRequest) Validate() error {
	return validate.New().Struct(req)
}

func (req *FindRequest) ToRequestParameter() (*iaas.FindCondition, error) {
	condition := &iaas.FindCondition{
		Filter: map[search.FilterKey]interface{}{},
	}
	if err := serviceutil.RequestConvertTo(req, condition); err != nil {
		return nil, err
	}

	filter, ok := ostype.ArchiveCriteria[req.OSType]
	if ok {
		for k, v := range filter {
			condition.Filter[k] = v
		}
	}
	if !objutil.IsEmpty(req.Names) {
		condition.Filter[search.Key("Name")] = search.AndEqual(req.Names...)
	}
	if !objutil.IsEmpty(req.Tags) {
		condition.Filter[search.Key("Tags.Name")] = search.TagsAndEqual(req.Tags...)
	}
	if !objutil.IsEmpty(req.Scope) {
		condition.Filter[search.Key("Scope")] = search.OrEqual(req.Scope)
	}
	return condition, nil
}
