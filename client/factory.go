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
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	sacloudhttp "github.com/sacloud/go-http"
)

// Factory client.HttpRequestDoerを作成して返すファクトリー
type Factory struct {
	options    *Options
	httpClient *http.Client // Transportの初期化を1度だけ行うためにoptionsのHttpClientの参照をここにコピーして保持しておく

	once sync.Once
}

// NewFactory 指定のオプションでFactoryを生成する
func NewFactory(options ...*Options) *Factory {
	var opts *Options
	if len(options) > 0 {
		opts = MergeOptions(options...)
	}
	if opts == nil {
		panic("options is nil")
	}

	return &Factory{
		options:    opts,
		httpClient: opts.HttpClient,
	}
}

// NewHttpRequestDoer オプションを反映したsacloud向けのHTTPクライアントを生成して返す
func (f *Factory) NewHttpRequestDoer() HttpRequestDoer {
	f.init()

	checkRetryFn := retryablehttp.DefaultRetryPolicy
	if f.options.CheckRetryFunc != nil {
		checkRetryFn = f.options.CheckRetryFunc
	}

	return &sacloudhttp.Client{
		AccessToken:       f.options.AccessToken,
		AccessTokenSecret: f.options.AccessTokenSecret,
		UserAgent:         f.options.UserAgent,
		AcceptLanguage:    f.options.AcceptLanguage,
		Gzip:              f.options.Gzip,
		CheckRetryFunc:    checkRetryFn,
		RetryMax:          f.options.RetryMax,
		RetryWaitMin:      time.Duration(f.options.RetryWaitMin) * time.Second,
		RetryWaitMax:      time.Duration(f.options.RetryWaitMax) * time.Second,
		HTTPClient:        f.httpClient,
		RequestCustomizer: sacloudhttp.ComposeRequestCustomizer(f.options.RequestCustomizers...),
	}
}

// Options Doerの生成で用いるOptionsを返す
func (f *Factory) Options() *Options {
	return f.options
}

func (f *Factory) init() {
	f.once.Do(func() {
		if f.httpClient == nil {
			f.httpClient = http.DefaultClient
		}

		timeout := f.options.HttpRequestTimeout
		if timeout == 0 {
			timeout = 300
		}
		f.httpClient.Timeout = time.Duration(timeout) * time.Second

		rateLimit := f.options.HttpRequestRateLimit
		if rateLimit == 0 {
			rateLimit = 10
		}
		f.httpClient.Transport = &sacloudhttp.RateLimitRoundTripper{
			Transport:       f.httpClient.Transport,
			RateLimitPerSec: f.options.HttpRequestRateLimit,
		}

		if f.options.Trace {
			f.httpClient.Transport = &sacloudhttp.TracingRoundTripper{
				Transport: f.httpClient.Transport,
			}
		}
	})
}
