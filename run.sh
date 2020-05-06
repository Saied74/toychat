#!/bin/bash


cd $GOPATH/bin
./nats-server &

cd $GOPATH/src/toychat

cd web
go build -o ui .
./ui &
cd ../mat
go build -o mat .
./mat &
cd ../chat
go build -o chat .
./chat &
