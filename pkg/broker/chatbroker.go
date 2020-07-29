package broker

import (
	"github.com/saied74/toychat/pkg/centerr"
)

//GetDialog returns true if the message is a new dialog and fales if it is not.
//If the dialog is not new, the UserMsg is populated from the dialog table.
func GetDialog(table string, id int) (*TableRow, error) {
	msg := TableRow{ID: id}
	msgs := TableRows{msg}
	exchange := Exchange{
		Table:    table,
		Put:      []string{},
		SpecList: []string{"user_id"},
		Get:      []string{"dialog_id", "user_id", "agent_id", "started", "ended"},
		Tables:   msgs,
		Action:   "get",
	}

	err := exchange.runGetExchange(msgs, exchange.SpecList)
	if err != nil {
		return &TableRow{}, err
	}
	userMsg := exchange.Tables[0] //.Pick()
	// if err != nil {
	// 	return &Dialog{}, err
	// }
	return &userMsg, exchange.DecodeErr()
}

//MakeDialog creates a new entry in the dialog table and returns the dialog_id
func MakeDialog(table string, id, agentID int) error {
	msg := TableRow{ID: id, AgentID: agentID}
	msgs := TableRows{msg}
	exchange := Exchange{
		Table:  table,
		Put:    []string{"user_id", "agent_id", "started"},
		Tables: msgs,
		Action: "insert",
	}
	for _, m := range msgs {
		c := m.BuildInsert(exchange.Put)
		exchange.Spec = append(exchange.Spec, c)
	}
	return exchange.runExchange()
}

//AddAgentToDialog adds the agent ID to the dialog table
func AddAgentToDialog(dialogID, agentID int) error {
	msg := TableRow{DialogID: dialogID, AgentID: agentID}
	exchange := Exchange{
		Table:    "dialogs",
		Put:      []string{"agent_id"},
		SpecList: []string{"dialog_id"},
		Get:      []string{},
		Tables:   TableRows{msg},
		Action:   "put",
	}
	exchange.Spec = append(exchange.Spec, msg.Specify(exchange.Put,
		exchange.SpecList))
	err := exchange.runExchange()
	if err != nil {
		return err
	}
	return nil
}

//SelectAgent executes a transaction on the database and selects an agent
//and returns the agent ID.  No agent available is shown in the error.
func SelectAgent() (agentID int, err error) {
	exchange := Exchange{
		Table:  "dialogs",
		Action: "agent",
	}
	err = exchange.runExchange()
	if err != nil {
		return 0, err
	}
	userMsgs := exchange.Tables
	userMsg := userMsgs[0]
	return userMsg.AgentID, err
}

//EnterMsg adds the next messsage into the message table
func EnterMsg(table string, dialogID int, message string) error {
	msg := TableRow{DialogID: dialogID, Msg: message}
	msgs := TableRows{msg}
	exchange := Exchange{
		Table:  table,
		Put:    []string{"dialog_id", "created", "message"},
		Tables: msgs,
		Action: "insert",
	}

	for _, m := range msgs {
		c := m.BuildInsert(exchange.Put)
		exchange.Spec = append(exchange.Spec, c)
	}
	return exchange.runExchange()

}

//MessageAgent sends a message to the agent and gets the reply
func MessageAgent(agentID, userID int, message string) (string, error) {
	centerr.InfoLog.Printf("and the message is: %s", message)
	return "", nil
}
