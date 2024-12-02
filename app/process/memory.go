package process

import (
	"base/actions"
	"encoding/json"
	"os"
	"time"
)

type AccountState struct {
	AccountID         int              `json:"account_id"`
	CompletedActions  []actions.Action `json:"completed_actions"`
	LastProcessedTime time.Time        `json:"last_processed_time"`
	LastActionTime    time.Time        `json:"last_action_time"`
	TotalElapsedTime  time.Duration    `json:"total_elapsed_time"`
	ActionIntervals   []time.Duration  `json:"action_intervals"`

	GeneratedActions   []actions.Action `json:"generated_actions"`
	GeneratedDuration  time.Duration    `json:"generated_duration"`
	GeneratedIntervals []time.Duration  `json:"generated_intervals"`
}

type Memory struct {
	StateFilePath string
}

func NewMemory(stateFilePath string) *Memory {
	return &Memory{StateFilePath: stateFilePath}
}

func (m *Memory) SaveState(state *AccountState) error {
	file, err := os.OpenFile(m.StateFilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	var states []AccountState
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&states); err != nil && err.Error() != "EOF" {
		return err
	}

	var found bool
	for i, existingState := range states {
		if existingState.AccountID == state.AccountID {
			states[i] = *state
			found = true
			break
		}
	}

	if !found {
		states = append(states, *state)
	}

	file.Seek(0, 0)
	file.Truncate(0)
	encoder := json.NewEncoder(file)
	return encoder.Encode(states)
}

func (m *Memory) LoadState(accountID int) (*AccountState, error) {
	file, err := os.Open(m.StateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var states []AccountState
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&states)
	if err != nil && err.Error() != "EOF" {
		return nil, err
	}

	for _, state := range states {
		if state.AccountID == accountID {
			return &state, nil
		}
	}
	return nil, nil
}

func (m *Memory) IsStateFileNotEmpty() (bool, error) {
	file, err := os.Open(m.StateFilePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	var states []AccountState
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&states)
	if err != nil && err.Error() != "EOF" {
		return false, err
	}

	return len(states) > 0, nil
}

func (m *Memory) UpdateState(accountID int, completedAction actions.Action, interval time.Duration) error {
	state, err := m.LoadState(accountID)
	if err != nil {
		return err
	}

	if state == nil {
		state = &AccountState{AccountID: accountID}
	}

	state.CompletedActions = append(state.CompletedActions, completedAction)
	state.TotalElapsedTime += interval
	state.LastProcessedTime = time.Now()
	state.ActionIntervals = append(state.ActionIntervals, interval)
	state.LastActionTime = time.Now()

	return m.SaveState(state)
}

func (m *Memory) ClearState() error {
	return os.Truncate(m.StateFilePath, 0)
}
