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
	"errors"
	"fmt"
	"reflect"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/plans"
	"github.com/sacloud/iaas-api-go/helper/power"
	"github.com/sacloud/iaas-api-go/helper/query"
	"github.com/sacloud/iaas-api-go/types"
	service "github.com/sacloud/iaas-service-go"
	disk "github.com/sacloud/iaas-service-go/disk/builder"
	"github.com/sacloud/packages-go/size"
)

// Builder サーバ作成時のパラメータ
type Builder struct {
	Name            string
	CPU             int
	MemoryGB        int
	GPU             int
	CPUModel        string
	Commitment      types.ECommitment
	Generation      types.EPlanGeneration
	InterfaceDriver types.EInterfaceDriver
	Description     string
	IconID          types.ID
	Tags            types.Tags
	BootAfterCreate bool
	CDROMID         types.ID
	PrivateHostID   types.ID
	NIC             NICSettingHolder
	AdditionalNICs  []AdditionalNICSettingHolder
	DiskBuilders    []disk.Builder

	UserData string

	Client *APIClient

	NoWait bool

	ServerID      types.ID
	ForceShutdown bool
}

func BuilderFromResource(ctx context.Context, caller iaas.APICaller, zone string, id types.ID) (*Builder, error) {
	serverOp := iaas.NewServerOp(caller)
	current, err := serverOp.Read(ctx, zone, id)
	if err != nil {
		return nil, err
	}

	var nic NICSettingHolder
	if len(current.Interfaces) > 0 {
		iface := current.Interfaces[0]
		switch {
		case iface.SwitchID.IsEmpty():
			nic = &DisconnectedNICSetting{}
		case iface.SwitchScope == types.Scopes.Shared:
			nic = &SharedNICSetting{PacketFilterID: iface.PacketFilterID}
		default:
			nic = &ConnectedNICSetting{
				SwitchID:         iface.SwitchID,
				DisplayIPAddress: iface.UserIPAddress,
				PacketFilterID:   iface.PacketFilterID,
			}
		}
	}

	var additionalNICs []AdditionalNICSettingHolder
	if len(current.Interfaces) > 1 {
		for i, iface := range current.Interfaces {
			if i == 0 {
				continue
			}
			switch {
			case iface.SwitchID.IsEmpty():
				additionalNICs = append(additionalNICs, &DisconnectedNICSetting{})
			default:
				additionalNICs = append(additionalNICs, &ConnectedNICSetting{
					SwitchID:         iface.SwitchID,
					DisplayIPAddress: iface.UserIPAddress,
					PacketFilterID:   iface.PacketFilterID,
				})
			}
		}
	}

	diskOp := iaas.NewDiskOp(caller)
	var diskBuilders []disk.Builder
	for _, d := range current.Disks {
		currentDisk, err := diskOp.Read(ctx, zone, d.ID)
		if err != nil {
			return nil, err
		}
		diskBuilders = append(diskBuilders, &disk.ConnectedDiskBuilder{
			ID:            currentDisk.ID,
			EditParameter: nil,
			Name:          currentDisk.Name,
			Description:   currentDisk.Description,
			Tags:          currentDisk.Tags,
			IconID:        currentDisk.IconID,
			Connection:    currentDisk.Connection,
			NoWait:        false,
			Client:        disk.NewBuildersAPIClient(caller),
		})
	}

	return &Builder{
		Name:            current.Name,
		CPU:             current.CPU,
		MemoryGB:        current.MemoryMB * size.GiB,
		GPU:             current.GPU,
		CPUModel:        current.ServerPlanCPUModel,
		Commitment:      current.ServerPlanCommitment,
		Generation:      current.ServerPlanGeneration,
		InterfaceDriver: current.InterfaceDriver,
		Description:     current.Description,
		IconID:          current.IconID,
		Tags:            current.Tags,
		BootAfterCreate: false,
		CDROMID:         current.CDROMID,
		PrivateHostID:   current.PrivateHostID,
		NIC:             nic,
		AdditionalNICs:  additionalNICs,
		DiskBuilders:    diskBuilders,
		Client:          NewBuildersAPIClient(caller),
		ServerID:        current.ID,
	}, nil
}

// BuildResult サーバ構築結果
type BuildResult struct {
	ServerID types.ID
	DiskIDs  []types.ID
}

