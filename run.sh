#!/bin/bash


if [ $# -lt 1 ]
then
  echo 'must provide at least one argument'
  exit 0
fi

if [ $1 == 'ui' ] && [ $# -lt 2 ]
then
  echo 'too few arguments for invoking ui'
  exit 0
fi

if [ $1 == 'dbmgr' ] && [ $# -lt 2 ]
then
  echo 'too few arguments for invoking dbmgr'
  exit 0
fi

if [ $1 == 'all' ] && [ $# -lt 2 ]
then
  echo 'too few arguments for invoking all'
  exit 0
fi

cd $GOPATH/src/toychat

if [ $1 == 'ui' ]
then
  cd web
  killall ui
  rm ui
  go build -o ui .
  ./ui -pw $2 &
  cd ..
  exit 0
fi

if [ $1 == 'mat' ]
then
  cd mat
  killall matMat
  rm matMat
  go build -o matMat .
  ./matMat &
  cd ..
  exit 0
fi

if [ $1 == 'chat' ]
then
  cd chat
  killall chat
  rm chat
  go build -o chat .
  ./chat &
  cd ..
  exit 0
fi

if [ $1 == 'dbmgr' ]
then
  cd dbmgr
  killall dbmgr
  rm dbmgr
  go build -o dbmgr .
  ./dbmgr -pw $2 &
  cd ..
  exit 0
fi

if [ $1 == 'nats' ]
then
  killall nats-server
  cd $GOPATH/bin
  ./nats-server &
  cd $GOPATH/src/toychat
  exit 0
fi

if [ $1 = all ]
  then

killall nats-server
killall ui
killall matMat
killall chat
killall dbmgr

cd $GOPATH/bin
./nats-server &

cd $GOPATH/src/toychat

cd web
rm ui
go build -o ui .
./ui -pw $2 &
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
./dbmgr -pw $2 &
exit 0
fi

if [ $1 == 'kill' ]
then
  killall nats-server
  killall ui
  killall matMat
  killall chat
  killall dbmgr
fi
