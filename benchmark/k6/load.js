import http from "k6/http";
import { check } from "k6";

const totalRequests = Number(__ENV.REQUESTS || 1000);
const virtualUsers = Number(__ENV.VUS || 10);
const targetURL = __ENV.TARGET_URL;

if (!targetURL) {
  throw new Error(
    "TARGET_URL is required. Example: -e TARGET_URL=http://127.0.0.1:6823",
  );
}

export const options = {
  scenarios: {
    exact_request_count: {
      executor: "shared-iterations",
      vus: virtualUsers,
      iterations: totalRequests,
      maxDuration: "5m",
    },
  },

  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<500"],
    checks: ["rate>0.99"],
  },
};

export default function () {
  const response = http.get(targetURL, {
    tags: {
      endpoint: "hawk-benchmark",
    },
    timeout: "10s",
  });

  check(response, {
    "status is 200": (res) => res.status === 200,
    "body is expected": (res) =>
      res.body.includes("hawk benchmark endpoint"),
  });
}