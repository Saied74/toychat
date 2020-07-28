//Copyright (c) 2020 Saied Seghatoleslami
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.
/*
broker package contains the objects and methods for communication over nats
broker.  It supports communication between the front end to the dbmgr, the
backend to the dbmger, and between the front and the backend.

The main vehicle for communication is the Excahnge type.  The indvidual Fields
of the Exchange type are described in the broker.go file next to each field.
The element of exchange that describes the database tables is People type
for admin, agent, and user roles (admins and agents are in the same table
but some agent fields are not relevant for the admin role.)  People is a slice
of the Person type to allow for sending or getting multiple rows from the db.
Dialogs type is for representing the messages and dialogs table and also for
communication between the frontend and backend.  The same is true of Dialogs
which is a slice of dialog for the same rason.  To accomodate these two data
types, the People field of the Exchange type is an interface (TableProxy)

type TableProxy interface {
	Length() int
	PickZero() (*Person, error)
	Pick() (*Dialog, error)
}

The methods defined by the interface are to allow extracting the length and the
first elelemtn (zeroth element) of each slice.  Pick and PickZero are mirror
images of each other for the two types PickZero for People and Pick for Dialogs.
For the other type, they are dummy functions.

The Put field of the Exchange type is a slice of the column names for the
insert and put functions.  Insert inserts a new record into the dabase and
put updates one or more fields of the database.  In the case of put, the
"WHERE" clause of the database is satisfied by the Spec field of the Exchange
type.  For example, for the statement:

"UPDATE admins SET online = ? WHERE id = ? AND role = ?"

the online would be the element of the Put slice and id and role would be the
two elements of the Spec slice.  This can be seen in the PutLine function in
the broker.go file of the broker package.  The data that is inserted into the
prepared statement for the ? positions is in the appropriate field of the
People or the Dialogs type.  The SpecList is a double slice of the pointers
to the fields of the People or Dialog types.  But since since the sql library
accepts interface{} types, they are built in this type to start with and
supplied as varidic functions arguments in the dbmgr application.  For the insert
statement, this is done by BuildInsert which is a method on Person.  The method
Specify creates the slice of SpecList for the put function since it has
both a set of columns to be updated and a set of filters to be satisified.
Currnetly only AND logic for the filers is supported.

// TODO: The parallel functions for Dialogs are not implemented.

The Get field of the Exchange type plays the same role for the get statement.
It specifies the fields to be extracted.  The SpecList field specifies the conditions
to be used for flitering (again only AND logic is supported at this time).
The method GetSpec on Person builds the Spec slice for meeting the WHERE clause
of the get function.  Because the get function can extract an unknown number of
rows, the ScanSpec slice is built while running rows.Scan method in the get
function of the dbmgr.  It is built bun the broker.GetItems.

When the data is returned, it is extraced back to the original type by the
broker.GetBack function.

In general, the individual methods called by the handlers in the fontend
and backend follow the following pattern:
1. Build the People or Dialogs MatchPattern
2. Build the exchange pattern and add the People or Dialogs to it.
3. In the case of the get Action, run runGetExchange and extract the returned fields.
4. In the case of the put Action, build the Spec slice and run runExchange
5. In the case of insert Action, do exactly like put.

runExchange is helper function that handles boilerplate code to Register
gob encorders, gob encode, exchange message and reply over nats with dbmger
gob decoded and decode error (more on this later).  runGetExchange first builds
the Spec field and then runs runExchange.

// TODO: This stuff can be further simplified.  Dobule slice may not be necessary.

EncodeErr and DecodeErr encode error at the source and decode it at the
destination since errors don't travel well over gob.







*/

package broker
