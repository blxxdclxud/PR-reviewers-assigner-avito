import http from 'k6/http';
import { check } from 'k6';

export const options = {
  vus: 10,
  duration: '30s',
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const id = `pr-k6-${__VU}-${__ITER}`;
  const res = http.post(
    `${BASE_URL}/pullRequest/create`,
    JSON.stringify({
      pull_request_id: id,
      pull_request_name: `k6 PR ${id}`,
      author_id: 'load-user-1',
    }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  check(res, { 'status 201': (r) => r.status === 201 });
}
