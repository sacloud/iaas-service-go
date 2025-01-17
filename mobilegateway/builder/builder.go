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
	"fmt"
	"reflect"
	"time"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/accessor"
	"github.com/sacloud/iaas-api-go/helper/power"
	"github.com/sacloud/iaas-api-go/types"
	setup2 "github.com/sacloud/iaas-service-go/setup"
)

// Builder モバイルゲートウェイの構築を行う
type Builder struct {
	ID   types.ID
	Zone string

	Name                            string
	Description                     string
	Tags                            types.Tags
	IconID                          types.ID
	PrivateInterface                *PrivateInterfaceSetting
	StaticRoutes                    []*iaas.MobileGatewayStaticRoute
	SIMRoutes                       []*SIMRouteSetting
	InternetConnectionEnabled       bool
	InterDeviceCommunicationEnabled bool
	DNS                             *iaas.MobileGatewayDNSSetting
	SIMs                            []*SIMSetting
	TrafficConfig                   *iaas.MobileGatewayTrafficControl

	SettingsHash string
	NoWait       bool

	SetupOptions *setup2.Options
	Client       *APIClient
}

// PrivateInterfaceSetting モバイルゲートウェイのプライベート側インターフェース設定
type PrivateInterfaceSetting struct {
	SwitchID       types.ID
	IPAddress      string
	NetworkMaskLen int
}

// SIMSetting モバイルゲートウェイに接続するSIM設定
type SIMSetting struct {
	SIMID     types.ID
	IPAddress string
}

// SIMRouteSetting SIMルート設定
type SIMRouteSetting struct {
	SIMID  types.ID
	Prefix string
}

// FromResource 既存のMobileGatewayからBuilderを組み立てて返す
func FromResource(ctx context.Context, caller iaas.APICaller, zone string, id types.ID) (*Builder, error) {
	mgwOp := iaas.NewMobileGatewayOp(caller)
	current, err := mgwOp.Read(ctx, zone, id)
	if err != nil {
		return nil, err
	}

	var privateInterface *PrivateInterfaceSetting
	for i, nic := range current.InterfaceSettings {
		if nic.Index == 1 {
			privateInterface = &PrivateInterfaceSetting{
				SwitchID:       current.Interfaces[i].SwitchID,
				IPAddress:      nic.IPAddress[0],
				NetworkMaskLen: nic.NetworkMaskLen,
			}
		}
	}

	simRoutes, err := mgwOp.GetSIMRoutes(ctx, zone, id)
	if err != nil {
		return nil, err
	}
	var simRouteSettings []*SIMRouteSetting
	for _, r := range simRoutes {
		simRouteSettings = append(simRouteSettings, &SIMRouteSetting{
			SIMID:  types.StringID(r.ResourceID),
			Prefix: r.Prefix,
		})
	}

	dns, err := mgwOp.GetDNS(ctx, zone, id)
	if err != nil {
		return nil, err
	}

	sims, err := mgwOp.ListSIM(ctx, zone, id)
	if err != nil {
		return nil, err
	}
	var simSettings []*SIMSetting
	for _, s := range sims {
		simSettings = append(simSettings, &SIMSetting{
			SIMID:     types.StringID(s.ResourceID),
			IPAddress: s.IP,
		})
	}

	trafficConfig, err := mgwOp.GetTrafficConfig(ctx, zone, id)
	if err != nil {
		return nil, err
	}

	return &Builder{
		ID:   id,
		Zone: zone,

		Name:                            current.Name,
		Description:                     current.Description,
		Tags:                            current.Tags,
		IconID:                          current.IconID,
		PrivateInterface:                privateInterface,
		StaticRoutes:                    current.StaticRoutes,
		SIMRoutes:                       simRouteSettings,
		InternetConnectionEnabled:       current.InternetConnectionEnabled.Bool(),
		InterDeviceCommunicationEnabled: current.InterDeviceCommunicationEnabled.Bool(),
		DNS:                             dns,
		SIMs:                            simSettings,
		TrafficConfig:                   trafficConfig,
		SettingsHash:                    current.SettingsHash,
		NoWait:                          false,
		Client:                          NewAPIClient(caller),
	}, nil
}

