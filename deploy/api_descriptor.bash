#!/usr/bin/bash

protoc \
  --include_imports \
  --include_source_info \
  --proto_path=../api \
  --proto_path=../internal/api \
  --proto_path=../third_party \
  --descriptor_set_out=api_descriptor.pb \
  ../api/service.proto
