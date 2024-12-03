package process

import (
	"base/actions"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
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
	mu            sync.Mutex
}

func NewMemory(stateFilePath string) *Memory {
	return &Memory{StateFilePath: stateFilePath}
}

func (m *Memory) SaveState(state *AccountState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

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
	m.mu.Lock()
	defer m.mu.Unlock()

	_, err := os.Stat(m.StateFilePath)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("ошибка доступа к файлу состояния: %v", err)
	}

	file, err := os.Open(m.StateFilePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла состояния: %v", err)
	}
	defer file.Close()

	var states []AccountState
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&states)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("ошибка декодирования файла состояния: %v", err)
	}

	for _, state := range states {
		if state.AccountID == accountID {
			return &state, nil
		}
	}
	return nil, nil
}

func (m *Memory) IsStateFileNotEmpty() (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, err := os.Stat(m.StateFilePath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("ошибка доступа к файлу состояния: %v", err)
	}

	file, err := os.Open(m.StateFilePath)
	if err != nil {
		return false, fmt.Errorf("ошибка открытия файла состояния: %v", err)
	}
	defer file.Close()

	var states []AccountState
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&states)
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("ошибка декодирования файла состояния: %v", err)
	}

	return len(states) > 0, nil
}

func (m *Memory) UpdateState(accountID int, completedAction actions.Action, interval time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.loadStateWithoutLock(accountID)
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

	return m.saveStateWithoutLock(state)
}

func (m *Memory) loadStateWithoutLock(accountID int) (*AccountState, error) {
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

func (m *Memory) saveStateWithoutLock(state *AccountState) error {
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

func (m *Memory) ClearState(accountID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
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

	var updatedStates []AccountState
	for _, state := range states {
		if state.AccountID != accountID {
			updatedStates = append(updatedStates, state)
		}
	}

	file.Seek(0, 0)
	file.Truncate(0)
	encoder := json.NewEncoder(file)
	return encoder.Encode(updatedStates)
}

func (m *Memory) ClearAllStates() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return os.Truncate(m.StateFilePath, 0)
}
