#!/bin/bash
CMD=$1
if [ -z "$CMD" ]; then
	echo "usage: ./do.sh append
       ./do.sh read
       ./do.sh read-at"
       exit 0
fi

case "$CMD" in
	"append")
         time grpcurl -plaintext -d @ localhost:2006 proto.EventStore/Append <stream-data-pretty.json
	 ;;
	"read")
         time grpcurl -plaintext -d @ localhost:2006 proto.EventStore/Read <streams-pretty.json
         ;;
	"read-at")
         time grpcurl -plaintext -d @ localhost:2006 proto.EventStore/ReadAt <streams-at-pretty.json
	 ;;
	*)
	echo "unknown command"
	exit 1
	;;
esac
