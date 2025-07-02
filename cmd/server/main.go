package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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

type HealthResponse struct {
	Status      string `json:"status"`
	DBConnected bool   `json:"db_connected"`
	Timestamp   string `json:"timestamp"`
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
	toEmail := "abubakarismail508@gmail.com" // Change to your email address

	// Load test endpoint
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

			err = store.SaveTestResult(context.Background(), req.URL, req.Count, req.Concurrency, result.Duration.Milliseconds(), result.SuccessCount)
			if err != nil {
				log.Printf("Failed to save test result: %v", err)
			}

			successRate := float64(result.SuccessCount) / float64(result.TotalCount) * 100
			failureCount := result.TotalCount - result.SuccessCount
			requestsPerSecond := float64(result.TotalCount) / result.Duration.Seconds()
			avgResponseTimeMs := float64(result.Duration.Milliseconds()) / float64(result.TotalCount)

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

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := store.DB.PingContext(ctx)
		dbConnected := err == nil

		status := "healthy"
		if !dbConnected {
			status = "unhealthy"
		}

		// Compose health email
		subject := "LoadTester Service Health Check"
		body := fmt.Sprintf(
			`Health Check Report:

Service Status: %s
Database Connected: %t
Timestamp: %s

%s

Regards,
LoadTester Monitoring
`,
			status,
			dbConnected,
			time.Now().Format(time.RFC1123),
			func() string {
				if dbConnected {
					return "All systems operational."
				}
				return "Database connection failed! Immediate attention required."
			}(),
		)

		// Send health email asynchronously
		go func() {
			if err := mailer.SendEmail(toEmail, subject, body); err != nil {
				log.Printf("Failed to send health check email: %v", err)
			} else {
				log.Printf("Health check email sent successfully to %s", toEmail)
			}
		}()

		// Respond with JSON health status
		resp := HealthResponse{
			Status:      status,
			DBConnected: dbConnected,
			Timestamp:   time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		if dbConnected {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		json.NewEncoder(w).Encode(resp)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
