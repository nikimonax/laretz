package domain

import (
	"sort"
	"time"
)

const hoursPerDay = 24

// QuoteCoverage is the set of hours of an ExternalAsset for which PricePoints
// exist. An hourly point covers its own hour; a daily point covers the whole
// day it starts, because the source serves nothing more granular for it.
type QuoteCoverage struct {
	hours map[time.Time]struct{}
}

// NewQuoteCoverage builds the coverage of one ExternalAsset from its
// recorded PricePoints.
func NewQuoteCoverage(points []PricePoint) QuoteCoverage {
	coverage := QuoteCoverage{hours: make(map[time.Time]struct{}, len(points))}

	for _, point := range points {
		at := point.At.UTC().Truncate(time.Hour)

		if point.Granularity == GranularityDay {
			for offset := range hoursPerDay {
				coverage.hours[at.Add(time.Duration(offset)*time.Hour)] = struct{}{}
			}

			continue
		}

		coverage.hours[at] = struct{}{}
	}

	return coverage
}

// Covers reports whether the PricePoints already cover the hour.
func (coverage QuoteCoverage) Covers(hour time.Time) bool {
	_, ok := coverage.hours[hour.UTC().Truncate(time.Hour)]
	return ok
}

// CoveragePlanner computes the QuoteRanges still missing for an ExternalAsset:
// the hours required by its HoldingIntervals minus its QuoteCoverage, merged
// into continuous spans.
type CoveragePlanner struct{}

// Plan returns the continuous ranges of hours required by the holding
// intervals of the ExternalAsset but missing from coverage.
func (CoveragePlanner) Plan(asset ExternalAsset, intervals []HoldingInterval, coverage QuoteCoverage) []QuoteRange {
	sorted := make([]HoldingInterval, len(intervals))

	copy(sorted, intervals)

	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].From.Before(sorted[j].From)
	})

	var ranges []QuoteRange

	for _, interval := range mergeHoldingIntervals(sorted) {
		for hour := interval.From.UTC(); !hour.After(interval.To.UTC()); hour = hour.Add(time.Hour) {
			if coverage.Covers(hour) {
				continue
			}

			if last := len(ranges) - 1; last >= 0 && hour.Equal(ranges[last].To.Add(time.Hour)) {
				ranges[last].To = hour
				continue
			}

			ranges = append(ranges, QuoteRange{
				SourceID:   asset.SourceID,
				ExternalID: asset.ExternalID,
				From:       hour,
				To:         hour,
			})
		}
	}

	return ranges
}
