#!/usr/bin/env bash

set -exu -o pipefail

if [[ "${TRAVIS_PULL_REQUEST}" == "false" && -z "${TRAVIS_TAG}" && "${TRAVIS_BRANCH}" == "master" ]]; then
    MAKE_COMMAND=latest
else
    MAKE_COMMAND=docker
fi

TAG=${TRAVIS_COMMIT} make ${MAKE_COMMAND}
