package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ahmdrz/meerkat/cmd/meerkat"
)

var configTemplate = `
username: "#####"
password: "#####"

# interval
# in seconds.
interval: 15 

# sleeptime
# in seconds.
# after each request to get user's information , 
# we have to sleep , because instagram may ban our account.
sleeptime: 10

# output types: choose how you wants to know about users activity.
# types are : ["logfile", "telegram"]
# you can select multiple options using ',' seprator. ex. "telegram,logfile"
outputtype: "logfile"

# telegram bot token
# fill this field if you choose telegram in outputtype.
telegramtoken: "###"

# telegram id from user
# get it using @userinfobot on telegram
telegramuser: 0

targetusers: 
  - "###"
`

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "init" {
			configFile := "meerkat.yaml"
			if len(os.Args) > 2 {
				configFile = os.Args[2]
			}
			file, err := os.OpenFile(configFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			file.WriteString(configTemplate)
			file.Close()

			fmt.Println(configFile, "generated.")

			os.Exit(0)
		}
	}

	run()
}

func run() {
	m, err := meerkat.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", err.Error())
		os.Exit(1)
	}

	err = m.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", err.Error())
		return
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println(sig)
		done <- true
	}()

	<-done

	m.Logout()
}
