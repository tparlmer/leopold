package protocol

// Protocol version. Included in every message so both sides can
// detect incompatibilities early. Bump this when the message shapes
// change in a breaking way.
const ProtocolVersion = 1


// --- Orchestrator -> Agent messages ---

// InitMessageis the first message sent to an agent after spawn.
// Tells the agent about its resource budget and behavioral constraints
// bfore any task is assigned.
//
// Why separate from Taskmessage: the agent may need to configure itself
// (set token limits, heartbeat interval) before it knows what work to do.
// Keeps "who you are" separate from "what to do"
type InitMessage struct {
	Type string `json:"type"` // always "init"
	Version int `json:"v"` // protocol version
	HeartbeatIntervalS int `json:"heartbeat_interval_s"` // how often agent should send status/metric
	MaxTokens int `json:"max_tokens,omitempty"` // token budget (0 = unlimited)
	// TODO: max_duration_s - optional wall-clock budget per task
	// TODO: model - preferred model to use (agent can ignore, but orchestrator can suggest)
	// TODO: env - key/value pairs fo ragent specific environment config
}

// TaskMessage assigns work. One task per agent lifetime - the agent
// processes this, sends complete, and exits.
type TaskMessage struct {
	Type string `json:"type"` // always "task"
	Version int `json:"v"` // protocol version
	ID string `json:"id"` // unique task identifier
	Prompt string `json:"prompt"` // what the agent should do
	Repo string `json:"repo"` // working directory
	Spec string `json:"spec,omitempty` // optional path to a spec file
	// TODO: context - prior failure output for test-driven supervision retries
	// TODO: files - list of files the agent should focus on (narrows scope)
	// TODO: depends_on - IDs of tasks that must complete first (for futur task graph)
}

// CancelMessage requests graceful shutdown of the current task.
// the agent should wrap up, send a CompleteMessage with state "cancelled",
// and exit
type CancelMessage struct {
	Type string `json:"type"` // always "cancel"
	Version int `json:"v"` // protocol version
	ID string `json:"id"`
	Reason string `json:"reason"`
}

// AnswerMessage replies toa BlockedMessage from the agent.
type AnswerMessage struct {
	Type string `json:"type"` // always "answer"
	Version int `json:"v"` // protocol version
	ID string `json:"id"`
	Response string `json:"response"`
}

// --- Agent -> Orchestrator messages ---

// Heartbeat message is the agent's periodic "I'm alive and here's what
// I'm doing signal. Combines progress info and resource usage into one message
//
// Design decision: the original design ahd separate "status" and "metric"
// messages. Merged them because in practice you always want both -- knowing
// an agent is "running bash" isn't useful without knowing its RSS, and
// knowing RSS without knowing what it's doing isn't useful either.
// One message, one heartbeat interval, simpler protocol.
type HeartbeatMessage struct {
	Type string `json:"type"` // always "heartbeat"
	Version int `json:"v"` // protocol version
	ID string `json:"id"`
	State string `json:"state"` // "running"
	Tool string `json:"tool"` // what tool is currently active
	Detail string `json:"detail"`
	RSSMB float64 `json:"rss_mb`
	TokensIn int `json:"tokens_in"`
	TokensOut int `json:"tokens_out`
	ElapsedS float64 `json:"elapsed_s"`
	// TODO: cost_usd - agent-reported cost so far (agent knows the model + pricing)
	// TODO: files_touched - files modified since last heartbeat (enables live monitoring)
}

// BlockedMessage signals that the agents needs human input to proceed.
// The orchestrator decides the policy: forward to a human via push
// notification, auto-reply, or cancel the task.
type BlockedMessage struct {
	Type string `json:"type"` // always "blocked"
	Version int `json:"v"` // protocol version
	ID string `json:"id"`
	Question string `json:"question"` // human-readable question
	Options []string `json:"options,omitempty"` // structured choices, if applicable
	// TODO: default_option - index into Options for auto-answer after timeout
	// TODO: timeout_s - how long the agent is willing to wait before giving up
}

// Completemessage is the terminal message. The agent must exit after
// sending this. No further messages should be sent.
type CompleteMessage struct {
	Type string `json:"type"` // always "complete"
	Version int `json:"v"` // protocol version
	ID string `json:"id"`
	State string `json:"state"`
	Summary string `json:"summary,omitempty` // on success
	Error string `json:"error,omitempty` // on failure
	FilesChanged []string `json:"files_changed,omitempty"`
	TokensIn int `json:"tokens_in"` // final totals
	TokensOut int `json:"tokens_out`
	ElapsedS float64 `json:"elapsed_s`
	// TODO: cost_usd - total cost for the task
	// TODO: exit_code - agent's self-reported exit status (separate from OS exit code)
}
