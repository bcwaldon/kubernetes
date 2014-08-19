package coreos_cloud

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/coreos/fleet/client"
	"github.com/coreos/fleet/etcd"
	"github.com/coreos/fleet/machine"
	"github.com/coreos/fleet/registry"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/cloudprovider"
)

func init() {
	cloudprovider.RegisterCloudProvider("coreos", func(config io.Reader) (cloudprovider.Interface, error) { return NewCoreOSCloud() })
}

type CoreOSCloud struct {
	fleet client.API

	// CoreOSCloud does not implement the full cloudprovider.Interface.
	// Embed the interface for the method signatures.
	cloudprovider.Interface
}

func NewCoreOSCloud() (*CoreOSCloud, error) {
	machines := []string{"http://127.0.0.1:4001"}
	transport := http.Transport{}
	timeout := 5 * time.Second
	eClient, err := etcd.NewClient(machines, transport, timeout)
	if err != nil {
		return nil, err
	}

	reg := registry.New(eClient, "/_coreos.com/fleet")
	rc := client.RegistryClient{reg}
	cc := CoreOSCloud{fleet: &rc}
	return &cc, nil
}

func (core *CoreOSCloud) Instances() (cloudprovider.Instances, bool) {
	return core, true
}

func (core *CoreOSCloud) IPAddress(instance string) (net.IP, error) {
	machines, err := core.fleet.Machines()
	if err != nil {
		return nil, err
	}

	var found *machine.MachineState
	for _, m := range machines {
		if m.PublicIP != instance {
			continue
		}

		m := m
		found = &m
	}

	if found == nil {
		return nil, fmt.Errorf("could not find instance %q", instance)
	}

	ip := net.ParseIP(found.PublicIP)
	if ip == nil {
		return nil, fmt.Errorf("failed parsing machine IP %q: %v", found.PublicIP)
	}

	return ip, nil
}

func (core *CoreOSCloud) List(filter string) ([]string, error) {
	machines, err := core.fleet.Machines()
	if err != nil {
		return nil, err
	}

	out := []string{}
	for _, m := range machines {
		out = append(out, m.PublicIP)
	}

	return out, nil
}
