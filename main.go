package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	stan "github.com/nats-io/go-nats-streaming"
)

const (
	clusterID = "test-cluster"
	clientID  = "nrm"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintln(os.Stderr, "Specify filter; format: <subject>:<sequence number>, only subject is required")
		os.Exit(1)
	}

	filters := strings.SplitN(os.Args[1], ":", 2)

	subject := filters[0]
	var seq uint64 = 0

	if len(filters) >= 2 {
		var err error
		seq, err = strconv.ParseUint(filters[1], 10, 64)
		check(err)
	}

	sc, err := stan.Connect(clusterID, clientID)
	check(err)
	defer sc.Close()

	var subFirst, subLast, sub stan.Subscription
	var seqFirst, seqLast uint64

	if seq == 0 {
		var wg sync.WaitGroup

		wg.Add(1)
		subFirst, err = sc.Subscribe(subject, func(m *stan.Msg) {
			defer wg.Done()
			// fmt.Printf("FIRST Received message %d: %s\n", m.Sequence, string(m.Data))
			seqFirst = m.Sequence
			subFirst.Unsubscribe()
		}, stan.DeliverAllAvailable())
		check(err)

		wg.Add(1)
		subLast, err = sc.Subscribe(subject, func(m *stan.Msg) {
			defer wg.Done()
			// fmt.Printf("LAST Received message %d: %s\n", m.Sequence, string(m.Data))
			seqLast = m.Sequence
			subLast.Unsubscribe()
		}, stan.StartWithLastReceived())
		check(err)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(1 * time.Second):
			fmt.Fprintln(os.Stderr, "Timeout: Could not detect sequence range")
			os.Exit(1)
		}

		var cnt int64 = int64(seqLast - seqFirst)
		if cnt > 0 {
			seq = seqFirst + uint64(rand.Int63n(cnt))
		} else {
			seq = 1
		}
	}

	done := make(chan struct{})
	sub, err = sc.Subscribe(subject, func(m *stan.Msg) {
		defer close(done)

		if seqFirst > 0 || seqLast > 0 {
			fmt.Fprintf(os.Stderr, "Seq %d-%d; ", seqFirst, seqLast)
		}
		fmt.Fprintf(os.Stderr, "Picked: %s:%d\n", subject, m.Sequence)
		os.Stdout.Write(m.Data)

		sub.Unsubscribe()
	}, stan.StartAtSequence(seq), stan.SetManualAckMode(), stan.MaxInflight(1))
	check(err)

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		fmt.Fprintln(os.Stderr, "Timeout: No message received for given filter")
		os.Exit(1)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
