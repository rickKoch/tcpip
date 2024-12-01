package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rickKoch/tcpip/net"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	devMgr := net.New()
	defer devMgr.Close()

	loopbackDev, err := net.NewLoopbackDevice("loopback0")
	if err != nil {
		panic(err)
	}

	if err := devMgr.Register(ctx, loopbackDev); err != nil {
		panic(err)
	}

	go func() {
		for {
			time.Sleep(5 * time.Second)
			if err := devMgr.Write("loopback0", []byte("testing loopback")); err != nil {
				fmt.Println("ERROR", err)
			}
		}
	}()

	f, err := os.CreateTemp("", "dummy1.dev")
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