var (
	defaultCPU             = 1
	defaultMemoryGB        = 1
	defaultGPU             = 0
	defaultCPUModel        = ""
	defaultCommitment      = types.Commitments.Standard
	defaultGeneration      = types.PlanGenerations.Default
	defaultInterfaceDriver = types.InterfaceDrivers.VirtIO
)

// Validate 入力値の検証
//
// 各種IDの存在確認のためにAPIリクエストが行われます。
func (b *Builder) Validate(ctx context.Context, zone string) error {
	b.setDefaults()

	// Fields
	if b.Client == nil {
		return errors.New("client is empty")
	}

	if b.NIC == nil && len(b.AdditionalNICs) > 0 {
		return errors.New("NIC is required when AdditionalNICs is specified")
	}

	if len(b.AdditionalNICs) > 9 {
		return errors.New("AdditionalNICs must be less than 9")
	}

	if b.InterfaceDriver != types.InterfaceDrivers.E1000 && b.InterfaceDriver != types.InterfaceDrivers.VirtIO {
		return fmt.Errorf("invalid InterfaceDriver: %s", b.InterfaceDriver)
	}

	// NICs
	if b.NIC != nil {
		if err := b.NIC.Validate(ctx, b.Client, zone); err != nil {
			return fmt.Errorf("invalid NIC: %s", err)
		}
	}
	for i, nic := range b.AdditionalNICs {
		if err := nic.Validate(ctx, b.Client, zone); err != nil {
			return fmt.Errorf("invalid AdditionalNICs[%d]: %s", i, err)
		}
	}

	// server plan
	_, err := query.FindServerPlan(ctx, b.Client.ServerPlan, zone, &query.FindServerPlanRequest{
		CPU:        b.CPU,
		MemoryGB:   b.MemoryGB,
		GPU:        b.GPU,
		CPUModel:   b.CPUModel,
		Commitment: b.Commitment,
		Generation: b.Generation,
	})
	if err != nil {
		return err
	}

	for _, diskBuilder := range b.DiskBuilders {
		if err := diskBuilder.Validate(ctx, zone); err != nil {
			return err
		}
		if b.NoWait && !diskBuilder.NoWaitFlag() {
			return errors.New("NoWait=true is not supported if the disks contain NoWait=false")
		}
	}

	if b.NoWait && b.BootAfterCreate {
		return errors.New("NoWait=true is not supported with BootAfterCreate=true")
	}

	return nil
}

// Build サーバ構築を行う
func (b *Builder) Build(ctx context.Context, zone string) (*BuildResult, error) {
	// validate
	if err := b.Validate(ctx, zone); err != nil {
		return nil, err
	}

	// create server
	server, err := b.createServer(ctx, zone)
	if err != nil {
		return nil, err
	}
	result := &BuildResult{
		ServerID: server.ID,
	}

	// create&connect disk(s)
	for _, diskReq := range b.DiskBuilders {
		builtDisk, err := diskReq.Build(ctx, zone, server.ID)
		if err != nil {
			return result, err
		}
		result.DiskIDs = append(result.DiskIDs, builtDisk.DiskID)
	}

	// connect packet filter
	if err := b.updateInterfaces(ctx, zone, server); err != nil {
		return result, err
	}

	// insert CD-ROM
	if !b.CDROMID.IsEmpty() {
		req := &iaas.InsertCDROMRequest{ID: b.CDROMID}
		if err := b.Client.Server.InsertCDROM(ctx, zone, server.ID, req); err != nil {
			return result, err
		}
	}

	// bool
	if !b.NoWait && b.BootAfterCreate {
		if err := power.BootServer(ctx, b.Client.Server, zone, server.ID, b.userData()...); err != nil {
			return result, err
		}
	}

	b.ServerID = result.ServerID
	return result, nil
}

