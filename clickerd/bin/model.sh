#!/usr/bin/env bash

readonly usage="usage: $0 <device-id> <model-json>"
readonly device="${1?${usage}}"
readonly model="${2?${usage}}"

particle call $device cancel
particle call $device addNodes "$model"
particle call $device align
