package protocol

import (
	"encoding/json"
	"fmt"
)

// envelope is used to peek at the type field before doing a full unmarshal
// We only decode what we need to route to the message -- the rest stays as
// raw JSON until we know which struct to pour it into
type envelope struct {
	Type string `json:"type"`
}

// ParseMessage takes a raw JSON line and returns the appropriate typed
// message struct. The caller uses a type switch to handle each case.
//
// Example:
//
// msg, err := protocol.ParseMessage(line)
// switch m := msg.(type) {
// case *HeartbeatMessage:
// 		// handle heartbeat
// case *CompleteMessage:
// 		// handle completion
// }
func ParseMessage(data []byte) (interface{}, error) {
	var env envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	switch env.Type {
	case "init":
		var msg InitMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("invalid init message: %w", err)
		}
		return &msg, nil

	case "task":
		var msg TaskMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("invalid task message: %w", err)
		}
		return &msg, nil

	case "cancel":
		var msg CancelMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("invalid cancel message: %w", err)
		}
		return &msg, nil

	case "answer":
		var msg AnswerMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("invalid answer message: %w", err)
		}
		return &msg, nil

	case "heartbeat":
		var msg HeartbeatMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("invalid heartbeat message: %w", err)
		}
		return &msg, nil

	case "blocked":
		var msg BlockedMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("invalid blocked message: %w", err)
		}
		return &msg, nil

	case "complete":
		var msg CompleteMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("invalid complete message: %w", err)
		}
		return &msg, nil

	case "":
		return nil, fmt.Errorf("missing message type")

	default:
		return nil, fmt.Errorf("unknown message type: %q", env.Type)
	}
}
