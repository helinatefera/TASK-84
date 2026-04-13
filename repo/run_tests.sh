#!/bin/bash
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

echo "============================================="
echo "  Local Insights Portal - Test Suite"
echo "============================================="
echo ""

UNIT_PASS=false
FRONTEND_PASS=false
API_PASS=false

# --- Cleanup ---
cleanup() {
    echo ""
    echo "--- Cleaning up ---"
    docker compose down --remove-orphans --volumes 2>/dev/null || true
}

trap cleanup EXIT

# --- Teardown existing ---
echo "--- Tearing down existing containers ---"
docker compose down --remove-orphans --volumes 2>/dev/null || true
docker rm -f local_insights_mysql local_insights_backend local_insights_frontend 2>/dev/null || true

# ========== UNIT TESTS ==========
# Unit tests run in a standalone Go container with no DB dependency.
echo ""
echo "============================================="
echo "  Running Unit Tests"
echo "============================================="

docker run --rm \
    -v "$SCRIPT_DIR/backend:/backend:ro" \
    -v "$SCRIPT_DIR/unit_tests:/unit_tests:ro" \
    -w /unit_tests \
    golang:1.25-alpine \
    sh -c "go test -v -count=1 ./... 2>&1" && UNIT_PASS=true || UNIT_PASS=false

if [ "$UNIT_PASS" = true ]; then
    echo ""
    echo "Unit Tests: PASS"
else
    echo ""
    echo "Unit Tests: FAIL"
fi

# ========== FRONTEND TESTS ==========
echo ""
echo "============================================="
echo "  Running Frontend Tests"
echo "============================================="

docker run --rm \
    -v "$SCRIPT_DIR/frontend:/frontend:ro" \
    node:20-alpine \
    sh -c "
      cp -r /frontend /tmp/fe && cd /tmp/fe &&
      rm -rf node_modules &&
      npm ci --legacy-peer-deps 2>/dev/null &&
      npx vitest run 2>&1
    " && FRONTEND_PASS=true || FRONTEND_PASS=false

if [ "$FRONTEND_PASS" = true ]; then
    echo ""
    echo "Frontend Tests: PASS"
else
    echo ""
    echo "Frontend Tests: FAIL"
fi

# ========== BUILD & START SERVICES ==========
echo ""
echo "--- Building services ---"
docker compose build 2>&1 | tail -10

echo ""
echo "--- Starting services ---"
docker compose up -d

# --- Wait for MySQL ---
echo ""
echo "--- Waiting for MySQL to be ready ---"
for i in $(seq 1 60); do
    if docker compose exec -T mysql mysqladmin ping -h localhost -u root -prootpassword 2>/dev/null | grep -q "alive"; then
        echo "MySQL is ready (attempt $i)"
        break
    fi
    if [ "$i" = "60" ]; then
        echo "FAIL: MySQL did not become ready in time"
        docker compose logs mysql | tail -20
        exit 1
    fi
    sleep 2
done

# --- Wait for Backend ---
echo ""
echo "--- Waiting for Backend to be ready ---"
for i in $(seq 1 60); do
    STATUS=$(docker compose exec -T backend wget --no-check-certificate -q -O- https://localhost:8443/api/v1/health 2>/dev/null || true)
    if echo "$STATUS" | grep -q "healthy"; then
        echo "Backend is ready (attempt $i)"
        break
    fi
    if [ "$i" = "60" ]; then
        echo "FAIL: Backend did not become ready"
        docker compose logs backend | tail -30
        exit 1
    fi
    sleep 3
done

# ========== API TESTS ==========
echo ""
echo "============================================="
echo "  Running API Tests"
echo "============================================="

# Get the Docker network name used by compose
NETWORK=$(docker inspect "$(docker compose ps -q backend)" --format '{{range $k, $v := .NetworkSettings.Networks}}{{$k}}{{end}}' 2>/dev/null | head -1)

if [ -z "$NETWORK" ]; then
    echo "FAIL: Could not determine Docker network"
    API_PASS=false
else
    docker run --rm \
        --network="$NETWORK" \
        -e API_BASE_URL=https://backend:8443/api/v1 \
        -v "$SCRIPT_DIR/API_tests:/tests:ro" \
        alpine:3.19 \
        sh -c "apk add --no-cache curl jq bash mysql-client >/dev/null 2>&1 && bash /tests/run_api_tests.sh" && API_PASS=true || API_PASS=false
fi

if [ "$API_PASS" = true ]; then
    echo ""
    echo "API Tests: PASS"
else
    echo ""
    echo "API Tests: FAIL"
fi

# ========== SUMMARY ==========
echo ""
echo "============================================="
echo "  Test Summary"
echo "============================================="
echo "  Unit Tests:     $([ "$UNIT_PASS" = true ] && echo 'PASS' || echo 'FAIL')"
echo "  Frontend Tests: $([ "$FRONTEND_PASS" = true ] && echo 'PASS' || echo 'FAIL')"
echo "  API Tests:      $([ "$API_PASS" = true ] && echo 'PASS' || echo 'FAIL')"
echo "============================================="

if [ "$UNIT_PASS" = true ] && [ "$FRONTEND_PASS" = true ] && [ "$API_PASS" = true ]; then
    echo ""
    echo "ALL TESTS PASSED"
    exit 0
else
    echo ""
    echo "SOME TESTS FAILED"
    exit 1
fi