// IsNeedShutdown Update時にシャットダウンが必要か
func (b *Builder) IsNeedShutdown(ctx context.Context, zone string) (bool, error) {
	if b.ServerID.IsEmpty() {
		return false, fmt.Errorf("server id required")
	}

	server, err := b.Client.Server.Read(ctx, zone, b.ServerID)
	if err != nil {
		return false, err
	}

	if b.UserData != "" {
		return true, nil
	}

	current := b.currentState(server)
	desired := b.desiredState()

	// シャットダウンが不要な項目には固定値を入れる
	var nics []*nicState
	nics = append(nics, current.nic)
	nics = append(nics, current.additionalNICs...)
	nics = append(nics, desired.nic)
	nics = append(nics, desired.additionalNICs...)
	b.fillDummyValueToState(nics...)

	if !reflect.DeepEqual(current, desired) {
		return true, nil
	}

	// ここに到達するときはserver.Disksとb.DiskBuildersは同数となっている
	for i, disk := range server.Disks {
		level := b.DiskBuilders[i].UpdateLevel(ctx, zone, &iaas.Disk{
			ID:              disk.ID,
			Name:            disk.Name,
			Availability:    disk.Availability,
			Connection:      disk.Connection,
			ConnectionOrder: disk.ConnectionOrder,
			ReinstallCount:  disk.ReinstallCount,
			SizeMB:          disk.SizeMB,
			DiskPlanID:      disk.DiskPlanID,
			Storage:         disk.Storage,
		})

		if level == service.UpdateLevelNeedShutdown {
			return true, nil
		}
	}
	return false, nil
}

func (b *Builder) fillDummyValueToState(state ...*nicState) {
	for _, s := range state {
		if s != nil {
			s.packetFilterID = types.ID(0)
			s.displayIP = ""
		}
	}
}

// Update サーバの更新
func (b *Builder) Update(ctx context.Context, zone string) (*BuildResult, error) {
	// validate
	if err := b.Validate(ctx, zone); err != nil {
		return nil, err
	}
	if b.ServerID.IsEmpty() {
		return nil, fmt.Errorf("server id required")
	}

	result := &BuildResult{ServerID: b.ServerID}

	server, err := b.Client.Server.Read(ctx, zone, b.ServerID)
	if err != nil {
		return result, err
	}

	isNeedShutdown, err := b.IsNeedShutdown(ctx, zone)
	if err != nil {
		return result, err
	}

	// shutdown
	running := server.InstanceStatus.IsUp()
	if isNeedShutdown && running {
		if b.NoWait {
			return nil, errors.New("NoWait option is not available due to the need to shut down")
		}
		if err := power.ShutdownServer(ctx, b.Client.Server, zone, server.ID, b.ForceShutdown); err != nil {
			return result, err
		}
	}

	// reconcile disks
	if err := b.reconcileDisks(ctx, zone, server, result); err != nil {
		return result, err
	}

	// reconcile interface
	if err := b.reconcileInterfaces(ctx, zone, server); err != nil {
		return result, err
	}

	// plan
	if b.isPlanChanged(server) {
		b.Tags = plans.AppendPreviousIDTagIfAbsent(b.Tags, server.ID)
		updated, err := b.Client.Server.ChangePlan(ctx, zone, server.ID, &iaas.ServerChangePlanRequest{
			CPU:                  b.CPU,
			MemoryMB:             b.MemoryGB * size.GiB,
			GPU:                  b.GPU,
			ServerPlanGeneration: b.Generation,
			ServerPlanCommitment: b.Commitment,
			ServerPlanCPUModel:   b.CPUModel,
		})
		if err != nil {
			return result, err
		}
		server = updated
	}

	// update
	updated, err := b.Client.Server.Update(ctx, zone, server.ID, &iaas.ServerUpdateRequest{
		Name:            b.Name,
		Description:     b.Description,
		Tags:            b.Tags,
		IconID:          b.IconID,
		PrivateHostID:   b.PrivateHostID,
		InterfaceDriver: b.InterfaceDriver,
	})
	if err != nil {
		return result, err
	}
	server = updated
	result.ServerID = server.ID

	// insert CD-ROM
	if !b.CDROMID.IsEmpty() && b.CDROMID != server.CDROMID {
		if !server.CDROMID.IsEmpty() {
			if err := b.Client.Server.EjectCDROM(ctx, zone, server.ID, &iaas.EjectCDROMRequest{ID: server.CDROMID}); err != nil {
				return result, err
			}
		}
		if err := b.Client.Server.InsertCDROM(ctx, zone, server.ID, &iaas.InsertCDROMRequest{ID: b.CDROMID}); err != nil {
			return result, err
		}
	}

	// boot
	if isNeedShutdown && running && server.InstanceStatus.IsDown() {
		if err := power.BootServer(ctx, b.Client.Server, zone, server.ID, b.userData()...); err != nil {
			return result, err
		}
	}

	result.ServerID = server.ID
	return result, nil
}

