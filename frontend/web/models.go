package main

import (
	"database/sql"

	"github.com/saied74/toychat/pkg/broker"
)

type userModel struct {
	dB *sql.DB
}

//sends user name, eamil and password to the far end to be inserted into the db.
// TODO: encode password here instead of sending it in plaintext over the wire.
//note that since it is struct that is sent, it is gob encoded.
//Also, since error type does not gob encode, it is string encoded at the
//far end and deoded back into the error type at this end.
//The requested action is text encoded into the struct field as well.
func (st *sT) insertUserR(name, email, password string) error {
	exchData := broker.ExchData{
		Name:     name,
		Email:    email,
		Password: password,
		Action:   "insert",
	}
	sendData, err := exchData.ToGob()
	if err != nil {
		return err
	}
	answer := st.chatConnection(string(sendData), "forDB", "")
	exchData.FromGob(answer)
	return exchData.DecodeErr()
}

//see comments for insertUserR above.
func (st *sT) authenticateUserR(email, password string) (int, error) {
	exchData := broker.ExchData{
		Email:    email,
		Password: password,
		Action:   "authenticate",
	}
	sendData, err := exchData.ToGob()
	if err != nil {
		return 0, err
	}
	answer := st.chatConnection(string(sendData), "forDB", "")
	exchData.FromGob(answer)
	return exchData.ID, exchData.DecodeErr()
}

//see comments for insertUserR above.
func (st *sT) getUserR(id int) (*broker.ExchData, error) {
	exchData := broker.ExchData{
		ID:     id,
		Action: "getuser",
	}
	sendData, err := exchData.ToGob()
	if err != nil {
		return &broker.ExchData{}, err
	}
	answer := st.chatConnection(string(sendData), "forDB", "")
	exchData.FromGob(answer)

	return &exchData, exchData.DecodeErr()
}
