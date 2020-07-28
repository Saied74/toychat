package broker

import (
	"fmt"
	"time"

	"github.com/saied74/toychat/pkg/centerr"
)

//Dialog is for exchange of agent message beween the back end the hub
type Dialog struct {
	ID        int //user ID
	DialogID  int
	AgentID   int
	MessageID int
	Dialog    int
	Name      string
	Email     string
	Started   time.Time
	Ended     time.Time
	Open      bool
	Msg       string
	Action    string //hello,
	ErrType   errMsg
	Err       string
}

//Dialogs allows for the possibility of getting back multiple messages
type Dialogs []Dialog

//Length is dummy function to meet TableProxy interface contract
func (m Dialogs) Length() int {
	return len(m)
}

//PickZero is a dummy function to meet TableProxy interface contract
func (m Dialogs) PickZero() (*Person, error) {
	return nil, nil
}

//Pick picks the user message out of the single entry list
func (m Dialogs) Pick() (*Dialog, error) {
	if m.Length() == 1 {
		return &m[0], nil
	}
	return &Dialog{}, fmt.Errorf("pick: incorrect length %d", m.Length())
}

//AddToDialog provides the append functon for the interface
func (m Dialogs) AddToDialog(d Dialog) Dialogs {
	m = append(m, d)
	return m
}

//AddToPeople is a dummy method to satisfy the interface
func (m Dialogs) AddToPeople(p Person) People {
	return People{}
}

//GetItems for Dialogs
func (m Dialog) GetItems(get []string) []interface{} {
	var g = []interface{}{}
	for _, sp := range get {
		switch sp {
		case "user_id":
			g = append(g, &m.ID)
		case "dialog_id":
			g = append(g, &m.DialogID)
		case "agent_id":
			g = append(g, &m.AgentID)
		case "messsage_id":
			g = append(g, &m.MessageID)
		case Name:
			g = append(g, &m.Name)
		case Email:
			g = append(g, &m.Email)
		case "started":
			// continue
			g = append(g, &m.Started)
		case "ended":
			g = append(g, &m.Ended)
		case "message":
			g = append(g, &m.Msg)
		}
	}
	return g
}

//GetSpec builds the list of items to get from the db table
func (m *Dialog) GetSpec(spec []string) []interface{} {
	var g = []interface{}{}
	for _, sp := range spec {
		switch sp {
		case "user_id":
			g = append(g, m.ID)
		case DialogID:
			g = append(g, m.DialogID)
		case AgentID:
			g = append(g, m.AgentID)
		case Started:
			continue
			// g = append(g, m.Started)
		case Ended:
			g = append(g, m.Ended)
		case Open:
			g = append(g, m.Open)
		case "message":
			g = append(g, m.Msg)
		}
	}
	return g
}

//Specify creates the []interface{} for meeting UDATE and WHERE clauses of the
//SQL EXECUTE statement
func (m *Dialog) Specify(put, spec []string) []interface{} {
	sp := []interface{}{}
	for _, pu := range put {
		switch pu {
		case iD:
			sp = append(sp, m.ID)
		case DialogID:
			sp = append(sp, m.DialogID)
		case AgentID:
			sp = append(sp, m.AgentID)
		case Name:
			sp = append(sp, m.Name)
		case Email:
			sp = append(sp, m.Email)
		case Started:
			sp = append(sp, m.Started)
		case Ended:
			sp = append(sp, m.Ended)
		case Open:
			sp = append(sp, m.Open)
		case Message:
			sp = append(sp, m.Msg)
		}
	}
	for _, s := range spec {
		switch s {
		case iD:
			sp = append(sp, m.ID)
		case DialogID:
			sp = append(sp, m.DialogID)
		case AgentID:
			sp = append(sp, m.AgentID)
		case Name:
			sp = append(sp, m.Name)
		case Email:
			sp = append(sp, m.Email)
		case Started:
			sp = append(sp, m.Started)
		case Ended:
			sp = append(sp, m.Ended)
		case Open:
			sp = append(sp, m.Open)
		case Message:
			sp = append(sp, m.Msg)
		}
	}
	return sp
}

