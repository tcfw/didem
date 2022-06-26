#!/bin/bash

SRC_DIR=./api
DST_DIR=./api
FILES=$(find api -type f -iname *.proto)

for f in $FILES 
do
	echo "Generating API for $f"
	protoc -I=$SRC_DIR \
		--go_opt=paths=source_relative \
		--go_out=$DST_DIR \
		--go-grpc_out=$DST_DIR \
		--go-grpc_opt=paths=source_relative \
		$f
done