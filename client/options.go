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

package client

import (
	"context"
	"net/http"
	"os"
	"strconv"

	sacloudhttp "github.com/sacloud/go-http"
	"github.com/sacloud/sacloud-go/client/profile"
)

// Options sacloudhttp.Clientを作成する際のオプション
type Options struct {
	// AccessToken APIキー:トークン
	AccessToken string
	// AccessTokenSecret APIキー:シークレット
	AccessTokenSecret string

	// AcceptLanguage APIリクエスト時のAccept-Languageヘッダーの値
	AcceptLanguage string

	// Gzip APIリクエストでgzipを有効にするかのフラグ
	Gzip bool

	// HttpClient APIリクエストで使用されるHTTPクライアント
	//
	// 省略した場合はhttp.DefaultClientが使用される
	HttpClient *http.Client

	// HttpRequestTimeout HTTPリクエストのタイムアウト秒数
	HttpRequestTimeout int
	// HttpRequestRateLimit 1秒あたりの上限リクエスト数
	HttpRequestRateLimit int

	// RetryMax リトライ上限回数
	RetryMax int

	// RetryWaitMax リトライ待ち秒数(最大)
	RetryWaitMax int
	// RetryWaitMin リトライ待ち秒数(最小)
	RetryWaitMin int

	// UserAgent ユーザーエージェント
	UserAgent string

	// Trace HTTPリクエスト/レスポンスのトレースログ(ダンプ)出力
	Trace bool

	// RequestCustomizers リクエスト前に*http.Requestのカスタマイズを行うためのfunc
	RequestCustomizers []sacloudhttp.RequestCustomizer

	// CheckRetryFunc リトライすべきか判定するためのfunc
	CheckRetryFunc func(ctx context.Context, resp *http.Response, err error) (bool, error)

	// profileConfigValue プロファイルから読み込んだ値を保持する
	profileConfigValue *profile.ConfigValue
}

// ProfileConfigValue プロファイルから読み込んだprofile.ConfigValueを返す
func (o *Options) ProfileConfigValue() *profile.ConfigValue {
	return o.profileConfigValue
}

// DefaultOption 環境変数、プロファイルからCallerOptionsを組み立てて返す
//
// プロファイルは環境変数`SAKURACLOUD_PROFILE`または`USACLOUD_PROFILE`でプロファイル名が指定されていればそちらを優先し、
// 未指定の場合は通常のプロファイル処理(~/.usacloud/currentファイルから読み込み)される。
// 同じ項目を複数箇所で指定していた場合、環境変数->プロファイルの順で上書きされたものが返される
func DefaultOption() (*Options, error) {
	return DefaultOptionWithProfile("")
}

// DefaultOptionWithProfile 環境変数、プロファイルからCallerOptionsを組み立てて返す
//
// プロファイルは引数を優先し、空の場合は環境変数`SAKURACLOUD_PROFILE`または`USACLOUD_PROFILE`が利用され、
// それも空の場合は通常のプロファイル処理(~/.usacloud/currentファイルから読み込み)される。
// 同じ項目を複数箇所で指定していた場合、環境変数->プロファイルの順で上書きされたものが返される
func DefaultOptionWithProfile(profileName string) (*Options, error) {
	if profileName == "" {
		profileName = stringFromEnvMulti([]string{"SAKURACLOUD_PROFILE", "USACLOUD_PROFILE"}, "")
	}
	fromProfile, err := OptionsFromProfile(profileName)
	if err != nil {
		return nil, err
	}
	return MergeOptions(OptionsFromEnv(), fromProfile, defaultOption), nil
}

var defaultOption = &Options{
	HttpRequestTimeout:   300,
	HttpRequestRateLimit: 5,
	RetryMax:             sacloudhttp.DefaultRetryMax,
	RetryWaitMax:         int(sacloudhttp.DefaultRetryWaitMax.Seconds()),
	RetryWaitMin:         int(sacloudhttp.DefaultRetryWaitMin.Seconds()),
}

