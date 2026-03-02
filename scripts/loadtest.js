// OpenMind Load Test Script (k6)
// Run: k6 run scripts/loadtest.js
//
// Prerequisites: Install k6 from https://k6.io
//
// Usage:
//   k6 run scripts/loadtest.js                          # default (10 VUs, 30s)
//   k6 run --vus 50 --duration 60s scripts/loadtest.js  # custom

import http from "k6/http";
import { check, sleep } from "k6";
import { Rate } from "k6/metrics";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080/api/v1";

// Custom metrics
const errorRate = new Rate("errors");

// Test configuration
export const options = {
  stages: [
    { duration: "10s", target: 10 },  // Ramp up to 10 users
    { duration: "30s", target: 10 },  // Stay at 10 users
    { duration: "10s", target: 25 },  // Ramp up to 25 users
    { duration: "30s", target: 25 },  // Stay at 25 users
    { duration: "10s", target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ["p(95)<500"], // 95% of requests under 500ms
    errors: ["rate<0.1"],             // Error rate under 10%
  },
};

// ── Health Check ──────────────────────────────────────────────────
function healthCheck() {
  const res = http.get(`${BASE_URL}/health`);
  const ok = check(res, {
    "health: status 200": (r) => r.status === 200,
    "health: status ok": (r) => {
      const body = JSON.parse(r.body);
      return body.status === "ok";
    },
    "health: db connected": (r) => {
      const body = JSON.parse(r.body);
      return body.database === "connected";
    },
  });
  errorRate.add(!ok);
}

// ── Auth Flow ─────────────────────────────────────────────────────
function authFlow() {
  const loginPayload = JSON.stringify({
    email: __ENV.TEST_EMAIL || "admin@test.com",
    password: __ENV.TEST_PASSWORD || "Password123!",
  });

  const res = http.post(`${BASE_URL}/auth/login`, loginPayload, {
    headers: { "Content-Type": "application/json" },
  });

  const ok = check(res, {
    "login: status 200 or 401": (r) => r.status === 200 || r.status === 401,
    "login: response time < 1s": (r) => r.timings.duration < 1000,
  });
  errorRate.add(!ok);

  // Return token if login succeeded
  if (res.status === 200) {
    try {
      const body = JSON.parse(res.body);
      return body.data?.access_token || null;
    } catch {
      return null;
    }
  }
  return null;
}

// ── Authenticated Patient List ────────────────────────────────────
function patientList(token) {
  if (!token) return;

  const res = http.get(`${BASE_URL}/patients?page=1&page_size=10`, {
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
  });

  const ok = check(res, {
    "patients: status 200 or 401": (r) => r.status === 200 || r.status === 401,
    "patients: response time < 500ms": (r) => r.timings.duration < 500,
  });
  errorRate.add(!ok);
}

// ── Public Plans ──────────────────────────────────────────────────
function publicPlans() {
  const res = http.get(`${BASE_URL}/plans`);
  const ok = check(res, {
    "plans: status 200": (r) => r.status === 200,
    "plans: response time < 300ms": (r) => r.timings.duration < 300,
  });
  errorRate.add(!ok);
}

// ── Main Test Function ────────────────────────────────────────────
export default function () {
  // Always test health and public endpoints
  healthCheck();
  sleep(0.5);

  publicPlans();
  sleep(0.5);

  // Try auth flow
  const token = authFlow();
  sleep(0.5);

  // Test authenticated endpoints if we got a token
  patientList(token);
  sleep(1);
}
