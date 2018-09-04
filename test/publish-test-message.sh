#!/usr/bin/env bash

set -e

MESSAGE=$(cat $(dirname $0)/pubsub-build-request.json | jq --compact-output '.')
gcloud pubsub topics publish projects/gitzup/topics/executions --message="${MESSAGE}"
