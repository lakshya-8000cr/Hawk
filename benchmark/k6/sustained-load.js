import http from "k6/http";
import { check } from "k6";

const targetURL = __ENV.TARGET_URL;

if (!targetURL) {
  throw new Error("TARGET_URL is required");
}

export const options = {
  scenarios: {
    sustained_load: {
      executor: "constant-vus",
      vus: Number(__ENV.VUS || 50),
      duration: __ENV.DURATION || "60s",
    },
  },

  thresholds: {
    http_req_failed: ["rate<0.01"],
    checks: ["rate>0.99"],
  },
};

export default function () {
  const response = http.get(targetURL, {
    timeout: "10s",
  });

  check(response, {
    "status is 200": (res) => res.status === 200,
    "body is expected": (res) =>
      typeof res.body === "string" &&
      res.body.includes("hawk benchmark endpoint"),
  });
}