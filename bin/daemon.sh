#!/bin/sh

EXE=$1
TIMEOUT=$2

while :; do
    $EXE 2>>stderr.log
    sleep $TIMEOUT
done
