# sacloud-goでのHTTP/APIクライアントの共通化

- URL: https://github.com/sacloud/sacloud-go/pull/8
- Author: @yamamoto-febc

## 概要

IaaS/オブジェクトストレージ/PhyでHTTP/APIクライアント周りの処理を共通化したい。

libsacloudでは`helper/api`や`sacloud/profile`パッケージを用いて`sacloud.Client`を組み立てる実装となっている。  
HTTPクライアントレベルでは`sacloud/go-http`で共通化されているが、環境変数からのAPIキーの読み込みやプロファイルのサポートなどは
libsacloudで実装されておりPHYやオブジェクトストレージで利用するのが困難。  

このためsacloud-goでこれらを実装し共通で利用可能にする。

## やること/やらないこと

### やること

TODO 必要に応じて追記

### やらないこと

TODO 必要に応じて追記

## 実装

sacloud-go側で各種設定値を保持し、sacloud/go-httpの`Client`を返すところまでを実装する。  
各クライアント側プロジェクトから参照されることによる循環参照を防ぐためにsacloud-go配下で独立したモジュールとして実装する。

## 改訂履歴

- 2022/3/8: 初版作成