package postgresql

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/lib/pq"
)

// The priority channel: a bulk, bounded update of the priority of processes that
// are still WAITING.
//
// Priority is stored twice -- as PRIORITY, which is what getprocess reports, and
// inside the derived ordering key PRIORITYTIME, which is what the assign path
// sorts by. Both must move together or the queue order disagrees with the
// reported priority.
//
// PRIORITYTIME is updated as a delta rather than recomputed from
// SUBMISSION_TIME. That keeps the arithmetic exact: SUBMISSION_TIME is a
// TIMESTAMPTZ (microsecond resolution) while the key was built from
// UnixNano, so recomputing would silently shift the key by up to 999 ns and
// could reorder processes submitted within the same microsecond.

type priorityRow struct {
	state    int
	priority int
	floor    int
	ceiling  int
}

func (db *PQDatabase) SetProcessPriorities(updates []core.PriorityUpdate) ([]core.PriorityUpdateResult, error) {
	if len(updates) == 0 {
		return []core.PriorityUpdateResult{}, nil
	}

	// The update joins on process id, so a repeated id would leave both the value
	// that lands and its reported outcome ambiguous.
	seen := make(map[string]struct{}, len(updates))
	for _, update := range updates {
		if update.ProcessID == "" {
			return nil, errors.New("Failed to set process priorities, empty process id in batch")
		}
		if _, duplicate := seen[update.ProcessID]; duplicate {
			return nil, errors.New("Failed to set process priorities, process id " + update.ProcessID + " appears more than once in the batch")
		}
		seen[update.ProcessID] = struct{}{}
	}

	tx, err := db.postgresql.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	landed, err := db.applyPriorityUpdates(tx, updates)
	if err != nil {
		return nil, err
	}

	missing := make([]string, 0, len(updates))
	for _, update := range updates {
		if _, ok := landed[update.ProcessID]; !ok {
			missing = append(missing, update.ProcessID)
		}
	}

	rejected, err := db.selectPriorityRows(tx, missing)
	if err != nil {
		return nil, err
	}

	results := make([]core.PriorityUpdateResult, 0, len(updates))
	for _, update := range updates {
		if priority, ok := landed[update.ProcessID]; ok {
			results = append(results, core.PriorityUpdateResult{ProcessID: update.ProcessID, Outcome: core.PriorityUpdated, Priority: priority})
			continue
		}

		row, found := rejected[update.ProcessID]
		if !found {
			results = append(results, core.PriorityUpdateResult{ProcessID: update.ProcessID, Outcome: core.PriorityNotFound})
			continue
		}
		if row.state != core.WAITING {
			results = append(results, core.PriorityUpdateResult{ProcessID: update.ProcessID, Outcome: core.PriorityNotWaiting, Priority: row.priority})
			continue
		}
		if update.Priority < row.floor || update.Priority > row.ceiling {
			results = append(results, core.PriorityUpdateResult{ProcessID: update.ProcessID, Outcome: core.PriorityOutOfBounds, Priority: row.priority})
			continue
		}

		// The row is WAITING and the write is in bounds, so the update should have
		// taken it. Report the inconsistency rather than inventing an outcome.
		return nil, errors.New("Failed to set process priorities, process " + update.ProcessID + " is WAITING and priority " + strconv.Itoa(update.Priority) + " is within [" + strconv.Itoa(row.floor) + ", " + strconv.Itoa(row.ceiling) + "] but the update did not apply")
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return results, nil
}

// applyPriorityUpdates runs the whole batch as one statement and returns the
// effective priority of every process it actually wrote.
func (db *PQDatabase) applyPriorityUpdates(tx *sql.Tx, updates []core.PriorityUpdate) (map[string]int, error) {
	// $1 is the (negative) priority-time unit, $2 the WAITING state, $3 the
	// default floor; the batch values start at $4.
	args := make([]interface{}, 0, 3+2*len(updates))
	args = append(args, core.PriorityTimeUnit, core.WAITING, core.MinPriority)

	values := make([]string, 0, len(updates))
	for i, update := range updates {
		values = append(values, "($"+strconv.Itoa(4+2*i)+"::text, $"+strconv.Itoa(5+2*i)+"::integer)")
		args = append(args, update.ProcessID, update.Priority)
	}

	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES AS p
	   SET PRIORITY = v.priority,
	       PRIORITYTIME = p.PRIORITYTIME + (v.priority - p.PRIORITY)::bigint * $1::bigint
	  FROM (VALUES ` + strings.Join(values, ", ") + `) AS v(process_id, priority)
	 WHERE p.PROCESS_ID = v.process_id
	   AND p.STATE = $2
	   AND v.priority BETWEEN COALESCE(p.PRIORITY_FLOOR, $3) AND COALESCE(p.PRIORITY_CEILING, p.PRIORITY)
	RETURNING p.PROCESS_ID, p.PRIORITY`

	rows, err := tx.Query(sqlStatement, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	landed := make(map[string]int)
	for rows.Next() {
		var processID string
		var priority int
		if err := rows.Scan(&processID, &priority); err != nil {
			return nil, err
		}
		landed[processID] = priority
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return landed, nil
}

// selectPriorityRows reads the state and effective bounds of the processes the
// update did not write, so each miss can be told apart.
func (db *PQDatabase) selectPriorityRows(tx *sql.Tx, processIDs []string) (map[string]priorityRow, error) {
	found := make(map[string]priorityRow)
	if len(processIDs) == 0 {
		return found, nil
	}

	sqlStatement := `SELECT PROCESS_ID, STATE, PRIORITY, COALESCE(PRIORITY_FLOOR, $1), COALESCE(PRIORITY_CEILING, PRIORITY) FROM ` + db.dbPrefix + `PROCESSES WHERE PROCESS_ID = ANY($2)`
	rows, err := tx.Query(sqlStatement, core.MinPriority, pq.Array(processIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var processID string
		var row priorityRow
		if err := rows.Scan(&processID, &row.state, &row.priority, &row.floor, &row.ceiling); err != nil {
			return nil, err
		}
		found[processID] = row
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return found, nil
}
