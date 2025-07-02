# Load Test

## Overview

**Load Test** is a simple, safe, and configurable HTTP load testing tool written in Go. It is designed to help developers and system administrators evaluate the performance and resilience of their web applications by simulating concurrent HTTP requests to a specified URL.

Unlike aggressive denial-of-service (DoS) attacks, Load Test focuses on **safe load testing** practices that allow you to:

- Measure how your server behaves under various levels of concurrent traffic.
- Identify performance bottlenecks and potential points of failure.
- Store and analyze test results for continuous improvement.

This tool is ideal for developers who want to understand their application's limits responsibly and prepare for real-world traffic surges.

## Features

- **Configurable test parameters:** Specify the target URL, total number of requests, and concurrency level.
- **Concurrency control:** Uses worker pools to manage the number of simultaneous requests, preventing resource exhaustion on the client side.
- **Timeout and error handling:** Implements HTTP client timeouts and gracefully handles network errors, including timeouts.
- **Result persistence:** Saves test results (such as request count, concurrency, duration, and success rate) into a PostgreSQL database.
- **REST API interface:** Exposes an HTTP endpoint to trigger load tests dynamically by sending JSON payloads.
- **Dockerized:** Ready to deploy in containerized environments like Render.com or other cloud platforms.

## How It Works

1. **Trigger a test:** Send a POST request to the `/start-test` endpoint with JSON specifying:
   - `url`: The target domain or endpoint to test.
   - `count`: Total number of HTTP GET requests to send.
   - `concurrency`: Number of concurrent workers sending requests simultaneously.

2. **Run the load test:** The server runs the test asynchronously, sending the specified number of requests with controlled concurrency.

3. **Collect results:** The load tester measures the total time taken and counts successful HTTP responses.

4. **Save results:** Test metadata and results are saved into a PostgreSQL database for later analysis.

5. **Monitor and analyze:** Use the stored data to understand your system’s behavior under load and plan scaling or optimizations.

## Getting Started

### Prerequisites

- Go 1.20 or later
- PostgreSQL database
- Docker (optional, for containerized deployment)
- Environment variable `DATABASE_URL` set with your PostgreSQL connection string

### Running Locally

1. Clone the repository:

git clone https://github.com/abubakar508/loadtest.git


2. Set your PostgreSQL connection string:


export DATABASE_URL="postgres://user:password@host:port/dbname?sslmode=disable"

3. Build and run the server:

go build -o loadtest ./cmd/server
./loadtester

4. Trigger a load test with curl:

curl -X POST http://localhost:8080/start-test
-H "Content-Type: application/json"
-d '{"url":"https://yourdomain.com","count":1000,"concurrency":20}'



### Deploying on Render.com

1. Push your code to GitHub.
2. Create a new Web Service on Render, connecting your GitHub repo.
3. Set build command: `go build -o loadtester ./cmd/server`
4. Set start command: `./loadtester`
5. Add environment variable `DATABASE_URL` with your PostgreSQL connection string.
6. Deploy and monitor logs.
7. Use the `/start-test` API to run load tests remotely.
8. {
  "url": "https://domain.com",
  "count": 20,
  "concurrency": 1000 //number of requests per seconnd i.e in a count of 20 requests 1000 are seotn simlataneously in each
}

## Important Notes

- **Safe testing only:** This tool is intended for authorized testing on your own infrastructure or environments where you have explicit permission.
- **Avoid production impact:** Start with low concurrency and request counts, gradually increasing while monitoring your system.
- **Not a DoS attack tool:** Designed for responsible load testing, not for malicious denial-of-service attacks.
- **Monitor resources:** Always watch CPU, memory, network, and database performance during tests.

## Contributing

Contributions, bug reports, and feature requests are welcome! Please open issues or submit pull requests on GitHub.

*Developed by Abubakar Ismail*
*Empowering developers with safe and practical load testing tools.*
