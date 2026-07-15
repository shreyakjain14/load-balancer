package main

import (
	"context"
	"load-balancer/config"
	"load-balancer/internal"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	con, err := config.GetConfig("../../config/config.yaml")

	if err != nil {
		log.Fatal("Not able to load config file", err)
	}

	lb := internal.NewLoadBalancer()

	if len(con.Backends) == 0 {
    	log.Fatal("at least one backend required")
	}

	for i := range con.Backends {
		err := lb.AddBackend(con.Backends[i].URL)

		if err != nil {
			log.Printf("failed to add backend %v", con.Backends[i].URL)
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/", lb.Handler())
	mux.HandleFunc("/metrics", lb.Metrics.Handler)

	server := &http.Server{
		Addr:         ":" + con.Server.Port,
		Handler:      mux,
		ReadTimeout:  10 * con.Server.ReadTimeout,
		WriteTimeout: 10 * con.Server.WriteTimeout,
	}

	errCh := make(chan error, 1)

	go func() {
    	errCh <- server.ListenAndServe()
	}()



	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go lb.HealthCheckLoop(con, ctx)

	select {
		case err := <-errCh:
    		if err != nil && err != http.ErrServerClosed {
        	log.Fatal(err)
    	}

		case <-ctx.Done():
    		log.Println("Shutdown signal received")
	}

	shutDownCtx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	if err := server.Shutdown(shutDownCtx); err != nil {
		log.Println(err)
	}


	log.Println("Server stopped gracefully")
}
