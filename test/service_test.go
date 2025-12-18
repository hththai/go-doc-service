package test

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
)

func TestPing(t *testing.T) {
	// e := httpexpect.New(t, "http://api.golang.localdomain/testupload")
	e := httpexpect.New(t, "http://api.golang.localdomain")

	// Get ping
	e.GET("/ping").
		Expect().
		Status(http.StatusOK).JSON().Object().ValueEqual("message", "pong")
}

func TestConcurrentPing(t *testing.T) {
	e := httpexpect.New(t, "http://api.golang.localdomain")

	// Number of concurrent requests.
	const workers = 15

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()

			// Each goroutine sends a request
			e.GET("/ping").Expect().Status(http.StatusOK).JSON().Object().ValueEqual("message", "pong")

			t.Logf("Worker %d finished", id)
		}(i)

	}
	wg.Wait()
}

// func TestUploadWithoutFile(t *testing.T) {
// 	e := httpexpect.New(t, "http://api.golang.localdomain")

// 	e.POST("/testupload").WithMultipart().
// 		WithFormField("name", "Testing by library2323").
// 		Expect().
// 		Status(http.StatusOK).
// 		Body().Contains(`{"message":"Success"}`)
// }

func TestUploadWithoutFilePen(t *testing.T) {
	e := httpexpect.New(t, "http://api.golang.localdomain")

	// Number of concurrent requests.
	const workers = 4

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()

			// Each goroutine sends a request
			e.POST("/testupload").WithMultipart().
				WithFormField("name", fmt.Sprintf("Testing by library %d", i)).
				Expect().
				Status(http.StatusOK).
				Body().Contains(`{"message":"Success"}`)

			t.Logf("Worker %d finished", id)
		}(i)

	}
	wg.Wait()
}

func TestUploadFileError(t *testing.T) {
	e := httpexpect.New(t, "http://api.golang.localdomain")

	// Open the file.
	file, err := os.Open("./Hayden Thai_Resume-2.pdf")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	defer file.Close()

	e.POST("/testupload").WithMultipart().
		WithFormField("name", "Testing with upload").
		WithFile("file", "Hayden Thai_Resume-2.pdf", file).
		Expect().
		Status(http.StatusInternalServerError)
}

func TestUploadFileSuccess(t *testing.T) {
	e := httpexpect.New(t, "http://api.golang.localdomain")

	// Open the file.
	file, err := os.Open("./example.txt")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	defer file.Close()

	e.POST("/testupload").WithMultipart().
		WithFormField("name", "Testing with upload").
		WithFile("file", "example.txt", file).
		Expect().
		Status(http.StatusOK)
}

func TestUploadConcurrentRandom(t *testing.T) {
	e := httpexpect.New(t, "http://api.golang.localdomain")

	const workers = 10
	var wg sync.WaitGroup
	wg.Add(workers)

	// Seed randomness
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()

			// Randomly pick scenario
			if rand.Intn(2) == 0 {
				// Error scenario: upload PDF, expect 500
				file, err := os.Open("./Hayden Thai_Resume-2.pdf")
				if err != nil {
					t.Errorf("worker %d failed to open file: %v", id, err)
					return
				}
				defer file.Close()

				e.POST("/testupload").WithMultipart().
					WithFormField("name", "Testing with upload").
					WithFile("file", "Hayden Thai_Resume-2.pdf", file).
					Expect().
					Status(500)

				t.Logf("Worker %d ran error scenario", id)
			} else {
				// Success scenario: upload TXT, expect 200
				file, err := os.Open("./example.txt")
				if err != nil {
					t.Errorf("worker %d failed to open file: %v", id, err)
					return
				}
				defer file.Close()

				e.POST("/testupload").WithMultipart().
					WithFormField("name", "Testing with upload").
					WithFile("file", "example.txt", file).
					Expect().
					Status(200)

				t.Logf("Worker %d ran success scenario", id)
			}
		}(i)
	}

	wg.Wait()
}

func BenchmarkUploadConcurrentRandom(b *testing.B) {
	e := httpexpect.New(b, "http://api.golang.localdomain")

	// Seed randomness once
	rand.Seed(time.Now().UnixNano())

	// Number of concurrent workers
	const workers = 50

	for n := 0; n < b.N; n++ {
		var wg sync.WaitGroup
		wg.Add(workers)

		for i := 0; i < workers; i++ {
			go func(id int) {
				defer wg.Done()
				b.Logf("Start worker %d", i)
				if rand.Intn(2) == 0 {
					// Error scenario: upload PDF, expect 500
					file, err := os.Open("./Hayden Thai_Resume-2.pdf")
					if err != nil {
						b.Errorf("worker %d failed to open file: %v", id, err)
						return
					}
					defer file.Close()

					e.POST("/testupload").WithMultipart().
						WithFormField("name", "Testing with upload").
						WithFile("file", "Hayden Thai_Resume-2.pdf", file).
						Expect().
						Status(500)
				} else {
					// Success scenario: upload TXT, expect 200
					file, err := os.Open("./example.txt")
					if err != nil {
						b.Errorf("worker %d failed to open file: %v", id, err)
						return
					}
					defer file.Close()

					e.POST("/testupload").WithMultipart().
						WithFormField("name", "Testing with upload").
						WithFile("file", "example.txt", file).
						Expect().
						Status(200)
				}
			}(i)
		}

		wg.Wait()
	}
}