func (b *Builder) init() {
	if b.SetupOptions == nil {
		b.SetupOptions = &setup2.Options{}
	}
	b.SetupOptions.Init()
	b.SetupOptions.ProvisioningRetryCount = 1
}

// Validate 設定値の検証
func (b *Builder) Validate(ctx context.Context, zone string) error {
	if b.PrivateInterface != nil {
		if b.PrivateInterface.SwitchID.IsEmpty() {
			return fmt.Errorf("switch id is required when specified private interface")
		}
		if b.PrivateInterface.IPAddress == "" {
			return fmt.Errorf("ip address is required when specified private interface")
		}
		if b.PrivateInterface.NetworkMaskLen == 0 {
			return fmt.Errorf("ip address is required when specified private interface")
		}
	}
	if len(b.SIMRoutes) > 0 && len(b.SIMs) == 0 {
		return fmt.Errorf("sim settings are required when specified sim routes")
	}

	if b.NoWait {
		if b.PrivateInterface != nil || len(b.StaticRoutes) > 0 || len(b.SIMRoutes) > 0 || b.DNS != nil || len(b.SIMs) > 0 || b.TrafficConfig != nil {
			return errors.New("NoWait=true is not supported with PrivateInterface/StaticRoutes/SIMRoutes/DNS/SIMs/TrafficConfig")
		}
	}
	return nil
}

// Build モバイルゲートウェイの作成や設定をまとめて行う
func (b *Builder) Build(ctx context.Context) (*iaas.MobileGateway, error) {
	if b.ID.IsEmpty() {
		return b.create(ctx, b.Zone)
	}
	return b.update(ctx, b.Zone, b.ID)
}

