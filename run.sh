#!/bin/bash

_term() {
 echo "SIGTERM received. Sending USR1 signal to application."
 kill -USR1 $(pidof ${executableName})
 sleep 10
}
trap _term SIGTERM

make all &
wait $!
