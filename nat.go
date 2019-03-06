// Package nat implements NAT handling facilities
package nat

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"net"
	"time"
)

var ErrNoExternalAddress = errors.New("no external address")
var ErrNoInternalAddress = errors.New("no internal address")
var ErrNoNATFound = errors.New("no NAT found")

// protocol is either "udp" or "tcp"
type NAT interface {
	// Type returns the kind of NAT port mapping service that is used
	Type() string

	// GetDeviceAddress returns the internal address of the gateway device.
	GetDeviceAddress() (addr net.IP, err error)

	// GetExternalAddress returns the external address of the gateway device.
	GetExternalAddress() (addr net.IP, err error)

	// GetInternalAddress returns the address of the local host.
	GetInternalAddress() (addr net.IP, err error)

	// AddPortMapping maps a port on the local host to an external port.
	AddPortMapping(protocol string, internalPort int, description string, timeout time.Duration) (mappedExternalPort int, err error)

	// DeletePortMapping removes a port mapping.
	DeletePortMapping(protocol string, internalPort int) (err error)
}

// DiscoverNATs returns all NATs discovered in the network.
func DiscoverNATs(ctx context.Context) <-chan NAT {
	nats := make(chan NAT)

	go func() {
		defer close(nats)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		upnpIg1 := discoverUPNP_IG1(ctx)
		upnpIg2 := discoverUPNP_IG2(ctx)
		natpmp := discoverNATPMP(ctx)
		upnpGenIGDev := discoverUPNP_GenIGDev(ctx)
		for upnpIg1 != nil || upnpIg2 != nil || natpmp != nil {
			var (
				nat NAT
				ok  bool
			)
			select {
			case nat, ok = <-upnpIg1:
				if !ok {
					upnpIg1 = nil
				}
			case nat, ok = <-upnpIg2:
				if !ok {
					upnpIg2 = nil
				}
			case nat, ok = <-upnpGenIGDev:
				if !ok {
					upnpGenIGDev = nil
				}
			case nat, ok = <-natpmp:
				if !ok {
					natpmp = nil
				}
			}
			if ok {
				select {
				case nats <- nat:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return nats
}

// DiscoverGateway attempts to find a gateway device.
func DiscoverGateway() (NAT, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	nat := <-DiscoverNATs(ctx)
	if nat == nil {
		return nil, ErrNoNATFound
	}
	return nat, nil
}

func randomPort() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(math.MaxUint16-10000) + 10000
}
