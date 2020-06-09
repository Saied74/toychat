//this file contains the end user broker methods

package broker

import (
	"fmt"
	"time"
)

const (
	iD             = "id"
	Name           = "name"
	Email          = "email"
	HashedPassword = "hashed_password"
	Created        = "created"
	Role           = "role"
	Active         = "active"
	Online         = "online"
)

//BuildInsert uses the "put" slice pattern to build an empty interface
//slice to be used with the INSERT statement
func (p *Person) BuildInsert(put []string) []interface{} {
	var c = []interface{}{}
	for _, sp := range put {
		switch sp {
		case iD:
			c = append(c, p.ID)
		case Name:
			c = append(c, p.Name)
		case Email:
			c = append(c, p.Email)
		case HashedPassword:
			c = append(c, p.HashedPassword)
		case Created:
			continue
		case Active:
			c = append(c, p.Active)
		case Online:
			c = append(c, p.Online)
		case Role:
			c = append(c, p.Role)
		}
	}
	return c
}

//GetSpec builds the specifications to provide to the WHERE clause of SQL
func (p *Person) GetSpec(spec []string) []interface{} {
	var g = []interface{}{}
	for _, sp := range spec {
		switch sp {
		case iD:
			g = append(g, p.ID)
		case Name:
			g = append(g, p.Name)
		case Email:
			g = append(g, p.Email)
		case HashedPassword:
			g = append(g, p.HashedPassword)
		case Created:
			g = append(g, p.Created)
		case Active:
			g = append(g, p.Active)
		case Online:
			g = append(g, p.Online)
		case Role:
			g = append(g, p.Role)
		}
	}
	return g
}

//GetItems uses the get string to generate an interface to be passed to the
//sql.Execute statement for the INSERT sql command.
func (p *Person) GetItems(get []string) []interface{} {
	var g = []interface{}{}
	for _, sp := range get {
		switch sp {
		case iD:
			g = append(g, &p.ID)
		case Name:
			g = append(g, &p.Name)
		case Email:
			g = append(g, &p.Email)
		case HashedPassword:
			g = append(g, &p.HashedPassword)
		case Created:
			g = append(g, &p.Created)
		case Role:
			g = append(g, &p.Role)
		case Active:
			g = append(g, &p.Active)
		case Online:
			g = append(g, &p.Online)
		}
	}
	return g
}

//GetBack reverses the get item and takes the interface items and gets the
//underlying data back.
func (p *Person) GetBack(get []string, g []interface{}) error {
	for i, sp := range get {
		switch sp {
		case iD:
			xID, ok := g[i].(*int)
			if !ok {
				return fmt.Errorf("ID (int) type assertion failed")
			}
			p.ID = *xID
		case Name:
			xName, ok := g[i].(*string)
			if !ok {
				return fmt.Errorf("Name (string) type assertion failed")
			}
			p.Name = *xName
		case Email:
			xEmail, ok := g[i].(*string)
			if !ok {
				return fmt.Errorf("Email (string) type assertion failed")
			}
			p.Email = *xEmail
		case HashedPassword:
			xPass, ok := g[i].(*string)
			if !ok {
				return fmt.Errorf("Hashed Password (string) type assertion failed")
			}
			p.HashedPassword = *xPass
		case Created:
			xCreated, ok := g[i].(*time.Time)
			if !ok {
				return fmt.Errorf("Created (time.Time) type assertion failed")
			}
			p.Created = *xCreated
		case Role:
			xRole, ok := g[i].(*string)
			if !ok {
				return fmt.Errorf("Role (string) type assertion failed")
			}
			p.Role = *xRole
		case Active:
			xActive, ok := g[i].(*bool)
			if !ok {
				return fmt.Errorf("Active (bool) type assertion failed")
			}
			p.Active = *xActive
		case Online:
			xOnline, ok := g[i].(*bool)
			if !ok {
				return fmt.Errorf("Online (bool) type assertion failed")
			}
			p.Online = *xOnline
		}
	}
	return nil
}

//InsertEUR is for inserting end users (EU) from the front end
func InsertEUR(table, name, email, password string) error {
	exchange := Exchange{
		Table: table,
		Put:   []string{Name, Email, HashedPassword, Created},
		Spec:  []string{},
		People: []Person{
			Person{
				Name:           name,
				Email:          email,
				HashedPassword: password,
			},
		},
		Action: "insert",
	}
	sendData, err := exchange.ToGob()
	if err != nil {
		return err
	}
	answer := chatConnection(string(sendData), "forDB")
	exchange.FromGob(answer)
	return exchange.DecodeErr()
}

//AuthenticateEUR gob encodes exchData and sends it to the dbmgr over nats
//EU stands for end user
func AuthenticateEUR(table, email string) (*Person, error) {
	exchange := Exchange{
		Table: table,
		Put:   []string{},
		Spec:  []string{Email},
		Get:   []string{iD, Name, Email, HashedPassword, Created, Active},
		People: []Person{
			Person{Email: email},
		},
		Action: "get",
	}
	sendData, err := exchange.ToGob()
	if err != nil {
		return &Person{}, err
	}
	answer := chatConnection(string(sendData), "forDB")
	exchange.FromGob(answer)
	err = exchange.DecodeErr()
	if err != nil {
		return &Person{}, err
	}
	length := len(exchange.People)
	if length == 1 {
		return &exchange.People[0], nil
	}
	return &Person{}, fmt.Errorf("AuthenticateUserR brought back %d records",
		length)
}

//GetEUR gets the user infromation for the database (EU for end user)
func GetEUR(table string, id int) (*Person, error) {
	exchange := Exchange{
		Table: table,
		Put:   []string{},
		Spec:  []string{iD},
		Get: []string{iD, Name, Email, HashedPassword, Created,
			Active, Online},
		People: []Person{
			Person{ID: id},
		},
		Action: "get",
	}
	sendData, err := exchange.ToGob()
	if err != nil {
		return &Person{}, err
	}
	answer := chatConnection(string(sendData), "forDB")
	exchange.FromGob(answer)
	length := len(exchange.People)
	if length == 1 {
		return &exchange.People[0], nil
	}
	return &Person{}, fmt.Errorf("GetUserR brought back %d",
		length)
}
