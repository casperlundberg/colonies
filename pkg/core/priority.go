package core

import (
	"encoding/json"
	"time"
)

// PriorityTimeUnit is the queue displacement of one priority unit, in
// nanoseconds. It is NEGATIVE: a higher priority subtracts more and therefore
// sorts earlier under the assign path's ORDER BY PRIORITYTIME.
//
// One unit is one day, so any practical priority ladder behaves as strict
// priority tiers with FIFO inside a tier.
const PriorityTimeUnit int64 = -1000000000 * 60 * 60 * 24

// ComputePriorityTime derives the ordering key from a priority and a submission
// time. Submission and every later priority update must go through this, or
// getprocess will report a priority that disagrees with the queue order.
func ComputePriorityTime(priority int, submissionTime time.Time) int64 {
	return int64(priority)*PriorityTimeUnit + submissionTime.UnixNano()
}

// PriorityUpdate is one entry of a bulk priority-channel write.
type PriorityUpdate struct {
	ProcessID string `json:"processid"`
	Priority  int    `json:"priority"`
}

// PriorityUpdateOutcome is why a single update did or did not land.
type PriorityUpdateOutcome string

const (
	PriorityUpdated     PriorityUpdateOutcome = "updated"
	PriorityNotWaiting  PriorityUpdateOutcome = "not_waiting"
	PriorityNotFound    PriorityUpdateOutcome = "not_found"
	PriorityOutOfBounds PriorityUpdateOutcome = "rejected_out_of_bounds"
)

type PriorityUpdateResult struct {
	ProcessID string                `json:"processid"`
	Outcome   PriorityUpdateOutcome `json:"outcome"`
	Priority  int                   `json:"priority"` // effective value after the call
}

func ConvertPriorityUpdateResultsToJSON(results []PriorityUpdateResult) (string, error) {
	jsonBytes, err := json.Marshal(results)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func ConvertJSONToPriorityUpdateResults(jsonString string) ([]PriorityUpdateResult, error) {
	var results []PriorityUpdateResult

	err := json.Unmarshal([]byte(jsonString), &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}
