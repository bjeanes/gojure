#!/usr/bin/env bash
find . -not \( -name ".git" -prune \) -type d | xargs go test
