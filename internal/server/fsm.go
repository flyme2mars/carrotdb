package server

import (
	"carrotdb/internal/engine"
	"encoding/json"
	"io"

	"github.com/hashicorp/raft"
)

// Command represents a change operation to be applied to the FSM.
type Command struct {
	Op    string `json:"op"`
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

// FSM is the Raft Finite State Machine for CarrotDB.
type FSM struct {
	engine *engine.Engine
}

func NewFSM(engine *engine.Engine) *FSM {
	return &FSM{engine: engine}
}

// Apply is called by Raft when a log entry is committed.
func (f *FSM) Apply(l *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(l.Data, &cmd); err != nil {
		return err
	}

	switch cmd.Op {
	case "SET":
		return f.engine.Put(cmd.Key, cmd.Value)
	case "DELETE":
		return f.engine.Delete(cmd.Key)
	default:
		return nil
	}
}

// Snapshot returns a snapshot of the current state.
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return &Snapshot{engine: f.engine}, nil
}

// Restore restores the FSM from a snapshot.
func (f *FSM) Restore(r io.ReadCloser) error {
	defer r.Close()
	return f.engine.ReadFrom(r)
}

type Snapshot struct {
	engine *engine.Engine
}

func (s *Snapshot) Persist(sink raft.SnapshotSink) error {
	if err := s.engine.WriteTo(sink); err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *Snapshot) Release() {}
