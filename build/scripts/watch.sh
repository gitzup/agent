#!/usr/bin/env bash

INOTIFY=$(which inotifywait 2>/dev/null)
[[ -z "${INOTIFY}" ]] && echo "Install inotify-tools first." >&2 && exit 1

function stop {
    pkill -P $$
}
trap stop EXIT

PID=""
FIRST_RUN="true"
while true; do

    # For the first run, simulate an event of type "first run"
    if [[ "${FIRST_RUN}" == "true" ]]; then
        EVENT="First run."
        RC=0
    else
        EVENT=$(inotifywait -e create,modify,delete -t 1 -r -q ./Makefile ./api ./cmd ./internal)
        RC=$?
    fi

    # If we got an event (exit code 0) make & run. Run the PID in the background, but save the PID.
    if [[ ${RC} == 2 ]]; then
        continue

    elif [[ ${RC} == 0 ]]; then
        echo >&2
        echo "=========================================" >&2
        echo "CHANGE DETECTED (${RC}): ${EVENT}" >&2
        echo "=========================================" >&2
        echo >&2

        if [[ -n "${PID}" && "${PID}" != "0" ]]; then
            kill ${PID}
            sleep 1
            if [[ "$(ps --pid ${PID} | wc -l)" != "0" ]]; then
                kill -9 ${PID}
            fi
        fi

        make
        RC=$?
        if [[ ${RC} != 0 ]]; then
            echo "Build terminated with exit code ${RC}!" >&2
            continue
        fi

        GOOGLE_APPLICATION_CREDENTIALS=$(ls gcp-buildagent-key-*.local.json|head -1) ./agent &
        PID=$?
        FIRST_RUN="false"

    else
        echo "File watcher failed! (${RC})" >&2
        exit 1
    fi

done