// Build モバイルゲートウェイの作成や設定をまとめて行う
func (b *Builder) create(ctx context.Context, zone string) (*iaas.MobileGateway, error) {
	b.init()

	if err := b.Validate(ctx, zone); err != nil {
		return nil, err
	}

	builder := &setup2.RetryableSetup{
		Create: func(ctx context.Context, zone string) (accessor.ID, error) {
			return b.Client.MobileGateway.Create(ctx, zone, &iaas.MobileGatewayCreateRequest{
				Name:                            b.Name,
				Description:                     b.Description,
				Tags:                            b.Tags,
				IconID:                          b.IconID,
				InternetConnectionEnabled:       types.StringFlag(b.InternetConnectionEnabled),
				InterDeviceCommunicationEnabled: types.StringFlag(b.InterDeviceCommunicationEnabled),
			})
		},
		ProvisionBeforeUp: func(ctx context.Context, zone string, id types.ID, target interface{}) error {
			if b.NoWait {
				return nil
			}
			mgw := target.(*iaas.MobileGateway)

			// スイッチの接続
			if b.PrivateInterface != nil {
				if err := b.Client.MobileGateway.ConnectToSwitch(ctx, zone, id, b.PrivateInterface.SwitchID); err != nil {
					return err
				}
			}

			// [HACK] スイッチ接続直後だとエラーになることがあるため数秒待つ
			time.Sleep(b.SetupOptions.NICUpdateWaitDuration)

			// Interface設定
			updated, err := b.Client.MobileGateway.UpdateSettings(ctx, zone, id, &iaas.MobileGatewayUpdateSettingsRequest{
				InterfaceSettings:               b.getInterfaceSettings(),
				InternetConnectionEnabled:       types.StringFlag(b.InternetConnectionEnabled),
				InterDeviceCommunicationEnabled: types.StringFlag(b.InterDeviceCommunicationEnabled),
				SettingsHash:                    mgw.SettingsHash,
			})
			if err != nil {
				return err
			}
			// [HACK] インターフェースの設定をConfigで反映させておかないとエラーになることへの対応
			// see: https://github.com/sacloud/libsacloud/issues/589
			if err := b.Client.MobileGateway.Config(ctx, zone, id); err != nil {
				return err
			}
			mgw = updated

			// traffic config
			if b.TrafficConfig != nil {
				if err := b.Client.MobileGateway.SetTrafficConfig(ctx, zone, id, b.TrafficConfig); err != nil {
					return err
				}
			}

			// dns
			if b.DNS != nil {
				if err := b.Client.MobileGateway.SetDNS(ctx, zone, id, b.DNS); err != nil {
					return err
				}
			}

			// static route
			if len(b.StaticRoutes) > 0 {
				_, err := b.Client.MobileGateway.UpdateSettings(ctx, zone, id, &iaas.MobileGatewayUpdateSettingsRequest{
					InterfaceSettings:               b.getInterfaceSettings(),
					StaticRoutes:                    b.StaticRoutes,
					InternetConnectionEnabled:       types.StringFlag(b.InternetConnectionEnabled),
					InterDeviceCommunicationEnabled: types.StringFlag(b.InterDeviceCommunicationEnabled),
					SettingsHash:                    mgw.SettingsHash,
				})
				if err != nil {
					return err
				}
			}

			// SIMs
			if len(b.SIMs) > 0 {
				for _, sim := range b.SIMs {
					if err := b.Client.MobileGateway.AddSIM(ctx, zone, id, &iaas.MobileGatewayAddSIMRequest{SIMID: sim.SIMID.String()}); err != nil {
						return err
					}
					if err := b.Client.SIM.AssignIP(ctx, sim.SIMID, &iaas.SIMAssignIPRequest{IP: sim.IPAddress}); err != nil {
						return err
					}
				}
			}

			// SIM routes
			if len(b.SIMRoutes) > 0 {
				if err := b.Client.MobileGateway.SetSIMRoutes(ctx, zone, id, b.getSIMRouteSettings()); err != nil {
					return err
				}
			}

			if err := b.Client.MobileGateway.Config(ctx, zone, id); err != nil {
				return err
			}

			if b.SetupOptions.BootAfterBuild {
				return power.BootMobileGateway(ctx, b.Client.MobileGateway, zone, id)
			}
			return nil
		},
		Delete: func(ctx context.Context, zone string, id types.ID) error {
			return b.Client.MobileGateway.Delete(ctx, zone, id)
		},
		Read: func(ctx context.Context, zone string, id types.ID) (interface{}, error) {
			return b.Client.MobileGateway.Read(ctx, zone, id)
		},
		IsWaitForCopy: !b.NoWait,
		IsWaitForUp:   !b.NoWait && b.SetupOptions.BootAfterBuild,
		Options:       b.SetupOptions,
	}

	result, err := builder.Setup(ctx, zone)
	var mgw *iaas.MobileGateway
	if result != nil {
		mgw = result.(*iaas.MobileGateway)
	}
	if err != nil {
		return mgw, err
	}

	// refresh
	refreshed, err := b.Client.MobileGateway.Read(ctx, zone, mgw.ID)
	if err != nil {
		return mgw, err
	}
	return refreshed, nil
}

