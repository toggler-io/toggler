#!/usr/bin/env bash
set -e -u

if [[ -e bin/toggler ]]; then
	exec bin/toggler ${@}
fi

go run cmd/toggler/main.go ${@}
