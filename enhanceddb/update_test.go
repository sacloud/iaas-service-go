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

package enhanceddb

import (
	"context"
	"testing"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/testutil"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/packages-go/pointer"
	sacloudTestUtil "github.com/sacloud/packages-go/testutil"
	"github.com/stretchr/testify/require"
)

func TestEnhancedDBService_convertUpdateRequest(t *testing.T) {
	caller := testutil.SingletonAPICaller()
	name := testutil.ResourceName("container-registry-service")
	dbName := sacloudTestUtil.Random(10, sacloudTestUtil.CharSetAlpha)
	password := sacloudTestUtil.Random(16, sacloudTestUtil.CharSetAlpha)

	// setup
	svc := New(caller)
	current, err := svc.Create(&CreateRequest{
		Name:         name,
		Description:  "desc",
		Tags:         types.Tags{"tag1", "tag2"},
		DatabaseName: dbName,
		Password:     password,
	})
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		iaas.NewEnhancedDBOp(caller).Delete(context.Background(), current.ID) //nolint
	}()

	// test
	cases := []struct {
		in     *UpdateRequest
		expect *ApplyRequest
	}{
		{
			in: &UpdateRequest{
				ID:           current.ID,
				Name:         pointer.NewString(current.Name + "-upd"),
				Password:     password,
				SettingsHash: "aaaaa",
			},
			expect: &ApplyRequest{
				ID:           current.ID,
				Name:         current.Name + "-upd",
				Description:  current.Description,
				Tags:         current.Tags,
				IconID:       current.IconID,
				DatabaseName: dbName,
				Password:     password,
				SettingsHash: "aaaaa",
			},
		},
	}

	for _, tc := range cases {
		req, err := tc.in.ApplyRequest(context.Background(), caller)
		require.NoError(t, err)
		require.EqualValues(t, tc.expect, req)
	}
}
