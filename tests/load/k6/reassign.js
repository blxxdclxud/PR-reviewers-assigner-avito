import http from 'k6/http';
import { check } from 'k6';

export const options = {
  vus: 10,
  duration: '30s',
};

// Each VU owns its own PR to avoid deadlocks from concurrent reassigns on same row.
// Requires seed.sh to have run first (creates load-pr-1 .. load-pr-50).
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const POOL_SIZE = 50;

export default function () {
  const prID = `load-pr-${(__VU % POOL_SIZE) + 1}`;

  // Get current reviewers to find an old one
  const getRes = http.get(
    `${BASE_URL}/team/get?team_name=load-test-team`,
  );
  if (getRes.status !== 200) return;

  const members = JSON.parse(getRes.body).members || [];
  const oldReviewer = members.find(
    (m) => m.user_id !== 'load-user-1',
  );
  if (!oldReviewer) return;

  const res = http.post(
    `${BASE_URL}/pullRequest/reassign`,
    JSON.stringify({
      pull_request_id: prID,
      old_user_id: oldReviewer.user_id,
    }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  // 200 = success, 409 = business rule conflict (acceptable)
  check(res, { 'no server error': (r) => r.status !== 500 });
}
