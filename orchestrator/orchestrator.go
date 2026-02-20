package orchestrator

import (
	"time"
)

// Config holds the runtime parameters for the orchestrator.
// In v0.0, tests create this directly. Later, TOML parsing
// will produce a Config.
type Config struct {
	AgentBin string // path to the agent binary
	HeartbeatTimeout time.Duration // kill agent if silent this long
	MaxRSSMB int // RSS budget (0 = unlimited)
}

// Orchestrator supervises a single agent process per RunTask call.
// Stateless between tasks - all per-task state lives inside RunTask
type Orchestrator struct {
	config Config
}

// New creates an orchestrator with the given configuration
func New(cfg Config) *Orchestrator {
	return &Orchestrator{config: cfg}
}
