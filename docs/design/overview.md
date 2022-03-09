# sacloud-go

- URL: https://github.com/sacloud/sacloud-go/pull/1
- Author: @yamamoto-febc

## 概要

さくらのクラウドではいくつかの性質の異なるAPIが提供されている。  

- IaaS API
- 請求関連
- ウェブアクセラレータ API
- オブジェクトストレージ API
- 専用サーバPHY API

従来はIaas API(sacloud/libsacloud)としてこれらの一部のAPI向けのAPIライブラリを提供していた。

さくらのクラウドのサービス拡大と共に専用サーバPHYやオブジェクトストレージのAPIなどが公開されたが、
これらのAPIはOpenAPIによるAPI定義が提供されていたりHTTP レスポンスステータスコードの扱いが異なったりと
Iaas APIとはライブラリで処理すべき内容が異なっている。

このままIaaS API側で実装すると(違いを吸収するための)共通部分の実装の肥大化が予想されるため、APIライブラリを以下のように分ける。  

- [IaaS API: sacloud/iaas-api-go](https://github.com/sacloud/iaas-api-go)  
  以下のAPIも含む  
    - 請求 API
    - ウェブアクセラレータ API
- [オブジェクトストレージ API: sacloud/object-storage-api-go](https://github.com/sacloud/object-storage-api-go)
- [専用サーバPHY API: sacloud/phy-go](https://github.com/sacloud/phy-go)

各ライブラリは単体でも利用可能とするが、Usacloudに対するlibsacloudのhelper/serviceパッケージのように、各APIに対し統一的にCRUD+L操作が行えるようにしたい。  

そこで、高レベルAPIライブラリとしてsacloud/sacloud-goを作成し、
libsacloudのhelper/serviceパッケージの移植やその他高レベル操作の実装/集約などを行いたい。  

## やること/やらないこと

### やること

- libsacloudからの移植
  - helper/builderとhelper/serviceの統合
  - 統合したhelper/serviceの移植
  - helper/newsfeedの移植
  - helper/waitを汎化して切り出し
  - helper/apiを汎化して切り出し
  - 汎用処理の切り出し
  - クライアント周り(helper/apiやprofile)の切り出し

### やらないこと

TODO: 必要に応じて記載

## 実装

- libsacloudから必要部分を移植  
- 下位ライブラリ(IaaS/オブジェクトストレージ/PHY)を意識せず使えるようなAPIの設計

### パッケージ構成

```console
- sacloud-go
  - pkg       # Note: libsacloudからの移植、iaas-api-goなどからも利用するため独立したパッケージにする
    - go.mod
    - cidr
    - mutexkv
    - size
    - pointer
    - wait      
    
  - client    # Note: libsacloud/helper/apiなどからの移植、iaas-api-goなどからも利用するため独立したパッケージにする
    - go.mod
    - profile
    - (helper/apiから切り出し)
  - newsfeed
  - service    
    - iaas          # libsacloudのhelper/serviceからの移植 
    - objectstorage # 新規
    - phy           # 新規
```

`pkg`は独立したリポジトリにするほどでもない、sacloudドメインに非依存のコードを配置する。

## 改訂履歴

- 2022/2/26: 初版作成
- 2022/3/7: パッケージ構成追加