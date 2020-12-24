#!/bin/env bash

set -ex

terraform destroy -target=google_compute_region_instance_group_manager.fabula
