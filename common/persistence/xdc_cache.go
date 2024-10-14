// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package persistence

import (
	"fmt"
	"time"

	commonpb "go.temporal.io/api/common/v1"
	historypb "go.temporal.io/api/history/v1"
	historyspb "go.temporal.io/server/api/history/v1"
	persistencepb "go.temporal.io/server/api/persistence/v1"
	workflowspb "go.temporal.io/server/api/workflow/v1"
	"go.temporal.io/server/common/cache"
	"go.temporal.io/server/common/definition"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
	"go.temporal.io/server/common/persistence/serialization"
	"go.temporal.io/server/common/persistence/versionhistory"
)

type (
	XDCCacheKey struct {
		WorkflowKey definition.WorkflowKey
		MinEventID  int64 // inclusive
		Version     int64
	}
	XDCCacheValue struct {
		BaseWorkflowInfo    *workflowspb.BaseExecutionInfo
		VersionHistoryItems []*historyspb.VersionHistoryItem
		EventBlobs          []*commonpb.DataBlob
		NextEventID         int64
	}

	XDCCache interface {
		Put(key XDCCacheKey, value XDCCacheValue)
		Get(key XDCCacheKey) (XDCCacheValue, bool)
	}

	XDCCacheImpl struct {
		cache      cache.Cache
		logger     log.Logger
		serializer serialization.Serializer
	}
)

const (
	xdcMinCacheSize = 64 * 1024 // 64KB
)

var _ XDCCache = (*XDCCacheImpl)(nil)
var _ cache.SizeGetter = XDCCacheValue{}

func NewXDCCacheKey(
	workflowKey definition.WorkflowKey,
	minEventID int64,
	version int64,
) XDCCacheKey {
	return XDCCacheKey{
		WorkflowKey: workflowKey,
		MinEventID:  minEventID,
		Version:     version,
	}
}

func NewXDCCacheValue(
	baseWorkflowInfo *workflowspb.BaseExecutionInfo,
	versionHistoryItems []*historyspb.VersionHistoryItem,
	eventBlobs []*commonpb.DataBlob,
	nextEventID int64,
) XDCCacheValue {
	return XDCCacheValue{
		BaseWorkflowInfo:    baseWorkflowInfo,
		VersionHistoryItems: versionHistoryItems,
		EventBlobs:          eventBlobs,
		NextEventID:         nextEventID,
	}
}

func (v XDCCacheValue) CacheSize() int {
	size := 0
	for _, item := range v.VersionHistoryItems {
		size += item.Size()
	}
	for _, blob := range v.EventBlobs {
		size += blob.Size()
	}
	return v.BaseWorkflowInfo.Size() + size
}

func NewEventsBlobCache(
	maxBytes int,
	ttl time.Duration,
	logger log.Logger,
) *XDCCacheImpl {
	return &XDCCacheImpl{
		cache: cache.New(
			max(xdcMinCacheSize, maxBytes),
			&cache.Options{
				TTL: ttl,
				Pin: false,
			},
		),
		logger:     logger,
		serializer: serialization.NewSerializer(),
	}
}

func (e *XDCCacheImpl) Put(
	key XDCCacheKey,
	value XDCCacheValue,
) {
	existingValue, found := e.Get(key)
	if found && existingValue.NextEventID != value.NextEventID {
		deserializeBlobs := func(blobs []*commonpb.DataBlob) [][]*historypb.HistoryEvent {
			events := make([][]*historypb.HistoryEvent, len(blobs))
			for i, blob := range blobs {
				var err error
				events[i], err = e.serializer.DeserializeEvents(blob)
				if err != nil {
					e.logger.Error("Error deserializing events", tag.Error(err))
					return nil
				}
			}
			return events
		}
		e.logger.Error(fmt.Sprintf("Putting duplicate key in XDC cache: wf-key: %v, existing event blobs: %v, new event blobs: %v", key.WorkflowKey, deserializeBlobs(existingValue.EventBlobs), deserializeBlobs(value.EventBlobs)))
	}
	e.cache.Put(key, value)
}

func (e *XDCCacheImpl) Get(key XDCCacheKey) (XDCCacheValue, bool) {
	value := e.cache.Get(key)
	if value == nil {
		return XDCCacheValue{}, false
	}
	return value.(XDCCacheValue), true
}

func GetXDCCacheValue(
	executionInfo *persistencepb.WorkflowExecutionInfo,
	eventID int64,
	version int64,
) ([]*historyspb.VersionHistoryItem, []byte, *workflowspb.BaseExecutionInfo, error) {
	baseWorkflowInfo := CopyBaseWorkflowInfo(executionInfo.BaseExecutionInfo)
	versionHistories := executionInfo.VersionHistories
	versionHistoryIndex, err := versionhistory.FindFirstVersionHistoryIndexByVersionHistoryItem(
		versionHistories,
		versionhistory.NewVersionHistoryItem(
			eventID,
			version,
		),
	)
	if err != nil {
		return nil, nil, nil, err
	}

	versionHistoryBranch, err := versionhistory.GetVersionHistory(versionHistories, versionHistoryIndex)
	if err != nil {
		return nil, nil, nil, err
	}
	return versionhistory.CopyVersionHistory(versionHistoryBranch).GetItems(), versionHistoryBranch.GetBranchToken(), baseWorkflowInfo, nil
}

func CopyBaseWorkflowInfo(
	baseWorkflowInfo *workflowspb.BaseExecutionInfo,
) *workflowspb.BaseExecutionInfo {
	if baseWorkflowInfo == nil {
		return nil
	}
	return &workflowspb.BaseExecutionInfo{
		RunId:                            baseWorkflowInfo.RunId,
		LowestCommonAncestorEventId:      baseWorkflowInfo.LowestCommonAncestorEventId,
		LowestCommonAncestorEventVersion: baseWorkflowInfo.LowestCommonAncestorEventVersion,
	}
}
