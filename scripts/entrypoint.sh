#!/bin/sh
set -e

PROJECT_DIR=$1
OUTPUT_FORMAT=$2
TARGET_STACK=$3

echo "🚀 Starting AITriage Security Scan..."

# Construct the base command
CMD="aitriage scan $PROJECT_DIR --format $OUTPUT_FORMAT"

# Append stack flag if provided
if [ -n "$TARGET_STACK" ]; then
  CMD="$CMD --stack $TARGET_STACK"
fi

echo "Running: $CMD"

# Since GitHub actions runs the docker container with the workspace mounted at /github/workspace,
# the project dir will be relative to that.
cd "$GITHUB_WORKSPACE" || cd /github/workspace || true

# Execute the command
# We use eval to appropriately expand the variables
eval $CMD

EXIT_CODE=$?

# If we output sarif, let's notify the user
if [ "$OUTPUT_FORMAT" = "sarif" ]; then
  echo "✅ SARIF output generated."
fi

exit $EXIT_CODE
