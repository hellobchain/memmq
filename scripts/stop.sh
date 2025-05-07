#!/bin/bash
MEMMQ_BIN="./bin/memmq.bin"
pid=`ps -ef | grep ${MEMMQ_BIN} | grep -v grep | awk '{print $2}'`
if [ ! -z ${pid} ];then
    echo "kill $pid"
    kill $pid
else
    echo "$MEMMQ_BIN already stopped"
fi