#!/bin/bash

VERSION=1.0

OUTPUT_PATH="./out"
MASTER_MAIN_PATH="./master/main/master.go"
WORKER_MAIN_PATH="./worker/main/worker.go"
STATIC_PATH="./static"

if [ ! -d $OUTPUT_PATH ]; then
  mkdir $OUTPUT_PATH
fi

echo "building master..."
go build -o $OUTPUT_PATH/master $MASTER_MAIN_PATH

echo "building worker..."
go build -o $OUTPUT_PATH/worker $WORKER_MAIN_PATH

echo "building standalone..."
go build -o $OUTPUT_PATH/standalone standalone.go

echo "other works..."
cp -rf $STATIC_PATH $OUTPUT_PATH/
cp master.json $OUTPUT_PATH/master.json
cp worker.json $OUTPUT_PATH/worker.json

echo "build done lazycron:$VERSION"
