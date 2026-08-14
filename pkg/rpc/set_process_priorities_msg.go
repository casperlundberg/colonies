package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const SetProcessPrioritiesPayloadType = "setprocessprioritiesmsg"

// SetProcessPrioritiesMsg is the priority channel: one bulk, bounded update of
// the priority of WAITING processes in a colony.
//
// The batch is deliberately colony-scoped rather than free-form. The server
// checks membership of one colony per call, so a batch able to reach outside it
// would make that check meaningless.
type SetProcessPrioritiesMsg struct {
	MsgType    string                `json:"msgtype"`
	ColonyName string                `json:"colonyname"`
	Updates    []core.PriorityUpdate `json:"updates"`
}

func CreateSetProcessPrioritiesMsg(colonyName string, updates []core.PriorityUpdate) *SetProcessPrioritiesMsg {
	msg := &SetProcessPrioritiesMsg{}
	msg.MsgType = SetProcessPrioritiesPayloadType
	msg.ColonyName = colonyName
	msg.Updates = updates

	return msg
}

func (msg *SetProcessPrioritiesMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *SetProcessPrioritiesMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *SetProcessPrioritiesMsg) Equals(msg2 *SetProcessPrioritiesMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType != msg2.MsgType || msg.ColonyName != msg2.ColonyName {
		return false
	}

	if len(msg.Updates) != len(msg2.Updates) {
		return false
	}

	// Order is part of the message: the reply reports one outcome per update, in
	// the order given.
	for i, update := range msg.Updates {
		if update != msg2.Updates[i] {
			return false
		}
	}

	return true
}

func CreateSetProcessPrioritiesMsgFromJSON(jsonString string) (*SetProcessPrioritiesMsg, error) {
	var msg *SetProcessPrioritiesMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
