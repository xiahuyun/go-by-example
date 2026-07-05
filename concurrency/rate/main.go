package main

import (
	"context"
	"log"
	"os"
	"sync"

	"golang.org/x/time/rate"
)

func Open() *APIConnection {
	return &APIConnection{
		rateLimiter: rate.NewLimiter(rate.Limit(1), 1), //1
	}
}

type APIConnection struct {
	rateLimiter *rate.Limiter
}

func (a *APIConnection) ReadFile(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil { //2
		return err
	}

	// Pretend we do work here
	log.Printf("ReadFile")
	return nil
}

func (a *APIConnection) ResolveAddress(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil { //2
		return err
	}

	// Pretend we do work here
	log.Printf("ResolveAddress")
	return nil
}

func main() {
	defer log.Printf("Done.")

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	apiConnection := Open()
	var wg sync.WaitGroup
	wg.Add(20)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := apiConnection.ReadFile(context.Background())
			if err != nil {
				log.Printf("cannot ReadFile: %v", err)
			}
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := apiConnection.ResolveAddress(context.Background())
			if err != nil {
				log.Printf("cannot ResolveAddress: %v", err)

			}
		}()
	}

	wg.Wait()
}
