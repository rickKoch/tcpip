package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/rickKoch/tcpip/net"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	devMgr := net.New()
	defer devMgr.Close()

	f, err := os.CreateTemp("", "dummy.dev")
	if err != nil {
		panic(err)
	}
	defer os.Remove(f.Name())

	fmt.Println("FILENAME::", f.Name())


	dummyDev, err := net.NewDummyDevice(f.Name())
	if err != nil {
		panic(err)
	}

	if err := devMgr.Register(ctx, dummyDev); err != nil {
		panic(err)
	}

	f1, err := os.CreateTemp("", "dummy1.dev")
	if err != nil {
		panic(err)
	}
	defer os.Remove(f1.Name())

	fmt.Println("FILENAME::", f1.Name())


	dummyDev, err = net.NewDummyDevice(f1.Name())
	if err != nil {
		panic(err)
	}

	if err := devMgr.Register(ctx, dummyDev); err != nil {
		panic(err)
	}

	go func() {
		for {
			select {
			case payload := <-devMgr.Queue():
				if string(payload) != "" {
					fmt.Printf("Payload: %s", payload)
				}
			case err := <-devMgr.Errors():
				if err != io.EOF {
					fmt.Println("ERROR::", err)
				}
			}
		}
	}()

	<-ctx.Done()
}
