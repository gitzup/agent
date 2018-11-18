#!/usr/bin/env bash

set -exu -o pipefail

if [[ -n "${TRAVIS_TAG}" ]]; then
    # TODO: if this tag is the GitHub Latest Release, use "make latest"
    PUSH=true TAG=${TRAVIS_TAG} make docker

elif [[ -n "${TRAVIS_PULL_REQUEST_SHA}" ]]; then
    # TODO: consider pushing to a staging Docker registry
    PUSH=false TAG=${TRAVIS_PULL_REQUEST_SHA} make docker

else
    # TODO: consider pushing to a staging Docker registry
    PUSH=false TAG=${TRAVIS_COMMIT} make docker
fi
