package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/abubakar508/loadtester/internal/db"
	loadtest "github.com/abubakar508/loadtester/internal/loadstest"
	"github.com/abubakar508/loadtester/internal/mail"
	"github.com/joho/godotenv"
)

type TestRequest struct {
	URL         string `json:"url"`
	Count       int    `json:"count"`
	Concurrency int    `json:"concurrency"`
}

type TestResponse struct {
	Message string `json:"message"`
}

func main() {
	// Load environment variables from .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable required")
	}

	store, err := db.NewStore(dsn)
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}

	mailer := mail.NewMailer()

	http.HandleFunc("/start-test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req TestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		if req.Count <= 0 || req.Concurrency <= 0 || req.URL == "" {
			http.Error(w, "Invalid parameters", http.StatusBadRequest)
			return
		}

		go func() {
			result, err := loadtest.RunLoadTest(req.URL, req.Count, req.Concurrency)
			if err != nil {
				log.Printf("Load test error: %v", err)
				return
			}

			// Save test results to DB
			err = store.SaveTestResult(context.Background(), req.URL, req.Count, req.Concurrency, result.Duration.Milliseconds(), result.SuccessCount)
			if err != nil {
				log.Printf("Failed to save test result: %v", err)
			}

			// Calculate metrics
			successRate := float64(result.SuccessCount) / float64(result.TotalCount) * 100
			failureCount := result.TotalCount - result.SuccessCount
			requestsPerSecond := float64(result.TotalCount) / result.Duration.Seconds()
			avgResponseTimeMs := float64(result.Duration.Milliseconds()) / float64(result.TotalCount) // Approximate avg response time

			// Compose email subject and body
			subject := fmt.Sprintf("Load Test Report: %s", req.URL)
			body := fmt.Sprintf(
				`Load Test Completed Successfully!

Target URL: %s

Summary:
- Total Requests Sent: %d
- Concurrency Level: %d
- Successful Responses: %d
- Failed Responses: %d
- Success Rate: %.2f%%
- Total Test Duration: %s
- Average Requests per Second: %.2f req/s
- Approximate Average Response Time: %.2f ms

Detailed Analysis:
- The server handled approximately %.2f requests per second.
- Failure rate of %.2f%% indicates potential issues under load.
- Concurrency level of %d means %d requests were sent simultaneously.
- A success rate above 90%% suggests the server is stable under this load.
- Lower success rates may indicate saturation, network issues, or rate limiting.

Recommendations:
- If failures occurred, check server logs and resource usage.
- Consider scaling infrastructure or optimizing application code.
- Use caching, CDNs, and load balancers to improve performance.
- Repeat tests with different concurrency and counts to find thresholds.

Thank you for using LoadTester!

--
LoadTester Automated Report
`,
				req.URL,
				result.TotalCount,
				req.Concurrency,
				result.SuccessCount,
				failureCount,
				successRate,
				result.Duration.String(),
				requestsPerSecond,
				avgResponseTimeMs,
				requestsPerSecond,
				float64(failureCount)*100/float64(result.TotalCount),
				req.Concurrency,
				req.Concurrency,
			)

			// Send email report
			toEmail := "abubakarismail508@gmail.com" // Change to your email address
			if err := mailer.SendEmail(toEmail, subject, body); err != nil {
				log.Printf("Failed to send email: %v", err)
			} else {
				log.Printf("Load test report email sent successfully to %s", toEmail)
			}
		}()

		resp := TestResponse{
			Message: fmt.Sprintf("Load test started for %s with %d requests at concurrency %d", req.URL, req.Count, req.Concurrency),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
