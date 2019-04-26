package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/consul/api"
)

var isLeader bool

func main() {
	go startAPI()

	// ttl in seconds
	ttl := 10
	ttlS := fmt.Sprintf("%ds", ttl)
	serviceKey := "service/distributed-logger/leader"
	serviceName := "distributed-logger"

	// build client
	config := api.DefaultConfig()
	config.Address = "consul:8500"
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatalf("client err: %v", err)
	}

	// create session
	sEntry := &api.SessionEntry{
		Name:      serviceName,
		TTL:       ttlS,
		LockDelay: 1 * time.Millisecond,
	}
	sID, _, err := client.Session().Create(sEntry, nil)
	if err != nil {
		log.Fatalf("session create err: %v", err)
	}

	// auto renew session
	doneCh := make(chan struct{})
	go func() {
		err = client.Session().RenewPeriodic(ttlS, sID, nil, doneCh)
		if err != nil {
			log.Fatalf("session renew err: %v", err)
		}
	}()

	log.Printf("Consul session : %+v\n", sID)

	// Lock acquisition loop
	go func() {
		hostName, err := os.Hostname()
		if err != nil {
			log.Fatalf("hostname err: %v", err)
		}

		acquireKv := &api.KVPair{
			Session: sID,
			Key:     serviceKey,
			Value:   []byte(hostName),
		}

		for {
			if !isLeader {
				acquired, _, err := client.KV().Acquire(acquireKv, nil)
				if err != nil {
					log.Fatalf("kv acquire err: %v", err)
				}

				if acquired {
					isLeader = true
					log.Printf("I'm the leader !\n")
				}
			}

			time.Sleep(time.Duration(ttl/2) * time.Second)
		}
	}()

	// wait for SIGINT or SIGTERM, clean up and exit
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	close(doneCh)
	log.Printf("Destroying session and leaving ...")
	_, err = client.Session().Destroy(sID, nil)
	if err != nil {
		log.Fatalf("session destroy err: %v", err)
	}
	os.Exit(0)
}

func startAPI() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/log", func(w http.ResponseWriter, r *http.Request) {
		if !isLeader {
			http.Error(w, "Not Leader", http.StatusBadRequest)
			return
		}

		msg, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		// log msg
		log.Printf("Received %v", string(msg))

		w.Write([]byte("OK"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	port := "3000"
	log.Printf("Starting API on port %s ....\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