func (b *Builder) setDefaults() {
	if b.CPU == 0 {
		b.CPU = defaultCPU
	}
	if b.MemoryGB == 0 {
		b.MemoryGB = defaultMemoryGB
	}
	if b.GPU == 0 {
		b.GPU = defaultGPU
	}
	if b.CPUModel == "" {
		b.CPUModel = defaultCPUModel
	}
	if b.Commitment == types.ECommitment("") {
		b.Commitment = defaultCommitment
	}
	if b.Generation == types.EPlanGeneration(0) {
		b.Generation = defaultGeneration
	}
	if b.InterfaceDriver == types.EInterfaceDriver("") {
		b.InterfaceDriver = defaultInterfaceDriver
	}
}

type serverState struct {
	privateHostID   types.ID
	interfaceDriver types.EInterfaceDriver
	memoryGB        int
	cpu             int
	gpu             int
	cpuModel        string
	commitment      types.ECommitment
	nic             *nicState   // hash
	additionalNICs  []*nicState // hash
	diskCount       int
}

func (b *Builder) desiredState() *serverState {
	var nic *nicState
	if b.NIC != nil {
		nic = b.NIC.state()
	}
	var additionalNICs []*nicState
	for _, n := range b.AdditionalNICs {
		additionalNICs = append(additionalNICs, n.state())
	}

	return &serverState{
		privateHostID:   b.PrivateHostID,
		interfaceDriver: b.InterfaceDriver,
		memoryGB:        b.MemoryGB,
		cpu:             b.CPU,
		gpu:             b.GPU,
		cpuModel:        b.CPUModel,
		commitment:      b.Commitment,
		nic:             nic,
		additionalNICs:  additionalNICs,
		diskCount:       len(b.DiskBuilders),
	}
}

func (b *Builder) currentNICState(nic *iaas.InterfaceView) *nicState {
	var state *nicState

	switch {
	case nic.SwitchScope == types.Scopes.Shared:
		state = &nicState{
			upstreamType:   types.UpstreamNetworkTypes.Shared,
			switchID:       types.ID(0),
			packetFilterID: nic.PacketFilterID,
			displayIP:      "",
		}
	case nic.SwitchID.IsEmpty():
		state = &nicState{
			upstreamType:   types.UpstreamNetworkTypes.None,
			switchID:       types.ID(0),
			packetFilterID: types.ID(0),
			displayIP:      "",
		}
	default:
		state = &nicState{
			upstreamType:   types.UpstreamNetworkTypes.Switch,
			switchID:       nic.SwitchID,
			packetFilterID: nic.PacketFilterID,
			displayIP:      nic.UserIPAddress,
		}
	}
	return state
}

func (b *Builder) currentState(server *iaas.Server) *serverState {
	var nic *nicState
	var additionalNICs []*nicState
	for i, n := range server.Interfaces {
		state := b.currentNICState(n)
		if i == 0 {
			nic = state
		} else {
			additionalNICs = append(additionalNICs, state)
		}
	}

	return &serverState{
		privateHostID:   server.PrivateHostID,
		interfaceDriver: server.InterfaceDriver,
		memoryGB:        server.GetMemoryGB(),
		cpu:             server.CPU,
		gpu:             server.GPU,
		cpuModel:        server.ServerPlanCPUModel,
		commitment:      server.ServerPlanCommitment,
		nic:             nic,
		additionalNICs:  additionalNICs,
		diskCount:       len(server.Disks),
	}
}

