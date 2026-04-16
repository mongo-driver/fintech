package middleware

import (
	"net"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func RateLimit(rps float64, burst int) gin.HandlerFunc {
	var (
		mu       sync.Mutex
		visitors = map[string]*visitor{}
	)

	go func() {
		t := time.NewTicker(1 * time.Minute)
		for range t.C {
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := clientIP(c.ClientIP())
		mu.Lock()
		v, ok := visitors[ip]
		if !ok {
			v = &visitor{
				limiter: rate.NewLimiter(rate.Limit(rps), burst),
			}
			visitors[ip] = v
		}
		v.lastSeen = time.Now()
		allowed := v.limiter.Allow()
		mu.Unlock()

		if !allowed {
			c.AbortWithStatusJSON(429, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

func clientIP(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ip
	}
	return parsed.String()
}
