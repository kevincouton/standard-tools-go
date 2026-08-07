#!/usr/bin/env bash
# Runs the GitHub Actions CI workflow locally using act + podman.
set -euo pipefail

cd "$(dirname "$0")/.."

# Detect the podman socket to use as the container daemon.
# On macOS the host-side API socket is exposed by the podman machine, which is
# what act needs to mount into its runner containers.
DOCKER_HOST=""
if command -v podman >/dev/null 2>&1; then
  # Prefer the host-side machine socket when available (macOS/Windows).
  machine_socket=$(podman machine inspect --format '{{.ConnectionInfo.PodmanSocket.Path}}' 2>/dev/null || true)
  if [ -n "${machine_socket:-}" ]; then
    DOCKER_HOST="unix://$machine_socket"
  else
    socket_path=$(podman info --format '{{.Host.RemoteSocket.Path}}' 2>/dev/null || true)
    if [ -n "${socket_path:-}" ]; then
      DOCKER_HOST="$socket_path"
    fi
  fi
fi

# Default to the conventional podman socket path.
export DOCKER_HOST="${DOCKER_HOST:-unix:///run/podman/podman.sock}"

echo "Using container daemon: $DOCKER_HOST"

# Ensure act is available, installing via mise if possible.
if ! command -v act >/dev/null 2>&1; then
  if command -v mise >/dev/null 2>&1; then
    mise install act
  else
    echo "WARN: act is not installed and mise is not available. Skipping local act run." >&2
    exit 0
  fi
fi

# Pin the default runner image so act does not prompt interactively.
ACT_PLATFORM="-P ubuntu-latest=catthehacker/ubuntu:act-latest"
# Start act's embedded artifact server so upload-artifact works locally.
ARTIFACT_PATH="/tmp/standard-tools-go-act-artifacts"
mkdir -p "$ARTIFACT_PATH"
ACT_ARTIFACT="--artifact-server-path $ARTIFACT_PATH"

echo "Running quality job locally with act..."
act push --defaultbranch main --job quality --container-daemon-socket "$DOCKER_HOST" $ACT_PLATFORM $ACT_ARTIFACT

echo "Running integration job locally with act..."
act push --defaultbranch main --job integration --container-daemon-socket "$DOCKER_HOST" $ACT_PLATFORM $ACT_ARTIFACT

# Produce a local visual test report and keep a local copy.
echo "Generating local visual test report..."
command -v go >/dev/null 2>&1 || { echo "Go is required to run local tests but is not installed" >&2; exit 1; }
[ -x ./scripts/visual-test-report.sh ] || { echo "Missing ./scripts/visual-test-report.sh" >&2; exit 1; }
set +e
bash -o pipefail -c 'go test ./... 2>&1 | tee test-output.log; test_exit=${PIPESTATUS[0]}; ./scripts/visual-test-report.sh test-output.log test-report.html; exit $test_exit'
test_exit=$?
set -e

if [ -f test-report.html ]; then
  cp test-report.html test-report-local.html
  echo "Local report copied to test-report-local.html"
fi

exit $test_exit
