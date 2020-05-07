#!/bin/bash


cd $GOPATH/bin
./nats-server &

cd $GOPATH/src/toychat

cd web
rm ui
go build -o ui .
./ui &
cd ../mat
rm matMat
go build -o matMat .
./matMat &
cd ../chat
rm chat
go build -o chat .
./chat &
cd ../dbmgr
rm dbmgr
go build -o dbmgr .
./dbmgr &