// Update モバイルゲートウェイの更新
//
// 更新中、SIMルートが一時的にクリアされます。また、接続先スイッチが変更されていた場合は再起動されます。
func (b *Builder) update(ctx context.Context, zone string, id types.ID) (*iaas.MobileGateway, error) {
	b.init()

	if err := b.Validate(ctx, zone); err != nil {
		return nil, err
	}

	// check MobileGateway is exists
	mgw, err := b.Client.MobileGateway.Read(ctx, zone, id)
	if err != nil {
		return nil, err
	}
	mgw.SettingsHash = b.SettingsHash // 更新ルートが複数あるためここに設定しておく

	isNeedShutdown, err := b.collectUpdateInfo(mgw)
	if err != nil {
		return nil, err
	}

	isNeedRestart := false
	if mgw.InstanceStatus.IsUp() && isNeedShutdown {
		if b.NoWait {
			return nil, errors.New("NoWait option is not available due to the need to shut down")
		}
		isNeedRestart = true
		if err := power.ShutdownMobileGateway(ctx, b.Client.MobileGateway, zone, id, false); err != nil {
			return nil, err
		}
	}

	// NICの切断/変更
	if b.isPrivateInterfaceChanged(mgw) {
		if len(mgw.Interfaces) > 1 && !mgw.Interfaces[1].SwitchID.IsEmpty() {
			if b.PrivateInterface != nil && mgw.Interfaces[1].SwitchID != b.PrivateInterface.SwitchID {
				// 切断
				if err := b.Client.MobileGateway.DisconnectFromSwitch(ctx, zone, id); err != nil {
					return nil, err
				}
				// [HACK] スイッチ接続直後だとエラーになることがあるため数秒待つ
				time.Sleep(b.SetupOptions.NICUpdateWaitDuration)

				updated, err := b.Client.MobileGateway.UpdateSettings(ctx, zone, id, &iaas.MobileGatewayUpdateSettingsRequest{
					InternetConnectionEnabled:       types.StringFlag(b.InternetConnectionEnabled),
					InterDeviceCommunicationEnabled: types.StringFlag(b.InterDeviceCommunicationEnabled),
					SettingsHash:                    mgw.SettingsHash,
				})
				if err != nil {
					return nil, err
				}
				// [HACK] インターフェースの設定をConfigで反映させておかないとエラーになることへの対応
				// see: https://github.com/sacloud/libsacloud/issues/589
				if err := b.Client.MobileGateway.Config(ctx, zone, id); err != nil {
					return nil, err
				}
				mgw = updated
			}
		}

		// 接続
		if b.PrivateInterface != nil {
			if len(mgw.Interfaces) == 1 {
				// スイッチの接続
				if err := b.Client.MobileGateway.ConnectToSwitch(ctx, zone, id, b.PrivateInterface.SwitchID); err != nil {
					return nil, err
				}

				// [HACK] スイッチ接続直後だとエラーになることがあるため数秒待つ
				time.Sleep(b.SetupOptions.NICUpdateWaitDuration)
			}

			// Interface設定
			updated, err := b.Client.MobileGateway.UpdateSettings(ctx, zone, id, &iaas.MobileGatewayUpdateSettingsRequest{
				InterfaceSettings:               b.getInterfaceSettings(),
				InternetConnectionEnabled:       types.StringFlag(b.InternetConnectionEnabled),
				InterDeviceCommunicationEnabled: types.StringFlag(b.InterDeviceCommunicationEnabled),
				SettingsHash:                    mgw.SettingsHash,
			})
			if err != nil {
				return nil, err
			}
			// [HACK] インターフェースの設定をConfigで反映させておかないとエラーになることへの対応
			// see: https://github.com/sacloud/libsacloud/issues/589
			if err := b.Client.MobileGateway.Config(ctx, zone, id); err != nil {
				return nil, err
			}
			mgw = updated
		}
	}

	mgw, err = b.Client.MobileGateway.Update(ctx, zone, id, &iaas.MobileGatewayUpdateRequest{
		Name:                            b.Name,
		Description:                     b.Description,
		Tags:                            b.Tags,
		IconID:                          b.IconID,
		InterfaceSettings:               b.getInterfaceSettings(),
		InternetConnectionEnabled:       types.StringFlag(b.InternetConnectionEnabled),
		InterDeviceCommunicationEnabled: types.StringFlag(b.InterDeviceCommunicationEnabled),
		SettingsHash:                    mgw.SettingsHash,
	})
	if err != nil {
		return nil, err
	}

	// traffic config
	trafficConfig, err := b.Client.MobileGateway.GetTrafficConfig(ctx, zone, id)
	if err != nil {
		if !iaas.IsNotFoundError(err) {
			return nil, err
		}
	}
	if !reflect.DeepEqual(trafficConfig, b.TrafficConfig) {
		if trafficConfig != nil && b.TrafficConfig == nil {
			if err := b.Client.MobileGateway.DeleteTrafficConfig(ctx, zone, id); err != nil {
				return nil, err
			}
		} else {
			if err := b.Client.MobileGateway.SetTrafficConfig(ctx, zone, id, b.TrafficConfig); err != nil {
				return nil, err
			}
		}
	}

	// dns
	dns, err := b.Client.MobileGateway.GetDNS(ctx, zone, id)
	if err != nil {
		if !iaas.IsNotFoundError(err) {
			return nil, err
		}
	}
	if !reflect.DeepEqual(dns, b.DNS) {
		if dns == nil {
			zone, err := b.Client.Zone.Read(ctx, mgw.ZoneID)
			if err != nil {
				return nil, err
			}
			b.DNS = &iaas.MobileGatewayDNSSetting{
				DNS1: zone.Region.NameServers[0],
				DNS2: zone.Region.NameServers[1],
			}
		}
		if err := b.Client.MobileGateway.SetDNS(ctx, zone, id, b.DNS); err != nil {
			return nil, err
		}
	}

	// static route(
	if len(b.StaticRoutes) > 0 {
		_, err := b.Client.MobileGateway.UpdateSettings(ctx, zone, id, &iaas.MobileGatewayUpdateSettingsRequest{
			InterfaceSettings:               b.getInterfaceSettings(),
			StaticRoutes:                    b.StaticRoutes,
			InternetConnectionEnabled:       types.StringFlag(b.InternetConnectionEnabled),
			InterDeviceCommunicationEnabled: types.StringFlag(b.InterDeviceCommunicationEnabled),
			SettingsHash:                    mgw.SettingsHash,
		})
		if err != nil {
			return nil, err
		}
	}

	// SIMs and SIMRoutes
	currentSIMs, err := b.currentConnectedSIMs(ctx, zone, id)
	if err != nil {
		return nil, err
	}
	currentSIMRoutes, err := b.currentSIMRoutes(ctx, zone, id)
	if err != nil {
		return nil, err
	}

	if !reflect.DeepEqual(currentSIMs, b.SIMs) || !reflect.DeepEqual(currentSIMRoutes, b.SIMRoutes) {
		if len(currentSIMRoutes) > 0 {
			// SIMルートクリア
			if err := b.Client.MobileGateway.SetSIMRoutes(ctx, zone, id, []*iaas.MobileGatewaySIMRouteParam{}); err != nil {
				return nil, err
			}
		}
		// SIM変更
		added, updated, deleted := b.changedSIMs(currentSIMs, b.SIMs)
		for _, sim := range deleted {
			if err := b.Client.SIM.ClearIP(ctx, sim.SIMID); err != nil {
				return nil, err
			}
			if err := b.Client.MobileGateway.DeleteSIM(ctx, zone, id, sim.SIMID); err != nil {
				return nil, err
			}
		}
		for _, sim := range updated {
			if err := b.Client.SIM.ClearIP(ctx, sim.SIMID); err != nil {
				return nil, err
			}
			if err := b.Client.SIM.AssignIP(ctx, sim.SIMID, &iaas.SIMAssignIPRequest{IP: sim.IPAddress}); err != nil {
				return nil, err
			}
		}
		for _, sim := range added {
			if err := b.Client.MobileGateway.AddSIM(ctx, zone, id, &iaas.MobileGatewayAddSIMRequest{SIMID: sim.SIMID.String()}); err != nil {
				return nil, err
			}
			if err := b.Client.SIM.AssignIP(ctx, sim.SIMID, &iaas.SIMAssignIPRequest{IP: sim.IPAddress}); err != nil {
				return nil, err
			}
		}
		if len(b.SIMRoutes) > 0 {
			if err := b.Client.MobileGateway.SetSIMRoutes(ctx, zone, id, b.getSIMRouteSettings()); err != nil {
				return nil, err
			}
		}
	}

	if err := b.Client.MobileGateway.Config(ctx, zone, id); err != nil {
		return nil, err
	}

	if isNeedRestart {
		if err := power.BootMobileGateway(ctx, b.Client.MobileGateway, zone, id); err != nil {
			return nil, err
		}
	}

	// refresh
	mgw, err = b.Client.MobileGateway.Read(ctx, zone, id)
	if err != nil {
		return nil, err
	}
	return mgw, err
}

