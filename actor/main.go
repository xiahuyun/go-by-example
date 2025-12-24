package main

import (
	"fmt"
	"github.com/oklog/run"
	"time"
)

func main() {
	g := run.Group{}
	{
		cancel := make(chan struct{})
		g.Add(
			func() error {

				select {
				case <-cancel:
					fmt.Println("Go routine 1 is closed")
					break
				}

				return nil
			},
			func(error) {
				fmt.Println("Go routine 1 is exit")
				close(cancel)
			},
		)
	}
	{
		cancel := make(chan struct{})
		g.Add(
			func() error {

				select {
				case <-cancel:
					fmt.Println("Go routine 2 is closed")
					break
				}

				return nil
			},
			func(error) {
				fmt.Println("Go routine 2 is exit")
				close(cancel)
			},
		)
	}
	{
		g.Add(
			func() error {
				for i := 0; i <= 3; i++ {
					time.Sleep(1 * time.Second)
					fmt.Println("Go routine 3 is sleeping...")
				}
				fmt.Println("Go routine 3 is closed")
				return nil
			},
			func(error) {
				fmt.Println("Go routine 3 is exit")
				return
			},
		)
	}
	g.Run()
}