//GetBack reverses the get item and takes the interface items and gets the
//underlying data back.
func (m *Dialog) GetBack(get []string, g []interface{}) error {
	for i, sp := range get {
		switch sp {
		case "dialog_id":
			xDialogID, ok := g[i].(*int)
			if !ok {
				return fmt.Errorf("DialogID (int) type assertion failed")
			}
			m.DialogID = *xDialogID
		case "user_id":
			xUserID, ok := g[i].(*int)
			if !ok {
				return fmt.Errorf("UserID (int) type assertion failed")
			}
			m.ID = *xUserID
		case "agent_id":
			xAgentID, ok := g[i].(*int)
			if !ok {
				return fmt.Errorf("AgentID (int) type assertion failed")
			}
			m.AgentID = *xAgentID
		case "message_id":
			xMessageID, ok := g[i].(*int)
			if !ok {
				return fmt.Errorf("MessageID (int) type assertion failed")
			}
			m.MessageID = *xMessageID
		case "name":
			xName, ok := g[i].(*string)
			if !ok {
				return fmt.Errorf("Name (string) type assertion failed")
			}
			m.Name = *xName
		case "email":
			xEmail, ok := g[i].(*string)
			if !ok {
				return fmt.Errorf("Email (string) type assertion failed")
			}
			m.Email = *xEmail
		case "started":
			xStarted, ok := g[i].(*time.Time)
			if !ok {
				return fmt.Errorf("Started (time.Time) type assertion failed")
			}
			m.Started = *xStarted
		case "ended":
			xEnded, ok := g[i].(*time.Time)
			if !ok {
				return fmt.Errorf("Ended (time.Time) type assertion failed")
			}
			m.Ended = *xEnded
		case "message":
			xMsg, ok := g[i].(*string)
			if !ok {
				return fmt.Errorf("Msg (string) type assertion failed")
			}
			m.Msg = *xMsg
		case "action":
			xAction, ok := g[i].(*string)
			if !ok {
				return fmt.Errorf("Action(string) type assertion failed")
			}
			m.Action = *xAction
		}
	}
	return nil
}

func (e *Exchange) runDialogExchange(dialogs Dialogs, getSpec []string) error {
	for _, d := range dialogs {
		c := d.GetSpec(getSpec)
		e.Spec = append(e.Spec, c)
	}
	err := e.runExchange()
	if err != nil {
		return err
	}
	return e.DecodeErr()
}

//GetDialog returns true if the message is a new dialog and fales if it is not.
//If the dialog is not new, the UserMsg is populated from the dialog table.
func GetDialog(table string, id int) (*Dialog, error) {
	msg := Dialog{ID: id}
	msgs := Dialogs{msg}
	exchange := Exchange{
		Table:    table,
		Put:      []string{},
		SpecList: []string{"user_id"},
		Get:      []string{"dialog_id", "user_id", "agent_id", "started", "ended"},
		People:   msgs,
		Action:   "getDialog",
	}

	err := exchange.runDialogExchange(msgs, exchange.SpecList)
	if err != nil {
		return &Dialog{}, err
	}
	userMsg, err := exchange.People.Pick()
	if err != nil {
		return &Dialog{}, err
	}
	return userMsg, exchange.DecodeErr()
}

//MakeDialog creates a new entry in the dialog table and returns the dialog_id
func MakeDialog(table string, id, agentID int) error {
	msg := Dialog{ID: id, AgentID: agentID}
	msgs := Dialogs{msg}
	exchange := Exchange{
		Table:  table,
		Put:    []string{"user_id", "agent_id", "started"},
		People: msgs,
		Action: "insert",
	}
	for _, m := range msgs {
		c := m.GetSpec(exchange.Put)
		exchange.Spec = append(exchange.Spec, c)
	}
	return exchange.runExchange()
}

//AddAgentToDialog adds the agent ID to the dialog table
func AddAgentToDialog(dialogID, agentID int) error {
	msg := Dialog{DialogID: dialogID, AgentID: agentID}
	exchange := Exchange{
		Table:    "dialogs",
		Put:      []string{"agent_id"},
		SpecList: []string{"dialog_id"},
		Get:      []string{},
		People:   Dialogs{msg},
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
	userMsgs := exchange.People.(Dialogs)
	userMsg, err := userMsgs.Pick()
	return userMsg.AgentID, err
}

//EnterMsg adds the next messsage into the message table
func EnterMsg(table string, dialogID int, message string) error {
	msg := Dialog{DialogID: dialogID, Msg: message}
	msgs := Dialogs{msg}
	exchange := Exchange{
		Table:  table,
		Put:    []string{"dialog_id", "created", "message"},
		People: msgs,
		Action: "insert",
	}

	for _, m := range msgs {
		c := m.GetSpec(exchange.Put)
		exchange.Spec = append(exchange.Spec, c)
	}
	return exchange.runExchange()

}

//MessageAgent sends a message to the agent and gets the reply
func MessageAgent(agentID, userID int, message string) (string, error) {
	centerr.InfoLog.Printf("and the message is: %s", message)
	return "", nil
}
