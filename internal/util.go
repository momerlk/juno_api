package internal 

import (
	"fmt"
	"os"
	"time"

	"crypto/sha256"

	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"
)

func Getenv(key string) string{
	return os.Getenv(key)
}

func Hash(s string) string {
	return fmt.Sprintf("%x" , sha256.Sum256([]byte(s)))
}

func HashAndSalt(b []byte) (string , error) {
	hash , err := bcrypt.GenerateFromPassword(b , bcrypt.MinCost)
	if err != nil {
		return "" , err
	}
	
	return string(hash) , nil
}


func GenerateId() string {
	return uuid.NewString()
}

// Stopwatch struct to hold start time and elapsed time
type Stopwatch struct {
	start   time.Time
	elapsed time.Duration
	running bool
}

// Start method to start the stopwatch
func (s *Stopwatch) Start() {
	if !s.running {
		s.start = time.Now()
		s.running = true
	}
}

// Stop method to stop the stopwatch and update elapsed time
func (s *Stopwatch) Stop() {
	if s.running {
		s.elapsed += time.Since(s.start)
		s.running = false
	}
}

// Reset method to reset the stopwatch
func (s *Stopwatch) Reset() {
	s.start = time.Now()
	s.elapsed = 0
	if !s.running {
		s.running = true
	}
}

// Elapsed method to get the elapsed time
func (s *Stopwatch) Elapsed() time.Duration {
	if s.running {
		return s.elapsed + time.Since(s.start)
	}
	return s.elapsed
}