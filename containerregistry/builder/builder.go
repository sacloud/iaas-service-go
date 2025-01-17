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

package builder

import (
	"context"
	"errors"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
)

// Builder コンテナレジストリのビルダー
type Builder struct {
	ID types.ID

	Name           string
	Description    string
	Tags           types.Tags
	IconID         types.ID
	AccessLevel    types.EContainerRegistryAccessLevel
	VirtualDomain  string
	SubDomainLabel string
	Users          []*User
	SettingsHash   string
	Client         iaas.ContainerRegistryAPI
}

// User represents API parameter/response structure
type User struct {
	UserName   string
	Password   string
	Permission types.EContainerRegistryPermission
}

func (b *Builder) Build(ctx context.Context) (*iaas.ContainerRegistry, error) {
	if b.ID.IsEmpty() {
		return b.create(ctx)
	}
	return b.update(ctx)
}

func (b *Builder) create(ctx context.Context) (*iaas.ContainerRegistry, error) {
	created, err := b.Client.Create(ctx, &iaas.ContainerRegistryCreateRequest{
		Name:           b.Name,
		Description:    b.Description,
		Tags:           b.Tags,
		IconID:         b.IconID,
		AccessLevel:    b.AccessLevel,
		VirtualDomain:  b.VirtualDomain,
		SubDomainLabel: b.SubDomainLabel,
	})
	if err != nil {
		return nil, err
	}

	if len(b.Users) == 0 {
		return created, nil
	}
	return created, b.reconcileUsers(ctx, created.ID)
}

func (b *Builder) update(ctx context.Context) (*iaas.ContainerRegistry, error) {
	current, err := b.Client.Read(ctx, b.ID)
	if err != nil {
		return nil, err
	}
	if current.SubDomainLabel != b.SubDomainLabel {
		return nil, errors.New("SubDomainLabel cannot be changed")
	}

	updated, err := b.Client.Update(ctx, b.ID, &iaas.ContainerRegistryUpdateRequest{
		Name:          b.Name,
		Description:   b.Description,
		Tags:          b.Tags,
		IconID:        b.IconID,
		AccessLevel:   b.AccessLevel,
		VirtualDomain: b.VirtualDomain,
		SettingsHash:  b.SettingsHash,
	})
	if err != nil {
		return nil, err
	}
	return updated, b.reconcileUsers(ctx, updated.ID)
}

func (b *Builder) reconcileUsers(ctx context.Context, id types.ID) error {
	currentUsers, err := b.Client.ListUsers(ctx, id)
	if err != nil {
		return err
	}

	if currentUsers != nil {
		// delete
		for _, username := range b.deletedUsers(currentUsers.Users) {
			if err := b.Client.DeleteUser(ctx, id, username); err != nil {
				return err
			}
		}
		// update
		for _, user := range b.updatedUsers(currentUsers.Users) {
			if err := b.Client.UpdateUser(ctx, id, user.UserName, &iaas.ContainerRegistryUserUpdateRequest{
				Password:   user.Password, // Note: パスワードが空(Update時など)もあるが、nakedでomitemptyがついてるため問題なし
				Permission: user.Permission,
			}); err != nil {
				return err
			}
		}
	}

	// create
	var users []*iaas.ContainerRegistryUser
	if currentUsers != nil {
		users = currentUsers.Users
	}
	for _, user := range b.createdUsers(users) {
		if err := b.Client.AddUser(ctx, id, &iaas.ContainerRegistryUserCreateRequest{
			UserName:   user.UserName,
			Password:   user.Password,
			Permission: user.Permission,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) deletedUsers(currentUsers []*iaas.ContainerRegistryUser) []string {
	var results []string
	for _, current := range currentUsers {
		exists := false
		for _, desired := range b.Users {
			if current.UserName == desired.UserName {
				exists = true
				break
			}
		}
		if !exists {
			results = append(results, current.UserName)
		}
	}
	return results
}

func (b *Builder) updatedUsers(currentUsers []*iaas.ContainerRegistryUser) []*User {
	var results []*User
	for _, current := range currentUsers {
		for _, desired := range b.Users {
			if current.UserName == desired.UserName {
				results = append(results, desired)
				break
			}
		}
	}
	return results
}

func (b *Builder) createdUsers(currentUsers []*iaas.ContainerRegistryUser) []*User {
	var results []*User
	for _, created := range b.Users {
		exists := false
		for _, current := range currentUsers {
			if created.UserName == current.UserName {
				exists = true
				break
			}
		}
		if !exists {
			results = append(results, created)
		}
	}
	return results
}
