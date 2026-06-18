#!/bin/sh
# Entrypoint for the AITriage GitHub Action (Docker container action).
#
# action.yml passes configuration via AITRIAGE_* environment variables
# (clean names chosen to avoid GitHub's INPUT_<NAME> hyphen quirk). This
# script assembles a safe argv and execs the aitriage binary.
set -e

CMD="${AITRIAGE_COMMAND:-scan}"
PROJECT_DIR="${AITRIAGE_PROJECT_DIR:-.}"

# Base: "<command> <path>"
set -- "$CMD" "$PROJECT_DIR"

# scan-specific flags (only meaningful for the deterministic gate)
if [ "$CMD" = "scan" ]; then
  [ -n "$AITRIAGE_FORMAT" ]      && set -- "$@" --format "$AITRIAGE_FORMAT"
  [ -n "$AITRIAGE_OUTPUT_FILE" ] && set -- "$@" --out "$AITRIAGE_OUTPUT_FILE"
  [ -n "$AITRIAGE_FAIL_ON" ]     && set -- "$@" --fail-on "$AITRIAGE_FAIL_ON"
  [ -n "$AITRIAGE_FAIL_SCORE" ]  && set -- "$@" --fail-score "$AITRIAGE_FAIL_SCORE"
  [ -n "$AITRIAGE_HEALTH_PROFILE" ] && set -- "$@" --health-profile "$AITRIAGE_HEALTH_PROFILE"
  [ -n "$AITRIAGE_STACK" ]       && set -- "$@" --stack "$AITRIAGE_STACK"
  [ -n "$AITRIAGE_DIFF" ]        && set -- "$@" --diff "$AITRIAGE_DIFF"
  [ "$AITRIAGE_BASELINE" = "true" ] && set -- "$@" --baseline
fi

# Freeform extra args — used for agent/fix/sbom or advanced scan flags.
# Intentionally word-split so callers can pass multiple flags.
if [ -n "$AITRIAGE_ARGS" ]; then
  # shellcheck disable=SC2086
  set -- "$@" $AITRIAGE_ARGS
fi

# Agent-specific flags (not applicable to scan/fix/sbom).
if [ "$CMD" = "agent" ]; then
  [ -n "$AITRIAGE_SUMMARY_FILE" ] && set -- "$@" --summary-out "$AITRIAGE_SUMMARY_FILE"
fi

echo "+ aitriage $*" >&2
exec aitriage "$@"