// MergeOptions 指定のCallerOptionsの非ゼロ値フィールドをoのコピーにマージして返す
func MergeOptions(opts ...*Options) *Options {
	merged := &Options{}
	for _, opt := range opts {
		if opt.AccessToken != "" {
			merged.AccessToken = opt.AccessToken
		}
		if opt.AccessTokenSecret != "" {
			merged.AccessTokenSecret = opt.AccessTokenSecret
		}
		if opt.AcceptLanguage != "" {
			merged.AcceptLanguage = opt.AcceptLanguage
		}
		if opt.HttpClient != nil {
			merged.HttpClient = opt.HttpClient
		}
		if opt.HttpRequestTimeout > 0 {
			merged.HttpRequestTimeout = opt.HttpRequestTimeout
		}
		if opt.HttpRequestRateLimit > 0 {
			merged.HttpRequestRateLimit = opt.HttpRequestRateLimit
		}
		if opt.RetryMax > 0 {
			merged.RetryMax = opt.RetryMax
		}
		if opt.RetryWaitMax > 0 {
			merged.RetryWaitMax = opt.RetryWaitMax
		}
		if opt.RetryWaitMin > 0 {
			merged.RetryWaitMin = opt.RetryWaitMin
		}
		if opt.UserAgent != "" {
			merged.UserAgent = opt.UserAgent
		}

		if opt.profileConfigValue != nil {
			merged.profileConfigValue = opt.profileConfigValue
		}

		// Note: bool値は一度trueにしたらMergeでfalseになることがない
		if opt.Trace {
			merged.Trace = true
		}
	}
	return merged
}

// OptionsFromEnv 環境変数からCallerOptionsを組み立てて返す
func OptionsFromEnv() *Options {
	return &Options{
		AccessToken:       stringFromEnv("SAKURACLOUD_ACCESS_TOKEN", ""),
		AccessTokenSecret: stringFromEnv("SAKURACLOUD_ACCESS_TOKEN_SECRET", ""),

		AcceptLanguage: stringFromEnv("SAKURACLOUD_ACCEPT_LANGUAGE", ""),

		HttpRequestTimeout:   intFromEnv("SAKURACLOUD_API_REQUEST_TIMEOUT", 0),
		HttpRequestRateLimit: intFromEnv("SAKURACLOUD_API_REQUEST_RATE_LIMIT", 0),

		RetryMax:     intFromEnv("SAKURACLOUD_RETRY_MAX", 0),
		RetryWaitMax: intFromEnv("SAKURACLOUD_RETRY_WAIT_MAX", 0),
		RetryWaitMin: intFromEnv("SAKURACLOUD_RETRY_WAIT_MIN", 0),

		Trace: stringFromEnv("SAKURACLOUD_TRACE", "") != "",
	}
}

// OptionsFromProfile 指定のプロファイルからCallerOptionsを組み立てて返す
// プロファイル名に空文字が指定された場合はカレントプロファイルが利用される
func OptionsFromProfile(profileName string) (*Options, error) {
	if profileName == "" {
		current, err := profile.CurrentName()
		if err != nil {
			return nil, err
		}
		profileName = current
	}

	config := profile.ConfigValue{}
	if err := profile.Load(profileName, &config); err != nil {
		return nil, err
	}

	return &Options{
		AccessToken:          config.AccessToken,
		AccessTokenSecret:    config.AccessTokenSecret,
		AcceptLanguage:       config.AcceptLanguage,
		HttpRequestTimeout:   config.HTTPRequestTimeout,
		HttpRequestRateLimit: config.HTTPRequestRateLimit,
		RetryMax:             config.RetryMax,
		RetryWaitMax:         config.RetryWaitMax,
		RetryWaitMin:         config.RetryWaitMin,
		Trace:                config.EnableHTTPTrace(),

		profileConfigValue: &config,
	}, nil
}

func stringFromEnv(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func stringFromEnvMulti(keys []string, defaultValue string) string {
	for _, key := range keys {
		v := os.Getenv(key)
		if v != "" {
			return v
		}
	}
	return defaultValue
}

func intFromEnv(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return defaultValue
	}
	return int(i)
}
