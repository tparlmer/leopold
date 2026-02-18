package protocol

import (
	"encoding/json"
	"testing"
)

func TestProtocolVersionIsSet(t *testing.T) {
	if ProtocolVersion < 1 {
		t.Errorf("ProtocolVersion = %d, want >= 1", ProtocolVersion)
	}
}

func TestTaskMessageRoundTrip(t *testing.T) {
	original := TaskMessage{
		Type: "task",
		Version: ProtocolVersion,
		ID: "task-001",
		Prompt: "add JWT auth to login endpoint",
		Repo: "/home/thomas/projects/foo",
		Spec: "specs/auth-login.md",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded TaskMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded != original {
		t.Errorf("round trip failed\ngot: %+v\nwant: %+v", decoded, original)
	}
}

func TestCompleteMessageRoundTrip(t *testing.T) {
	original := CompleteMessage{
		Type: "complete",
		Version: ProtocolVersion,
		ID: "task-001",
		State: "done",
		Summary: "added JWT auth",
		FilesChanged: []string{"internal/auth/login.go", "internal/auth/login_test.go"},
		TokensIn: 15000,
		TokensOut: 4800,
		ElapsedS: 262.5,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded CompleteMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Can't use == on structs with slices, compare fields
	if decoded.ID != original.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.State != original.State {
		t.Errorf("State = %q, want %q", decoded.State, original.State)
	}
	if len(decoded.FilesChanged) != len(original.FilesChanged) {
		t.Errorf("FilesChanged len = %d, want %d", len(decoded.FilesChanged), len(original.FilesChanged))
	}
}

func TestHeartbeatMessageRoundTrip(t *testing.T) {
	original := HeartbeatMessage{
		Type: "heartbeat",
		Version: ProtocolVersion,
		ID: "task-001",
		State: "running",
		Tool: "bash",
		Detail: "running go test ./...",
		RSSMB: 42.5,
		TokensIn: 12000,
		TokensOut: 3400,
		ElapsedS: 180.0,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded HeartbeatMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded != original {
		t.Errorf("round trip failed\ngot: %+v\nwant: %+v", decoded, original)
	}
}

func TestBlockedMessageRoundTripWithOptions(t *testing.T) {
	original := BlockedMessage{
		Type: "blocked",
		Version: ProtocolVersion,
		ID: "task-001",
		Question: "Should I add rate limiting?",
		Options: []string{"yes", "no", "yes with 100 req/min limit"},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded BlockedMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Question != original.Question {
		t.Errorf("Question = %q, want %q", decoded.Question, original.Question)
	}
	if len(decoded.Options) != 3 {
		t.Errorf("Options len = %d, want 3", len(decoded.Options))
	}
}

func TestBlockedMessageOmitsOptionsWhenEmpty(t *testing.T) {
	msg := BlockedMessage{
		Type: "blocked",
		Version: ProtocolVersion,
		ID: "task-001",
		Question: "What should I name the package?",
		// Options deliberately left nil
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// The JSON should not contain "options" at all
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	if _, exists := raw["options"]; exists {
		t.Errorf("expected options to be omitted form JSON when nil, but it was present")
	}
}

func TestTaskMessageOmitsSpecWhenEmpty(t *testing.T) {
	msg := TaskMessage{
		Type: "task",
		Version: ProtocolVersion,
		ID: "task-001",
		Prompt: "do the thing",
		Repo: "/home/thomas/foo",
		// Spec deliberately left empty
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	if _, exists := raw["spec"]; exists {
		t.Error("expected spec to be omitted from JSON when empty, but it was present")
	}
}

func TestMalformedJSONReturnsError(t *testing.T) {
	garbage := []string{
		`{"type": "task", broken`,
		`not json at all`,
		``,
		`{{{`,
	}

	for _, input := range garbage {
		t.Run(input, func(t *testing.T) {
			var msg TaskMessage
			if err := json.Unmarshal([]byte(input), &msg); err == nil {
				t.Errorf("expected error for input %q, got nil", input)
			}
		})
	}
}
