package command

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/docker-slim/docker-slim/pkg/app/master/config"
)

// Message errors
var (
	ErrUnknownMessage = errors.New("unknown command type")
)

const (
	ResponseStatusOk    = "ok"
	ResponseStatusError = "error"
)

// Response contains the command response status information
type Response struct {
	Status string `json:"status"`
}

// MessageName is a message ID type
type MessageName string

// Supported messages
const (
	StartMonitorName   MessageName = "cmd.monitor.start"
	StopMonitorName    MessageName = "cmd.monitor.stop"
	ShutdownSensorName MessageName = "cmd.sensor.shutdown"
)

// Message represents the message interface
type Message interface {
	GetName() MessageName
}

// StartMonitor contains the start monitor command fields
type StartMonitor struct {
	RTASourcePT     bool     `json:"rta_source_ptrace"`
	AppName         string   `json:"app_name"`
	AppArgs         []string `json:"app_args,omitempty"`
	AppUser         string   `json:"app_user,omitempty"`
	RunTargetAsUser bool     `json:"run_tas_user,omitempty"`
	KeepPerms       bool     `json:"keep_perms,omitempty"`

	*config.FileMatcherConfig `json:",inline"`
}

// GetName returns the command message ID for the start monitor command
func (m *StartMonitor) GetName() MessageName {
	return StartMonitorName
}

// StopMonitor contains the stop monitor command fields
type StopMonitor struct {
}

// GetName returns the command message ID for the stop monitor command
func (m *StopMonitor) GetName() MessageName {
	return StopMonitorName
}

// ShutdownSensor contains the 'shutdown sensor' command fields
type ShutdownSensor struct{}

// GetName returns the command message ID for the 'shutdown sensor' command
func (m *ShutdownSensor) GetName() MessageName {
	return ShutdownSensorName
}

type messageWrapper struct {
	Name MessageName     `json:"name"`
	Data json.RawMessage `json:"data,omitempty"`
}

// Encode encodes the message instance to a JSON buffer object
func Encode(m Message) ([]byte, error) {
	obj := messageWrapper{
		Name: m.GetName(),
	}

	switch v := m.(type) {
	case *StartMonitor:
		var b bytes.Buffer
		encoder := json.NewEncoder(&b)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(v); err != nil {
			return nil, err
		}

		obj.Data = b.Bytes()
	case *StopMonitor:
	case *ShutdownSensor:
	default:
		return nil, ErrUnknownMessage
	}

	return json.Marshal(&obj)
}

// Decode decodes JSON data into a message instance
func Decode(data []byte) (Message, error) {
	var wrapper messageWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}

	switch wrapper.Name {
	case StartMonitorName:
		var cmd StartMonitor
		if err := json.Unmarshal(wrapper.Data, &cmd); err != nil {
			return nil, err
		}

		return &cmd, nil
	case StopMonitorName:
		return &StopMonitor{}, nil
	case ShutdownSensorName:
		return &ShutdownSensor{}, nil
	default:
		return nil, ErrUnknownMessage
	}
}
