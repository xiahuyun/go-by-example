package main

import (
	"context"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter interface { //1
	Wait(context.Context) error
	Limit() rate.Limit
}

type multiLimiter struct {
	limiters []RateLimiter
}

func MultiLimiter(limiters ...RateLimiter) *multiLimiter {
	byLimit := func(i, j int) bool {
		return limiters[i].Limit() < limiters[j].Limit()
	}
	sort.Slice(limiters, byLimit) //2
	return &multiLimiter{limiters: limiters}
}

func (l *multiLimiter) Wait(ctx context.Context) error {
	for _, l := range l.limiters {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (l *multiLimiter) Limit() rate.Limit {
	return l.limiters[0].Limit() //3
}

type APIConnection struct {
	rateLimiter RateLimiter
}

func Per(eventCount int, duration time.Duration) rate.Limit {
	return rate.Every(duration / time.Duration(eventCount))
}

func Open() *APIConnection {
	secondLimit := rate.NewLimiter(Per(2, time.Second), 1)   //1
	minuteLimit := rate.NewLimiter(Per(10, time.Minute), 10) //2

	return &APIConnection{
		rateLimiter: MultiLimiter(secondLimit, minuteLimit), //3
	}
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
