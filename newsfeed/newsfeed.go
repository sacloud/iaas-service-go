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

package newsfeed

import (
	"strconv"
	"time"
)

// FeedItems メンテナンス/障害情報お知らせ
type FeedItems []*FeedItem

// ByURL 指定のURLを持つFeedItemを返す
func (items *FeedItems) ByURL(url string) *FeedItem {
	for _, item := range *items {
		if item.URL == url {
			return item
		}
	}
	return nil
}

// FeedItem メンテナンス/障害情報お知らせ(個別)
type FeedItem struct {
	StrDate       string `json:"date,omitempty"`
	Description   string `json:"desc,omitempty"`
	StrEventStart string `json:"event_start,omitempty"`
	StrEventEnd   string `json:"event_end,omitempty"`
	Title         string `json:"title,omitempty"`
	URL           string `json:"url,omitempty"`
}

// Date 対象日時
func (f *FeedItem) Date() time.Time {
	return f.parseTime(f.StrDate)
}

// EventStart 掲載開始日時
func (f *FeedItem) EventStart() time.Time {
	return f.parseTime(f.StrEventStart)
}

// EventEnd 掲載終了日時
func (f *FeedItem) EventEnd() time.Time {
	return f.parseTime(f.StrEventEnd)
}

func (f *FeedItem) parseTime(sec string) time.Time {
	s, _ := strconv.ParseInt(sec, 10, 64)
	return time.Unix(s, 0)
}
