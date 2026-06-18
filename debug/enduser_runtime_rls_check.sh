#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:9080}"
ORG_NAME="${ORG_NAME:-luke_e6kz}"
PROJECT_SLUG="${PROJECT_SLUG:-luke}"
DB_NAME="${DB_NAME:-demo_ecommerce}"
MODEL_NAME="${MODEL_NAME:-users}"
PAT="${PAT:-mc_pat_dd2e173ea4def1d752e2a2e17fc5b2e37e802e72a911d700c39e27910f61cab3}"
TEST_USER_ID="${TEST_USER_ID:-rls-test-user-001}"
TEST_USER_NAME="${TEST_USER_NAME:-rls-test-user}"

endpoint="${BASE_URL}/end-user/graphql/org/${ORG_NAME}/project/${PROJECT_SLUG}/db/${DB_NAME}/model/${MODEL_NAME}"

query_find_many='{"query":"query RlsCheck { findMany(take: 20, skip: 0, orderBy: [{id: asc}]) { items { id } totalCount timeCost reqId } count { count } }","operationName":"RlsCheck"}'
query_count_only='{"query":"query RlsCountOnly { count { count } }","operationName":"RlsCountOnly"}'
mutation_create='{"query":"mutation RlsCreate { create(data: {}) { success } }","operationName":"RlsCreate"}'
mutation_update='{"query":"mutation RlsUpdate { update(where: { id: \"u001\" }, data: {}) { success } }","operationName":"RlsUpdate"}'
mutation_delete='{"query":"mutation RlsDelete { delete(where: { id: \"u001\" }) { success } }","operationName":"RlsDelete"}'

check_base_url() {
  local code
  code="$(curl -sS --max-time 3 -o /dev/null -w '%{http_code}' "${BASE_URL}/" || true)"
  if [ -z "${code}" ] || [ "${code}" = "000" ]; then
    printf 'Error: cannot reach %s\n' "${BASE_URL}" >&2
    printf 'Hint: start APISIX (default port 9080) or override BASE_URL.\n' >&2
    exit 1
  fi
}

run_request() {
  local title="$1"
  local role="$2"
  local payload="$3"

  printf '\n=== %s | role=%s ===\n' "$title" "$role"
  curl -sS -X POST \
    "$endpoint" \
    -H "Authorization: Bearer ${PAT}" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -H "X-MC-Auth-Userid-Str: ${TEST_USER_ID}" \
    -H "X-MC-Auth-Username: ${TEST_USER_NAME}" \
    -H "X-MC-Auth-Roles: ${role}" \
    --data-raw "$payload"
  printf '\n'
}

printf 'Endpoint: %s\n' "$endpoint"
printf 'Database: %s\n' "$DB_NAME"
printf 'Model: %s\n' "$MODEL_NAME"

check_base_url

run_request "READ findMany+count" "admin" "$query_find_many"
run_request "READ findMany+count" "viewer" "$query_find_many"
run_request "READ count only" "admin" "$query_count_only"
run_request "CREATE should be denied" "viewer" "$mutation_create"
run_request "UPDATE should be denied" "viewer" "$mutation_update"
run_request "DELETE should be denied" "viewer" "$mutation_delete"
