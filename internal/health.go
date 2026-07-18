package internal

import (
	"context"
	"load-balancer/config"
	"log"
	"net/http"
	"time"
)

func (lb *LoadBalancer) checkHealthOfBackend(ctx context.Context, backend *Backend) {
	old := backend.IsAlive()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, backend.url.String()+"/health", nil)
	if err != nil {
		log.Printf("failed to create health-check request for %s: %v", backend.url.String(), err)
		return
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		// Cancellation is expected during shutdown, not a backend failure.
		if ctx.Err() != nil {
			return
		}
		if old {
			backend.SetAlive(false)
			log.Printf("Backend server %s is DOWN\n", backend.url.String())
		}
		return
	}
	defer resp.Body.Close()

	new := resp.StatusCode == http.StatusOK

	statusStr := func(status bool) string {
		if status {
			return "UP"
		}
		return "DOWN"
	}

	if old != new {
		backend.SetAlive(new)
		log.Printf("Backend server %s is %v\n", backend.url.String(), statusStr(new))
	}
}

func (lb *LoadBalancer) HealthCheckLoop(config config.Config, ctx context.Context) {
	for i := range lb.servers {
		if ctx.Err() != nil {
			return
		}
		lb.checkHealthOfBackend(ctx, lb.servers[i])
	}
	ticker := time.NewTicker(config.Health.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping health checkup")
			return

		case <-ticker.C:
			for i := range lb.servers {
				if ctx.Err() != nil {
					log.Println("Stopping health checkup")
					return
				}
				lb.checkHealthOfBackend(ctx, lb.servers[i])
			}
		}
	}
}
