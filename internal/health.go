package internal

import (
	"context"
	"fmt"
	"load-balancer/config"
	"log"
	"net/http"
	"time"
)

func (lb *LoadBalancer) checkHealthOfBackend(backend *Backend) {
	old := backend.IsAlive()
	resp, err := http.Get(backend.url.String() + "/health")

	if err != nil {
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
		lb.checkHealthOfBackend(lb.servers[i])
	}
	ticker := time.NewTicker(config.Health.Interval)
	defer ticker.Stop()


	select {
		case <-ctx.Done(): 
				fmt.Println("Stopping heath checkup")
				return
		case <-ticker.C:
			for range ticker.C {
				for i := range lb.servers {
					lb.checkHealthOfBackend(lb.servers[i])
				}
	 		}
	}
	
}