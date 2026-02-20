package orchestrator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// agentBin returns the path to a compiled fake agent binary.
func agentBin(name string) string {
	abs, err := filepath.Abs(filepath.Join("testdata", "bin", name))
	if err != nil {
		panic(err)
	}
	return abs
}

func TestMain(m *testing.M) {
	// Build all fake agents before tests run
	agents := []string{"happy", "hang", "crash", "leak", "garbage"}
	for _, a := range agents {
		cmd := exec.Command("go", "build", "-o",
			filepath.Join("testdata", "bin", a),
			filepath.Join("..", "testdata", "agents", a))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to build %s agent: %v\n", a, err)
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}

func TestOrchestratorCompletesHappyPath(t *testing.T) {
	orch := New(Config{
		AgentBin: agentBin("happy"),
		HeartbeatTimeout: 5 * time.Second,
		MaxRSSMB: 100,
	})

	result, err := orch.RunTask("test-1", "do the thing", t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.State != "done" {
		t.Errorf("state = %q, want %q", result.State, "done")
	}
	if result.Summary != "task completed" {
		t.Errorf("summary = %q, want %q", result.Summary, "task completed")
	}
}

func TestOrchestratorKillsAgentOnHeartbeatTimeout(t *testing.T) {
	orch := New(Config{
		AgentBin: agentBin("hang"),
		HeartbeatTimeout: 500 * time.Millisecond,
		MaxRSSMB: 100,
	})

	_, err := orch.RunTask("test-3", "do the thing", t.TempDir())
	if err == nil {
		t.Fatal("expected error for crashed agent, got nil")
	}
}

func TestOrchestratorHandlesAgentCrash(t *testing.T) {
	orch := New(Config{
		AgentBin: agentBin("crash"),
		HeartbeatTimeout: 5 * time.Second,
		MaxRSSMB: 100,
	})

	_, err := orch.RunTask("test-3", "do the thing", t.TempDir())
	if err == nil {
		t.Fatal("expected error for crashed agent, got nil")
	}
}

func TestOrchestratorKillsAgentExceedingRSS(t *testing.T) {
	orch := New(Config{
		AgentBin: agentBin("leak"),
		HeartbeatTimeout: 10 * time.Second,
		MaxRSSMB: 5, // tiny budget - leak agent exceeds it fast
	})

	_, err := orch.RunTask("test-4", "do the thing", t.TempDir())
	if err == nil {
		t.Fatal("expected error for RSS-exceeded agent, got nil")
	}
}

func TestOrchestratorHandlesMalformedJSON(t *testing.T) {
	orch := New(Config{
		AgentBin: agentBin("garbage"),
		HeartbeatTimeout: 5 * time.Second,
		MaxRSSMB: 100,
	})

	_, err := orch.RunTask("test-5", "do the thing", t.TempDir())
	if err == nil {
		t.Fatal("expected error for garbage-output agent, got nil")
	}
}
