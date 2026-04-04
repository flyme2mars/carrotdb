package server

import (
	"carrotdb/internal/engine"
	"encoding/json"
	"io"

	"github.com/fatih/color"
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
	Trace  bool
}

func NewFSM(engine *engine.Engine, trace bool) *FSM {
	return &FSM{engine: engine, Trace: trace}
}

// Apply is called by Raft when a log entry is committed.
func (f *FSM) Apply(l *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(l.Data, &cmd); err != nil {
		return err
	}

	if f.Trace {
		color.Cyan("[TRACE] FSM: Committing and applying %s to Engine (Key: %s)", cmd.Op, cmd.Key)
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
	if f.Trace {
		color.Cyan("[TRACE] Raft: Taking snapshot for log compaction")
	}
	return &Snapshot{engine: f.engine, Trace: f.Trace}, nil
}

// Restore restores the FSM from a snapshot.
func (f *FSM) Restore(r io.ReadCloser) error {
	if f.Trace {
		color.Cyan("[TRACE] Raft: Restoring state from snapshot")
	}
	defer r.Close()
	return f.engine.ReadFrom(r)
}

type Snapshot struct {
	engine *engine.Engine
	Trace  bool
}

func (s *Snapshot) Persist(sink raft.SnapshotSink) error {
	if s.Trace {
		color.Cyan("[TRACE] Raft: Persisting snapshot to storage")
	}
	if err := s.engine.WriteTo(sink); err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *Snapshot) Release() {}
