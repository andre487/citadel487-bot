#!/usr/bin/env bash
set -eufo pipefail

cd "$(dirname "$0")"

go run . --token-file ~/.tokens/andre487-bot
