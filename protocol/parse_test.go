package protocol

import (
	"testing"
)

func TestParseMessageDispatchesCorrectType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType string // we'll check this with a type switch
	}{
		{
			"init message",
			`{"type":"init","v":1,"heartbeat_interval_s":30,"max_tokens":50000}`,
			"init",
		},
		{
			"task message",
			`{"type":"task","v":1,"id":"t1","prompt":"do it","repo":"/tmp"}`,
			"task",
		},
		{
			"cancel message",
			`{"type":"cancel","v":1,"id":"t1","reason":"timeout"}`,
			"cancel",
		},
		{
			"answer message",
			`{"type":"answer","v":1,"id":"t1","response":"yes"}`,
			"answer",
		},
		{
			"heartbeat message",
			`{"type":"heartbeat","v":1,"id":"t1","state":"running","tool":"bash","detail":"compiling","rss_mb":42,"tokens_in":100,"tokens_out":50,"elapsed_s":10}`,
			"heartbeat",
		},
		{
			"blocked message",
			`{"type":"blocked","v":1,"id":"t1","question":"should I?"}`,
			"blocked",
		},
		{
			"complete message done",
			`{"type":"complete","v":1,"id":"t1","state":"done","summary":"finished","tokens_in":100,"tokens_out":50,"elapsed_s":10}`,
			"complete",
		},
		{
			"complete message failed",
			`{"type":"complete","v":1,"id":"t1","state":"failed","error":"out of tokens","tokens_in":100,"tokens_out":50,"elapsed_s":10}`,
			"complete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := ParseMessage([]byte(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := typeOf(msg)
			if got != tt.wantType {
				t.Errorf("got type %q, want %q", got, tt.wantType)
			}
		})
	}
}

func TestParseMessagePreservesFields(t *testing.T) {
	t.Run("task fields survive round trip", func(t *testing.T) {
		input := `{"type":"task","v":1,"id":"task-99","prompt":"add auth","repo":"/home/thomas/foo","spec":"specs/auth.md"}`

		msg, err := ParseMessage([]byte(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		task, ok := msg.(*TaskMessage)
		if !ok {
			t.Fatalf("expected *TaskMessage, got %T", msg)
		}

		if task.ID != "task-99" {
			t.Errorf("ID = %q, want %q", task.ID, "task-99")
		}
		if task.Prompt != "add auth" {
			t.Errorf("Prompt = %q, want %q", task.Prompt, "add auth")
		}
		if task.Repo != "/home/thomas/foo" {
			t.Errorf("Repo = %q, want %q", task.Repo, "/home/thomas/foo")
		}
		if task.Spec != "specs/auth.md" {
			t.Errorf("Spec = %q, want %q", task.Spec, "specs/auth.md")
		}
		if task.Version != 1 {
			t.Errorf("Version = %d, want 1", task.Version)
		}
	})

	t.Run("heartbeat fields survive round trip", func(t *testing.T) {
		input := `{"type":"heartbeat","v":1,"id":"t1","state":"running","tool":"bash","detail":"testing","rss_mb":42.5,"tokens_in":8000,"tokens_out":2000,"elapsed_s":120.5}`

		msg, err := ParseMessage([]byte(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		hb, ok := msg.(*HeartbeatMessage)
		if !ok {
			t.Fatalf("expected *HeartbeatMessage, got %T", msg)
		}

		if hb.RSSMB != 42.5 {
			t.Errorf("RSSMB = %f, want 42.5", hb.RSSMB)
		}
		if hb.Tool != "bash" {
			t.Errorf("Tool = %q, want %q", hb.Tool, "bash")
		}
		if hb.ElapsedS != 120.5 {
			t.Errorf("ElapsedS = %f, want 120.5", hb.ElapsedS)
		}
	})

	t.Run("blocked with options", func(t *testing.T) {
		input := `{"type":"blocked","v":1,"id":"t1","question":"which approach?","options":["A","B","C"]}`

		msg, err := ParseMessage([]byte(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		blocked, ok := msg.(*BlockedMessage)
		if !ok {
			t.Fatalf("expected *BlockedMessage, got %T", msg)
		}

		if len(blocked.Options) != 3 {
			t.Errorf("Options len = %d, want 3", len(blocked.Options))
		}
		if blocked.Options[0] != "A" {
			t.Errorf("Options[0] = %q, want %q", blocked.Options[0], "A")
		}
	})

	t.Run("complete with files changed", func(t *testing.T) {
		input := `{"type":"complete","v":1,"id":"t1","state":"done","summary":"done","files_changed":["a.go","b.go"],"tokens_in":100,"tokens_out":50,"elapsed_s":10}`

		msg, err := ParseMessage([]byte(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		complete, ok := msg.(*CompleteMessage)
		if !ok {
			t.Fatalf("expected *CompleteMessage, got %T", msg)
		}

		if len(complete.FilesChanged) != 2 {
			t.Errorf("FilesChanged len = %d, want 2", len(complete.FilesChanged))
		}
		if complete.State != "done" {
			t.Errorf("State = %q, want %q", complete.State, "done")
		}
	})
}

func TestParseMessageRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"garbage", `not json at all`},
		{"broken json", `{"type": "task", broken`},
		{"empty string", ``},
		{"empty object", `{}`},
		{"missing type", `{"id":"t1","prompt":"hello"}`},
		{"unknown type", `{"type":"explode","id":"t1"}`},
		{"type is number", `{"type":42}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseMessage([]byte(tt.input))
			if err == nil {
				t.Errorf("expected error for input %q, got nil", tt.input)
			}
		})
	}
}

// TestParseMessageIgnoresUnknownFields verifies that extra fields in the
// JSON don't cause parse errors. This is important for forward compatibility —
// a newer agent might send fields that an older Leopold doesn't know about.
func TestParseMessageIgnoresUnknownFields(t *testing.T) {
	input := `{"type":"heartbeat","v":1,"id":"t1","state":"running","tool":"bash","detail":"ok","rss_mb":10,"tokens_in":1,"tokens_out":1,"elapsed_s":1,"some_future_field":"surprise"}`

	msg, err := ParseMessage([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error on message with unknown field: %v", err)
	}

	if _, ok := msg.(*HeartbeatMessage); !ok {
		t.Errorf("expected *HeartbeatMessage, got %T", msg)
	}
}

// typeOf is a test helper that returns the message type string for
// assertion purposes. Keeps the test table clean.
func typeOf(msg interface{}) string {
	switch msg.(type) {
	case *InitMessage:
		return "init"
	case *TaskMessage:
		return "task"
	case *CancelMessage:
		return "cancel"
	case *AnswerMessage:
		return "answer"
	case *HeartbeatMessage:
		return "heartbeat"
	case *BlockedMessage:
		return "blocked"
	case *CompleteMessage:
		return "complete"
	default:
		return "unknown"
	}
}