// createServer サーバ作成
func (b *Builder) createServer(ctx context.Context, zone string) (*iaas.Server, error) {
	param := &iaas.ServerCreateRequest{
		CPU:                  b.CPU,
		MemoryMB:             b.MemoryGB * size.GiB,
		GPU:                  b.GPU,
		ServerPlanCPUModel:   b.CPUModel,
		ServerPlanCommitment: b.Commitment,
		ServerPlanGeneration: b.Generation,
		InterfaceDriver:      b.InterfaceDriver,
		Name:                 b.Name,
		Description:          b.Description,
		Tags:                 b.Tags,
		IconID:               b.IconID,
		WaitDiskMigration:    false,
		PrivateHostID:        b.PrivateHostID,
		ConnectedSwitches:    []*iaas.ConnectedSwitch{},
	}
	if b.NIC != nil {
		cs := b.NIC.GetConnectedSwitchParam()
		if cs == nil {
			param.ConnectedSwitches = append(param.ConnectedSwitches, nil)
		} else {
			param.ConnectedSwitches = append(param.ConnectedSwitches, cs)
		}
	}
	if len(b.AdditionalNICs) > 0 {
		for _, nic := range b.AdditionalNICs {
			switchID := nic.GetSwitchID()
			if switchID.IsEmpty() {
				param.ConnectedSwitches = append(param.ConnectedSwitches, nil)
			} else {
				param.ConnectedSwitches = append(param.ConnectedSwitches, &iaas.ConnectedSwitch{ID: switchID})
			}
		}
	}
	return b.Client.Server.Create(ctx, zone, param)
}

type updateInterfaceRequest struct {
	index          int
	packetFilterID types.ID
	displayIP      string
}

func (b *Builder) collectInterfaceParameters() []*updateInterfaceRequest {
	var reqs []*updateInterfaceRequest
	if b.NIC != nil {
		reqs = append(reqs, &updateInterfaceRequest{
			index:          0,
			packetFilterID: b.NIC.GetPacketFilterID(),
			displayIP:      b.NIC.GetDisplayIPAddress(),
		})
	}
	for i, nic := range b.AdditionalNICs {
		reqs = append(reqs, &updateInterfaceRequest{
			index:          i + 1,
			packetFilterID: nic.GetPacketFilterID(),
			displayIP:      nic.GetDisplayIPAddress(),
		})
	}
	return reqs
}

