package runstore

import "fmt"

// Status is a run lifecycle state. Transitions are validated so a run can never
// jump, for example, from awaiting_user_approval straight to fixing without a
// recorded approval.
type Status string

const (
	StatusScanning             Status = "scanning"
	StatusAwaitingAgent        Status = "awaiting_agent"
	StatusTriaging             Status = "triaging"
	StatusFinalized            Status = "finalized"
	StatusAwaitingUserApproval Status = "awaiting_user_approval"
	StatusFixing               Status = "fixing"
	StatusVerifying            Status = "verifying"
	StatusCompleted            Status = "completed"
	StatusFailed               Status = "failed"
)

// allowedTransitions maps each status to the set of statuses it may move to.
// Any status may move to StatusFailed (handled in canTransition).
var allowedTransitions = map[Status]map[Status]bool{
	StatusScanning:      {StatusAwaitingAgent: true, StatusTriaging: true, StatusFinalized: true},
	StatusAwaitingAgent: {StatusTriaging: true, StatusAwaitingAgent: true, StatusFinalized: true},
	StatusTriaging:      {StatusAwaitingAgent: true, StatusFinalized: true},
	// Finalization produces artifacts; the user then decides whether to fix.
	StatusFinalized:            {StatusAwaitingUserApproval: true, StatusCompleted: true},
	StatusAwaitingUserApproval: {StatusFixing: true, StatusCompleted: true},
	StatusFixing:               {StatusVerifying: true},
	StatusVerifying:            {StatusCompleted: true, StatusAwaitingUserApproval: true},
	StatusCompleted:            {},
	StatusFailed:               {},
}

func isValidStatus(s Status) bool {
	switch s {
	case StatusScanning, StatusAwaitingAgent, StatusTriaging, StatusFinalized,
		StatusAwaitingUserApproval, StatusFixing, StatusVerifying, StatusCompleted, StatusFailed:
		return true
	}
	return false
}

// canTransition reports whether from -> to is a legal move.
func canTransition(from, to Status) error {
	if !isValidStatus(to) {
		return fmt.Errorf("invalid target status %q", to)
	}
	if from == to {
		return nil // idempotent set to the same status
	}
	// Any state may fail.
	if to == StatusFailed {
		return nil
	}
	if allowedTransitions[from][to] {
		return nil
	}
	return fmt.Errorf("illegal status transition %q -> %q", from, to)
}

// isTerminal reports whether a run in this status has finished.
func (s Status) isTerminal() bool {
	return s == StatusCompleted || s == StatusFailed
}

// IsTerminalPublic is the exported terminal-status check used by callers
// outside this package (e.g. the run manager).
func (s Status) IsTerminalPublic() bool { return s.isTerminal() }
