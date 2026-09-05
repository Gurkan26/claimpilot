#!/bin/bash
# ollama-entrypoint.sh — Start Ollama server and auto-pull configured model.
# Usage: OLLAMA_MODEL=gemma2:2b ./ollama-entrypoint.sh

set -e

MODEL="${OLLAMA_MODEL:-gemma2:2b}"

echo "==> Starting Ollama server..."
ollama serve &
SERVER_PID=$!

# Wait for server to be ready
echo "==> Waiting for Ollama server to be ready..."
MAX_RETRIES=60
RETRY=0
until curl -sf http://localhost:11434/api/tags > /dev/null 2>&1; do
    RETRY=$((RETRY + 1))
    if [ "$RETRY" -ge "$MAX_RETRIES" ]; then
        echo "ERROR: Ollama server failed to start after ${MAX_RETRIES} seconds"
        exit 1
    fi
    sleep 1
done
echo "==> Ollama server is ready."

# Pull model if not already present
echo "==> Checking if model '${MODEL}' is available..."
if ollama list | grep -q "${MODEL}"; then
    echo "==> Model '${MODEL}' already available."
else
    echo "==> Pulling model '${MODEL}' (this may take a few minutes on first run)..."
    ollama pull "${MODEL}"
    echo "==> Model '${MODEL}' pulled successfully."
fi

echo "==> Ollama ready with model '${MODEL}'. Server PID: ${SERVER_PID}"

# Keep the server running in foreground
wait $SERVER_PID