func (b *Builder) getInterfaceSettings() []*iaas.MobileGatewayInterfaceSetting {
	if b.PrivateInterface == nil {
		return nil
	}
	return []*iaas.MobileGatewayInterfaceSetting{
		{
			Index:          1,
			NetworkMaskLen: b.PrivateInterface.NetworkMaskLen,
			IPAddress:      []string{b.PrivateInterface.IPAddress},
		},
	}
}

func (b *Builder) getSIMRouteSettings() []*iaas.MobileGatewaySIMRouteParam {
	var results []*iaas.MobileGatewaySIMRouteParam
	for _, route := range b.SIMRoutes {
		results = append(results, &iaas.MobileGatewaySIMRouteParam{
			ResourceID: route.SIMID.String(),
			Prefix:     route.Prefix,
		})
	}
	return results
}

func (b *Builder) collectUpdateInfo(mgw *iaas.MobileGateway) (isNeedShutdown bool, err error) {
	// スイッチの変更/削除は再起動が必要
	isNeedShutdown = b.isPrivateInterfaceChanged(mgw)
	return
}

func (b *Builder) isPrivateInterfaceChanged(mgw *iaas.MobileGateway) bool {
	current := b.currentPrivateInterfaceState(mgw)
	return !reflect.DeepEqual(current, b.PrivateInterface)
}