func (b *Builder) updateInterfaces(ctx context.Context, zone string, server *iaas.Server) error {
	requests := b.collectInterfaceParameters()
	for _, req := range requests {
		if req.index < len(server.Interfaces) {
			iface := server.Interfaces[req.index]

			if !req.packetFilterID.IsEmpty() {
				if err := b.Client.Interface.ConnectToPacketFilter(ctx, zone, iface.ID, req.packetFilterID); err != nil {
					return err
				}
			}

			if req.displayIP != "" {
				if _, err := b.Client.Interface.Update(ctx, zone, iface.ID, &iaas.InterfaceUpdateRequest{
					UserIPAddress: req.displayIP,
				}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (b *Builder) reconcileDisks(ctx context.Context, zone string, server *iaas.Server, result *BuildResult) error {
	// reconcile disks
	isDiskUpdated := len(server.Disks) != len(b.DiskBuilders) // isDiskUpdateがtrueの場合、後でディスクの取外&接続を行う
	for i, diskReq := range b.DiskBuilders {
		if diskReq.DiskID().IsEmpty() {
			_, err := diskReq.Build(ctx, zone, server.ID)
			if err != nil {
				return err
			}
			isDiskUpdated = true
		}
		if len(server.Disks) > i {
			disk := server.Disks[i]
			level := diskReq.UpdateLevel(ctx, zone, &iaas.Disk{
				ID:              disk.ID,
				Name:            disk.Name,
				Availability:    disk.Availability,
				Connection:      disk.Connection,
				ConnectionOrder: disk.ConnectionOrder,
				ReinstallCount:  disk.ReinstallCount,
				SizeMB:          disk.SizeMB,
				DiskPlanID:      disk.DiskPlanID,
				Storage:         disk.Storage,
			})
			if level != service.UpdateLevelNone {
				_, err := diskReq.Update(ctx, zone)
				if err != nil {
					return err
				}
			}
			if disk.ID != diskReq.DiskID() {
				isDiskUpdated = true
			}
		}
	}
	if isDiskUpdated {
		refreshed, err := b.Client.Server.Read(ctx, zone, server.ID)
		if err != nil {
			return err
		}
		server = refreshed

		// disconnect all
		for i := range server.Disks {
			// disconnect
			if err := b.Client.Disk.DisconnectFromServer(ctx, zone, server.Disks[i].ID); err != nil {
				return err
			}
		}
		// reconnect all
		for _, diskReq := range b.DiskBuilders {
			result.DiskIDs = []types.ID{}
			if err := b.Client.Disk.ConnectToServer(ctx, zone, diskReq.DiskID(), server.ID); err != nil {
				return err
			}
			result.DiskIDs = append(result.DiskIDs, diskReq.DiskID())
		}
	}
	return nil
}

func (b *Builder) reconcileInterfaces(ctx context.Context, zone string, server *iaas.Server) error {
	desiredState := b.desiredState()
	for i, nic := range server.Interfaces {
		current := b.currentNICState(nic)
		var desired *nicState
		if i == 0 {
			desired = desiredState.nic
		} else if len(desiredState.additionalNICs) > i-1 {
			desired = desiredState.additionalNICs[i-1]
		}

		if desired == nil {
			// disconnect and delete
			if !nic.SwitchID.IsEmpty() {
				if err := b.Client.Interface.DisconnectFromSwitch(ctx, zone, nic.ID); err != nil {
					return err
				}
			}
			if err := b.Client.Interface.Delete(ctx, zone, nic.ID); err != nil {
				return err
			}
			continue
		}
		if current.upstreamType != desired.upstreamType ||
			current.switchID != desired.switchID {
			if !nic.SwitchID.IsEmpty() {
				if err := b.Client.Interface.DisconnectFromSwitch(ctx, zone, nic.ID); err != nil {
					return err
				}
			}
		}
	}

	desiredNICs := []*nicState{desiredState.nic}
	desiredNICs = append(desiredNICs, desiredState.additionalNICs...)

	for i, desired := range desiredNICs {
		if desired == nil {
			continue
		}
		var nic *iaas.InterfaceView
		if len(server.Interfaces) > i {
			nic = server.Interfaces[i]
		}
		if nic == nil {
			created, err := b.Client.Interface.Create(ctx, zone, &iaas.InterfaceCreateRequest{
				ServerID: server.ID,
			})
			if err != nil {
				return err
			}
			nic = &iaas.InterfaceView{
				ID:             created.ID,
				MACAddress:     created.MACAddress,
				IPAddress:      created.IPAddress,
				UserIPAddress:  created.UserIPAddress,
				HostName:       created.HostName,
				SwitchID:       created.SwitchID,
				SwitchScope:    created.SwitchScope,
				PacketFilterID: created.PacketFilterID,
			}
		}
		switch desired.upstreamType {
		case types.UpstreamNetworkTypes.None:
			// noop
		case types.UpstreamNetworkTypes.Shared:
			if nic.SwitchScope != types.Scopes.Shared {
				if err := b.Client.Interface.ConnectToSharedSegment(ctx, zone, nic.ID); err != nil {
					return err
				}
			}
		default:
			if nic.SwitchID != desired.switchID {
				if err := b.Client.Interface.ConnectToSwitch(ctx, zone, nic.ID, desired.switchID); err != nil {
					return err
				}
			}
		}
		if desired.packetFilterID != nic.PacketFilterID {
			if !nic.PacketFilterID.IsEmpty() {
				if err := b.Client.Interface.DisconnectFromPacketFilter(ctx, zone, nic.ID); err != nil {
					return err
				}
			}
			if !desired.packetFilterID.IsEmpty() {
				if err := b.Client.Interface.ConnectToPacketFilter(ctx, zone, nic.ID, desired.packetFilterID); err != nil {
					return err
				}
			}
		}
		if desired.displayIP != nic.UserIPAddress {
			if _, err := b.Client.Interface.Update(ctx, zone, nic.ID, &iaas.InterfaceUpdateRequest{
				UserIPAddress: desired.displayIP,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Builder) isPlanChanged(server *iaas.Server) bool {
	return b.CPU != server.CPU ||
		b.MemoryGB != server.GetMemoryGB() ||
		b.GPU != server.GPU ||
		(b.CPUModel != "" && b.CPUModel != server.ServerPlanCPUModel) ||
		b.Commitment != server.ServerPlanCommitment ||
		(b.Generation != types.PlanGenerations.Default && b.Generation != server.ServerPlanGeneration)
	// b.Generation != server.ServerPlanGeneration
}

func (b *Builder) userData() []string {
	if b.UserData == "" {
		return nil
	}
	return []string{b.UserData}
}
