package orchestrator

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"
	"bufio"

	"github.com/tparlmer/leopold/protocol"
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

// msgResult carries a parsed message (or error) from the reader goroutine
// to the control loop.
type msgResult struct {
	msg interface{}
	err error
}

// New creates an orchestrator with the given configuration
func New(cfg Config) *Orchestrator {
	return &Orchestrator{config: cfg}
}

// sendMessage marshals a protocol message and writes it as a JSON line.
// used to send InitMessage and TaskMessage to the agent's stdin.
func sendMessage(w io.Writer, msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	// need to append newline because JSON protocol requires one JSON object per line
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}

// startREader launches a goroutine that reads JSON lines from the agent's
// stdout, parses each line via ParseMessage, and sends results to the
// returned channel. The channel is closed when the pipe closes or errors.
func startReader(stdout io.Reader) <-chan msgResult {
	ch := make(chan msgResult)
	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			msg, err := protocol.ParseMessage(scanner.Bytes())
			if err != nil {
				ch <- msgResult{err: fmt.Errorf("parse: %w", err)}
				continue
			}
			ch <- msgResult{msg: msg}
		}
		// scanner.Err() is nil on celan EOF (pipe closed).
		// Non-nil means an I/O error - but either way, we're done
		if err := scanner.Err(); err != nil {
			ch <- msgResult{err: fmt.Errorf("read stdout: %w", err)}
		}
	}()
	return ch
}

// RunTask spawns an agent, sends it a task, and supervises it to completion.
// Returns the agent's CompleteMessage on success, or an error fi the agent
// misbehaved (timeout, crash, RSS exceeded, etc.).
func (o *Orchestrator) RunTask(taskID, prompt, repo string) (*protocol.CompleteMessage, error) {
	// --- Phase 1: Spawn the process and wire pipes ---
	cmd := exec.Command(o.config.AgentBin)
	// Sets the agen'ts working directory
	cmd.Dir = repo

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdin pipe: %w", err)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start agent: %w", err)
	}

	// Ensure cleanup: if we return early for any reason, kill the process.
	// This is the safety net - specific paths may kill it earlier.
	// NOTE - defer runs in reverse order - Wait then Kill
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	// --- Phase 2: Send init + task messages ---
	init := protocol.InitMessage{
		Type: "init",
		Version: protocol.ProtocolVersion,
		HeartbeatIntervalS: int(o.config.HeartbeatTimeout.Seconds()) / 2,
	}
	if err := sendMessage(stdinPipe, init); err != nil {
		return nil, fmt.Errorf("send init: %w", err)
	}

	task := protocol.TaskMessage{
		Type: "task",
		Version: protocol.ProtocolVersion,
		ID: taskID,
		Prompt: prompt,
		Repo: repo,
	}
	if err := sendMessage(stdinPipe, task); err != nil {
		return nil, fmt.Errorf("send task: %w", err)
	}

	// --- Phase 3: Monitor ---

	// Track process exit in background
	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	msgCh := startReader(stdoutPipe)

	heartbeat := time.NewTimer(o.config.HeartbeatTimeout)
	defer heartbeat.Stop()

	for {
		select {
		case result, ok := <-msgCh:
			// Channel closed - reader goroutine exited
			if !ok {
				// Wait for process to exit so we can report the exit error
				err := <-waitCh
				if err != nil {
					return nil, fmt.Errorf("agent exited unexpectedly: %w", err)
				}
				return nil, fmt.Errorf("agent exited without completing")
			}

			// Parse error - agent sent garbage
			if result.err != nil {
				cmd.Process.Kill()
				<-waitCh
				return nil, fmt.Errorf("agent protocol error: %w ", err)
			}

			// Valid message - agent is alive, reset the watchdog
			heartbeat.Reset(o.config.HeartbeatTimeout)

			// Handle by type
			switch msg := result.msg.(type) {
			case *protocol.HeartbeatMessage:
				// Check RSS budget
				if o.config.MaxRSSMB > 0 && int(msg.RSSMB) > o.config.MaxRSSMB {
					cmd.Process.Kill()
					<-waitCh
					return nil, fmt.Errorf(
						"agent exceeeded RSS limit: %d MB > %d MB",
						int(msg.RSSMB), o.config.MaxRSSMB,
					)
				}
				// Otherwise: agent is a live and within budget, continue

			case *protocol.BlockedMessage:
				// v0.0: nobody's home to answer. Kill and report
				cmd.Process.Kill()
				<-waitCh
				return nil, fmt.Errorf(
					"agent blocked with question: %s", msg.Question,
				)

			case *protocol.CompleteMessage:
				// Happy path - agent finished its task
				<-waitCh
				return msg, nil

			default:
				// Unknown message type from a well-parsed message.
				// Shouldn't happen, but don't crash - log and continue.
			}

		case <-heartbeat.C:
			// Agent went silent. Kill it.
			cmd.Process.Kill()
			<-waitCh
			return nil, fmt.Errorf(
				"agent heartbeat timeout after %s", o.config.HeartbeatTimeout,
			)

		case err := <-waitCh:
			// Process exited before sending CompleteMessage
			if err != nil {
				return nil, fmt.Errorf("agent crashed: %w", err)
			}
			return nil, fmt.Errorf("agent exited without completing")
		}
	}

	return nil, fmt.Errorf("not yet implemented")
}
