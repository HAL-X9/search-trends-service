package usecases

import (
	"testing"
	"time"
)

func TestAntiFraudDetector_IsSpam(t *testing.T) {
	af := NewAntiFraudDetector()
	event := SearchEvent{
		UserID:    "user1",
		Query:     "nike",
		IPAddress: "127.0.0.1",
	}

	for i := 0; i < 5; i++ {
		if af.IsSpam(event) {
			t.Fatalf("request %d should not be marked as spam", i+1)
		}
	}

	if !af.IsSpam(event) {
		t.Error("expected 6th request to be marked as spam")
	}
}

func TestAntiFraudDetector_Reset(t *testing.T) {
	af := NewAntiFraudDetector()
	event := SearchEvent{UserID: "user2", Query: "adidas"}

	for i := 0; i < 5; i++ {
		af.IsSpam(event)
	}

	af.mu.Lock()
	af.lastReset = time.Now().Add(-2 * time.Second)
	af.mu.Unlock()

	if af.IsSpam(event) {
		t.Error("expected anti-fraud counter to reset after 1 second")
	}
}