func (b *Builder) currentPrivateInterfaceState(mgw *iaas.MobileGateway) *PrivateInterfaceSetting {
	if len(mgw.Interfaces) > 1 {
		switchID := mgw.Interfaces[1].SwitchID
		var setting *iaas.MobileGatewayInterfaceSetting
		for _, s := range mgw.InterfaceSettings {
			if s.Index == 1 {
				setting = s
			}
		}
		if setting != nil {
			var ip string
			if len(setting.IPAddress) > 0 {
				ip = setting.IPAddress[0]
			}
			return &PrivateInterfaceSetting{
				SwitchID:       switchID,
				IPAddress:      ip,
				NetworkMaskLen: setting.NetworkMaskLen,
			}
		}
	}
	return nil
}

func (b *Builder) currentConnectedSIMs(ctx context.Context, zone string, id types.ID) ([]*SIMSetting, error) {
	var results []*SIMSetting

	sims, err := b.Client.MobileGateway.ListSIM(ctx, zone, id)
	if err != nil && !iaas.IsNotFoundError(err) {
		return results, err
	}
	for _, sim := range sims {
		results = append(results, &SIMSetting{
			SIMID:     types.StringID(sim.ResourceID),
			IPAddress: sim.IP,
		})
	}
	return results, nil
}

func (b *Builder) currentSIMRoutes(ctx context.Context, zone string, id types.ID) ([]*iaas.MobileGatewaySIMRoute, error) {
	return b.Client.MobileGateway.GetSIMRoutes(ctx, zone, id)
}

func (b *Builder) changedSIMs(current []*SIMSetting, desired []*SIMSetting) (added, updated, deleted []*SIMSetting) {
	for _, c := range current {
		isExists := false
		for _, d := range desired {
			if c.SIMID == d.SIMID {
				isExists = true
				if c.IPAddress != d.IPAddress {
					updated = append(updated, d)
				}
			}
		}
		if !isExists {
			deleted = append(deleted, c)
		}
	}
	for _, d := range desired {
		isExists := false
		for _, c := range current {
			if c.SIMID == d.SIMID {
				isExists = true
				continue
			}
		}
		if !isExists {
			added = append(added, d)
		}
	}
	return
}
