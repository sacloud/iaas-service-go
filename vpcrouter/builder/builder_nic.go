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
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
)

// NICSettingHolder VPCルータのeth0の設定 SharedNICSettingまたはRouterNICSettingを指定する
type NICSettingHolder interface {
	getConnectedSwitch() *iaas.ApplianceConnectedSwitch
	getIPAddresses() []string
	getInterfaceSetting() *iaas.VPCRouterInterfaceSetting
}

// StandardNICSetting VPCルータのeth0を共有セグメントに接続するためのSetting(スタンダードプラン)
type StandardNICSetting struct{}

func (s *StandardNICSetting) getConnectedSwitch() *iaas.ApplianceConnectedSwitch {
	return &iaas.ApplianceConnectedSwitch{Scope: types.Scopes.Shared}
}

func (s *StandardNICSetting) getIPAddresses() []string {
	return nil
}

func (s *StandardNICSetting) getInterfaceSetting() *iaas.VPCRouterInterfaceSetting {
	return nil
}

// PremiumNICSetting VPCルータのeth0をスイッチ+ルータに接続するためのSetting(プレミアム/ハイスペックプラン)
type PremiumNICSetting struct {
	SwitchID         types.ID
	IPAddresses      []string
	VirtualIPAddress string
	IPAliases        []string
}

func (s *PremiumNICSetting) getConnectedSwitch() *iaas.ApplianceConnectedSwitch {
	return &iaas.ApplianceConnectedSwitch{ID: s.SwitchID}
}

func (s *PremiumNICSetting) getIPAddresses() []string {
	return s.IPAddresses
}

func (s *PremiumNICSetting) getInterfaceSetting() *iaas.VPCRouterInterfaceSetting {
	return &iaas.VPCRouterInterfaceSetting{
		IPAddress:        s.IPAddresses,
		VirtualIPAddress: s.VirtualIPAddress,
		IPAliases:        s.IPAliases,
		Index:            0,
	}
}

// AdditionalNICSettingHolder VPCルータのeth1-eth7の設定
type AdditionalNICSettingHolder interface {
	getSwitchInfo() (switchID types.ID, index int)
	getInterfaceSetting() *iaas.VPCRouterInterfaceSetting
}

// AdditionalStandardNICSetting VPCルータのeth1-eth7の設定(スタンダードプラン向け)
type AdditionalStandardNICSetting struct {
	SwitchID       types.ID
	IPAddress      string
	NetworkMaskLen int
	Index          int
}

func (s *AdditionalStandardNICSetting) getSwitchInfo() (switchID types.ID, index int) {
	return s.SwitchID, s.Index
}

func (s *AdditionalStandardNICSetting) getInterfaceSetting() *iaas.VPCRouterInterfaceSetting {
	return &iaas.VPCRouterInterfaceSetting{
		IPAddress:      []string{s.IPAddress},
		NetworkMaskLen: s.NetworkMaskLen,
		Index:          s.Index,
	}
}

// AdditionalPremiumNICSetting VPCルータのeth1-eth7の設定(プレミアム/ハイスペックプラン向け)
type AdditionalPremiumNICSetting struct {
	SwitchID         types.ID
	IPAddresses      []string
	VirtualIPAddress string
	NetworkMaskLen   int
	Index            int
}

func (s *AdditionalPremiumNICSetting) getSwitchInfo() (switchID types.ID, index int) {
	return s.SwitchID, s.Index
}

func (s *AdditionalPremiumNICSetting) getInterfaceSetting() *iaas.VPCRouterInterfaceSetting {
	return &iaas.VPCRouterInterfaceSetting{
		IPAddress:        s.IPAddresses,
		VirtualIPAddress: s.VirtualIPAddress,
		NetworkMaskLen:   s.NetworkMaskLen,
		Index:            s.Index,
	}
}
