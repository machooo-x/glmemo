#!/bin/sh

# kill
PIDS=$(ps aux | grep "${PWD}/linux-amd64-glmemo" | grep -v 'grep' | awk '{print $2}')
if [ -n "$PIDS" ]; then
    kill -9 $PIDS
fi

# restart
./daemon.sh $PWD"/linux-amd64-glmemo" 10 &
