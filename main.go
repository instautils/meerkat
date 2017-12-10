package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/ahmdrz/meerkat/cmd/meerkat"
)

func main() {
	run()
}

func run() {
	m, err := meerkat.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", err.Error())
		os.Exit(1)
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, os.Interrupt)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		err = m.Run(done)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[meerkat] Error, %s\n", err.Error())
		}
		fmt.Fprintf(os.Stderr, "[meerkat] Logging out from Instagram , please wait ...\n")
		m.Logout()
		wg.Done()
	}()

	go func() {
		sig := <-sigs
		fmt.Fprintf(os.Stderr, "\n[meerkat] Signal %s detected , please wait ...\n", sig.String())
		done <- true

		wg.Done()
	}()

	wg.Wait()

	fmt.Fprintf(os.Stdout, "[meerkat] Finished ! \n")
}
