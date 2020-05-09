package main

import (
	"database/sql"
	"time"

	"github.com/saied74/toychat/pkg/broker"
)

type userModel struct {
	dB *sql.DB
}

type user struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
	Active         bool
}

func (st *sT) insertUserR(name, email, password string) error {
	// var err error
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
