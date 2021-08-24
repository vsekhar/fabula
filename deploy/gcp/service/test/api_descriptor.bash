#!/usr/bin/bash

protoc \
  --include_imports \
  --include_source_info \
  --proto_path=. \
  --proto_path=../../../../third_party \
  --descriptor_set_out=api_descriptor.pb \
  hello.proto
