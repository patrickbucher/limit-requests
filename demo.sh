#!/usr/bin/bash

for i in {1..10}
do
    sleep 2
    curl http://localhost:8080 &
done
