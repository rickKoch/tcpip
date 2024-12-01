package net

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	netDev "github.com/rickKoch/tcpip/net/device"
)

// deviceManager will manage and register devices
type deviceManager struct {
	sync.RWMutex
	devices map[string]*device
	wg      sync.WaitGroup
	queue   chan []byte
	errors  chan error
}

func New() *deviceManager {
	return &deviceManager{
		devices: make(map[string]*device),
		queue:   make(chan []byte),
		errors:  make(chan error),
	}
}

// Register registers new device
func (mgr *deviceManager) Register(ctx context.Context, dev *device) error {
	if mgr == nil {
		return errors.New("device manager not set")
	}

	mgr.Lock()
	defer mgr.Unlock()
	if _, ok := mgr.devices[dev.name]; ok {
		return errors.New("device already registered")
	}

	mgr.devices[dev.name] = dev

	go func(ctx context.Context, mgr *deviceManager, dev *device) {
		mgr.wg.Add(1)
		defer mgr.wg.Done()

		if err := dev.read(ctx, mgr.queue, mgr.errors); err != nil {
			fmt.Println("failed device read", err)
		}
	}(ctx, mgr, dev)

	return nil
}

func (mgr *deviceManager) Queue() chan []byte {
	return mgr.queue
}

func (mgr *deviceManager) Errors() chan error {
	return mgr.errors
}

func (mgr *deviceManager) Close() {
	mgr.wg.Wait()
	close(mgr.queue)
}

// device represents network device. It provided connection to the network.
// We should be able to handle different and multiple devices in the same time.
type device struct {
	// Name of the device
	name string
	// Maximum Transmission Unit (payload size - IP packet)
	mtu int
	// header size
	headerSize int
	// the raw device
	raw io.ReadWriteCloser
}

func NewDummyDevice(name string) (*device, error) {
	raw, name, err := netDev.OpenDummyDevice(name)
	if err != nil {
		return nil, err
	}

	dev := &device{
		name:       name,
		raw:        raw,
		mtu:        1500,
		headerSize: 14,
	}

	return dev, nil
}

// Read
func (d *device) read(ctx context.Context, queue chan<- []byte, errs chan<- error) error {
	if d == nil {
		return errors.New("device does not exist")
	}

	buf := make([]byte, d.headerSize+d.mtu)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("exiting device read")
			if err := d.close(); err != nil {
				return err
			}

			return nil
		default:
			n, err := d.raw.Read(buf)
			if err != nil {
				errs <- err
			}

			queue <- buf[:n]
		}
	}
}

// Close
func (d *device) close() error {
	if d == nil {
		return errors.New("device does not exists")
	}

	return d.raw.Close()
}
