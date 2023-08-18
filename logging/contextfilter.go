// Copyright 2016-2023 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

import (
	"context"
	"sort"
)

// FilteredContextData holds information that has been extracted from a context.Context in a format suitable for logging
type FilteredContextData map[string]string

// ContextFilter takes a context and extracts some or all of the data in it a returns in a form suitable for
// inclusion in application and access logs.
type ContextFilter interface {
	Extract(ctx context.Context) FilteredContextData
}

// PrioritisedContextFilter groups together a series of other ContextFilter implementations in a priority order. When
// requests are received for filtered data, the data is gathered from the ContextFilter with the lowest priority. If the
// key is present in a higher priority filter, it overwrites the value in the lower priority filter.
type PrioritisedContextFilter struct {
	filters []ContextFilter
}

// Add inserts another ContextFilter to the list of prioritised filters and sorts the list
func (pcf *PrioritisedContextFilter) Add(cf ContextFilter) {

	if _, okay := cf.(FilterPriority); !okay {
		cf = zeroPriorityWrapper{cf}
	}

	if pcf.filters == nil {
		pcf.filters = make([]ContextFilter, 1)
		pcf.filters[0] = cf

		return
	}

	pcf.filters = append(pcf.filters, cf)
	pcf.sort()
}

func (pcf *PrioritisedContextFilter) sort() {

	sort.Slice(pcf.filters, func(i, j int) bool {

		return pcf.filters[i].(FilterPriority).Priority() > pcf.filters[j].(FilterPriority).Priority()

	})

}

// Extract takes all of the FilteredContextData from the lowest priority Context filter then overwrites it with data
// from higher priority instances.
func (pcf PrioritisedContextFilter) Extract(ctx context.Context) FilteredContextData {

	filterCount := len(pcf.filters)

	if filterCount == 1 {
		return pcf.filters[0].Extract(ctx)
	}

	var d = make(FilteredContextData)

	for i := filterCount - 1; i >= 0; i-- {

		p := pcf.filters[i].Extract(ctx)

		for k, v := range p {
			d[k] = v
		}

	}

	return d
}

// FilterPriority is implemented by ContextFilters that need to make sure their view of FilteredContextData is prioritised
// over those in another ContextFilter. Higher values are considered to be higher priority and negative values are allowed.
type FilterPriority interface {
	Priority() int64
}

type zeroPriorityWrapper struct {
	wrapped ContextFilter
}

func (zpm zeroPriorityWrapper) Extract(ctx context.Context) FilteredContextData {
	return zpm.wrapped.Extract(ctx)
}

func (zpm zeroPriorityWrapper) Priority() int64 {
	return 0
}
