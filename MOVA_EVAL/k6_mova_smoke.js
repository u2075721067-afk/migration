import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 100,
  duration: '60s',
  thresholds: {
    http_req_duration: ['p(95)<500'], // ms
  },
};

const envelope = {
  "mova_version": "3.1",
  "intent": "smoke",
  "payload": {"name": "k6"},
  "action": [
    {"type": "print", "params": {"value": "Hello, {{payload.name}}!"}},
    {"type": "sleep", "params": {"seconds": 0.01}}
  ]
};

export default function () {
  const url = 'http://localhost:8080/v1/execute?wait=true';
  const res = http.post(url, JSON.stringify(envelope), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(res, {
    'status is 200': (r) => r.status === 200,
    'run succeeded': (r) => r.json('status') === 'success',
  });
  sleep(0.5);
}
