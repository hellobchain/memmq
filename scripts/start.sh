#!/bin/bash
MEMMQ_BIN="./bin/memmq.bin"
nohup ./${MEMMQ_BIN} >memmq.log 2>&1 &