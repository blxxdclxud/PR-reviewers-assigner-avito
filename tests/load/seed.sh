#!/usr/bin/env bash
# Seed script for load testing.
# Creates 1 team with 15 active members and 50 open PRs (each gets 2 reviewers).
# Run once before load tests: bash tests/load/seed.sh [BASE_URL]

set -e

BASE_URL="${1:-http://localhost:8080}"

echo "==> Seeding data at $BASE_URL"

# Create team with 15 users
MEMBERS='[]'
for i in $(seq 1 15); do
  MEMBERS=$(echo "$MEMBERS" | \
    python3 -c "import sys,json; m=json.load(sys.stdin); m.append({'user_id':'load-user-'+'$i','username':'LoadUser$i','is_active':True}); print(json.dumps(m))")
done

curl -sf -X POST "$BASE_URL/team/add" \
  -H "Content-Type: application/json" \
  -d "{\"team_name\":\"load-test-team\",\"members\":$MEMBERS}" > /dev/null

echo "   team created with 15 members"

# Create 50 PRs (author = load-user-1, each auto-assigns 2 reviewers)
for i in $(seq 1 50); do
  curl -sf -X POST "$BASE_URL/pullRequest/create" \
    -H "Content-Type: application/json" \
    -d "{\"pull_request_id\":\"load-pr-$i\",\"pull_request_name\":\"Load PR $i\",\"author_id\":\"load-user-1\"}" > /dev/null
done

echo "   50 PRs created"
echo "==> Seed complete"
