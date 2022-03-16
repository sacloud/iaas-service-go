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

package disk

import (
	"context"
	"fmt"
	"time"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
)

// EditRequest 汎用ディスクの修正リクエストパラメータ DiskDirectorが利用する
type EditRequest UnixEditRequest

// ToUnixDiskEditRequest Unix系パラメータへの変換
func (d *EditRequest) ToUnixDiskEditRequest() *UnixEditRequest {
	if d == nil {
		return nil
	}
	req := UnixEditRequest(*d)
	return &req
}

// ToWindowsDiskEditRequest Windows系パラメータへの変換
func (d *EditRequest) ToWindowsDiskEditRequest() *WindowsEditRequest {
	if d == nil {
		return nil
	}
	return &WindowsEditRequest{
		IPAddress:      d.IPAddress,
		NetworkMaskLen: d.NetworkMaskLen,
		DefaultRoute:   d.DefaultRoute,
	}
}

// UnixEditRequest Unix系の場合のディスクの修正リクエスト
type UnixEditRequest struct {
	HostName string
	Password string

	DisablePWAuth       bool
	EnableDHCP          bool
	ChangePartitionUUID bool

	IPAddress      string
	NetworkMaskLen int
	DefaultRoute   string

	SSHKeys   []string
	SSHKeyIDs []types.ID

	// IsSSHKeysEphemeral trueの場合、SSHキーを生成する場合に生成したSSHキーリソースをサーバ作成後に削除する
	IsSSHKeysEphemeral bool
	// GenerateSSHKeyName 設定されていた場合、クラウドAPIを用いてキーペアを生成する。
	GenerateSSHKeyName        string
	GenerateSSHKeyPassPhrase  string
	GenerateSSHKeyDescription string

	IsNotesEphemeral bool
	NoteContents     []string
	Notes            []*iaas.DiskEditNote
}

// Validate 設定値の検証
func (u *UnixEditRequest) Validate(ctx context.Context, client *APIClient) error {
	for _, id := range u.SSHKeyIDs {
		if _, err := client.SSHKey.Read(ctx, id); err != nil {
			return err
		}
	}
	for _, note := range u.Notes {
		if _, err := client.Note.Read(ctx, note.ID); err != nil {
			return err
		}
	}
	return nil
}

func (u *UnixEditRequest) prepareDiskEditParameter(ctx context.Context, client *APIClient) (*iaas.DiskEditRequest, *iaas.SSHKeyGenerated, []*iaas.Note, error) {
	editReq := &iaas.DiskEditRequest{
		Background:          true,
		Password:            u.Password,
		DisablePWAuth:       u.DisablePWAuth,
		EnableDHCP:          u.EnableDHCP,
		ChangePartitionUUID: u.ChangePartitionUUID,
		HostName:            u.HostName,
	}

	if u.IPAddress != "" {
		editReq.UserIPAddress = u.IPAddress
	}
	if u.NetworkMaskLen > 0 || u.DefaultRoute != "" {
		editReq.UserSubnet = &iaas.DiskEditUserSubnet{
			NetworkMaskLen: u.NetworkMaskLen,
			DefaultRoute:   u.DefaultRoute,
		}
	}

	// ssh key
	var sshKeys []*iaas.DiskEditSSHKey
	for _, key := range u.SSHKeys {
		sshKeys = append(sshKeys, &iaas.DiskEditSSHKey{
			PublicKey: key,
		})
	}
	for _, id := range u.SSHKeyIDs {
		sshKeys = append(sshKeys, &iaas.DiskEditSSHKey{
			ID: id,
		})
	}

	var generatedSSHKey *iaas.SSHKeyGenerated
	if u.GenerateSSHKeyName != "" {
		generated, err := client.SSHKey.Generate(ctx, &iaas.SSHKeyGenerateRequest{
			Name:        u.GenerateSSHKeyName,
			PassPhrase:  u.GenerateSSHKeyPassPhrase,
			Description: u.GenerateSSHKeyDescription,
		})
		if err != nil {
			return nil, nil, nil, err
		}
		generatedSSHKey = generated
		sshKeys = append(sshKeys, &iaas.DiskEditSSHKey{
			ID: generated.ID,
		})
	}
	editReq.SSHKeys = sshKeys

	// startup script
	var notes []*iaas.DiskEditNote
	var generatedNotes []*iaas.Note

	for _, note := range u.NoteContents {
		created, err := client.Note.Create(ctx, &iaas.NoteCreateRequest{
			Name:    fmt.Sprintf("note-%s", time.Now().Format(time.RFC3339)),
			Class:   "shell",
			Content: note,
		})
		if err != nil {
			return nil, nil, nil, err
		}
		notes = append(notes, &iaas.DiskEditNote{
			ID: created.ID,
		})
		generatedNotes = append(generatedNotes, created)
	}
	for _, note := range u.Notes {
		notes = append(notes, &iaas.DiskEditNote{
			ID:        note.ID,
			APIKeyID:  note.APIKeyID,
			Variables: note.Variables,
		})
	}
	editReq.Notes = notes

	return editReq, generatedSSHKey, generatedNotes, nil
}

// WindowsEditRequest Windows系の場合のディスクの修正リクエスト
type WindowsEditRequest struct {
	IPAddress      string
	NetworkMaskLen int
	DefaultRoute   string
}

func (w *WindowsEditRequest) prepareDiskEditParameter() *iaas.DiskEditRequest {
	editReq := &iaas.DiskEditRequest{Background: true}

	if w.IPAddress != "" {
		editReq.UserIPAddress = w.IPAddress
	}
	if w.NetworkMaskLen > 0 || w.DefaultRoute != "" {
		editReq.UserSubnet = &iaas.DiskEditUserSubnet{
			NetworkMaskLen: w.NetworkMaskLen,
			DefaultRoute:   w.DefaultRoute,
		}
	}
	return editReq
}
