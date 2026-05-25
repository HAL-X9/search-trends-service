package usecases

import (
	"sync"
	"time"
)

type AntiFraudDetector struct {
	mu           sync.RWMutex
	userRequests map[string]int
	lastReset    time.Time
	maxLimit     int
}

func NewAntiFraudDetector() *AntiFraudDetector {
	return &AntiFraudDetector{
		userRequests: make(map[string]int),
		lastReset:    time.Now(),
		maxLimit:     5,
	}
}

// IsSpam проверяет не превысил ли лимит данный пользователь или IP-адрес
func (af *AntiFraudDetector) IsSpam(event SearchEvent) bool {
	af.mu.Lock()
	defer af.mu.Unlock()

	if time.Since(af.lastReset) >= time.Second {
		af.userRequests = make(map[string]int)
		af.lastReset = time.Now()
	}

	clientKey := event.UserID + ":" + event.Query
	if event.UserID == "" {
		clientKey = event.IPAddress + ":" + event.Query
	}

	af.userRequests[clientKey]++

	if af.userRequests[clientKey] > af.maxLimit {
		return true
	}

	return false
}
