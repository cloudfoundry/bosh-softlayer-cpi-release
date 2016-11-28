package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"time"
)

type Route struct {
	Network string
	Netmask string
	Gateway string
}

type Interface struct {
	Name           string
	Auto           bool
	AllowHotplug   bool
	Address        string
	Netmask        string
	DefaultGateway bool
	Gateway        string
	Routes         []Route
	DNS            []string
}

type Interfaces []Interface

// first 2 and last reserved if secondary
type Subnet struct {
	NetworkIdentifier string `json:"networkIdentifier"`
	Gateway           string `json:"gateway"`
	BroadcastAddress  string `json:"broadcastAddress"`
	Netmask           string `json:"netmask"`
}

func (s Subnet) contains(address string) bool {
	ipNet := net.IPNet{
		IP:   net.ParseIP(s.NetworkIdentifier),
		Mask: net.IPMask(net.ParseIP(s.Netmask)),
	}

	return ipNet.Contains(net.ParseIP(address))
}

type Subnets []Subnet

func (s Subnets) containing(address string) (Subnet, error) {
	for _, subnet := range s {
		if subnet.contains(address) {
			return subnet, nil
		}
	}

	return Subnet{}, fmt.Errorf("subnet not found for %q", address)
}

type NetworkVLAN struct {
	Name    string  `json:"name"`
	Subnets Subnets `json:"subnets"`
}

type NetworkComponent struct {
	Name             string      `json:"name"`
	Port             int         `json:"port"`
	PrimaryIPAddress string      `json:"primaryIpAddress"`
	NetworkVLAN      NetworkVLAN `json:"networkVlan"`
}

type VirtualGuestNetworkComponents struct {
	PrimaryBackendNetworkComponent NetworkComponent `json:"primaryBackendNetworkComponent,omitempty"`
	PrimaryNetworkComponent        NetworkComponent `json:"primaryNetworkComponent,omitempty"`
}

//go:generate counterfeiter -o fakes/fake_softlayer_client.go . softLayerClient
type softLayerClient interface {
	DoRawHttpRequest(path string, requestType string, requestBody *bytes.Buffer) ([]byte, int, error)
}

//go:generate counterfeiter -o fakes/fake_sl_file_service.go --fake-name FakeSLFileService . softLayerFileService
type softLayerFileService interface {
	Upload(user string, password string, target string, destinationPath string, contents []byte) error
}

//go:generate counterfeiter -o fakes/fake_ssh_client.go --fake-name FakeSSHClient . sshClient
type sshClient interface {
	Output(cmd string) ([]byte, error)
}

type Ubuntu struct {
	SoftLayerClient      softLayerClient
	SSHClient            sshClient
	SoftLayerFileService softLayerFileService
}

func SoftlayerPrivateRoutes(gateway string) []Route {
	return []Route{
		{Network: "10.0.0.0", Netmask: "255.0.0.0", Gateway: gateway},
		{Network: "161.26.0.0", Netmask: "255.255.0.0", Gateway: gateway},
	}
}

func (u *Ubuntu) ConfigureNetwork(networks Networks, vm VM) error {
	interfaces, err := u.GetInterfaces(networks, vm.ID())
	if err != nil {
		return err
	}

	config, err := interfaces.Configuration()
	if err != nil {
		return err
	}

	timeout := 5 * time.Minute
	pollingInterval := 15 * time.Second

	totalTime := time.Duration(0)
	for totalTime < timeout {
		err = u.SoftLayerFileService.Upload("root", vm.GetRootPassword(), vm.GetPrimaryBackendIP(), "/etc/network/interfaces.bosh", config)
		if err == nil {
			break
		}

		totalTime += pollingInterval
		time.Sleep(pollingInterval)
	}

	if err != nil {
		return err
	}

	_, err = u.SSHClient.Output("bash -c 'ifdown -a && mv /etc/network/interfaces.bosh /etc/network/interfaces && ifup -a'")
	if err != nil {
		return fmt.Errorf("nework configuration reload failed: %s", err)
	}

	return nil
}

func (u *Ubuntu) GetInterfaces(networks Networks, virtualGuestId int) (Interfaces, error) {
	path := fmt.Sprintf("SoftLayer_Virtual_Guest/%d/getObject?objectMask=mask[primaryBackendNetworkComponent.networkVlan.subnets,primaryNetworkComponent.networkVlan.subnets]", virtualGuestId)
	response, responseCode, err := u.SoftLayerClient.DoRawHttpRequest(path, "GET", &bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	if responseCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response code: %d", responseCode)
	}

	var networkComponents VirtualGuestNetworkComponents
	err = json.Unmarshal(response, &networkComponents)
	if err != nil {
		return nil, err
	}

	dynamic, manual, err := categorizeNetworks(networks)
	if err != nil {
		return nil, err
	}

	dynamicInterfaces, err := u.dynamicInterfaces(networkComponents, dynamic)
	if err != nil {
		return nil, err
	}

	manualInterfaces, err := u.manualInterfaces(networkComponents, manual)
	if err != nil {
		return nil, err
	}

	return append(dynamicInterfaces, manualInterfaces...), nil
}

func categorizeNetworks(networks Networks) (Networks, Networks, error) {
	dynamic := Networks{}
	manual := Networks{}

	for name, nw := range networks {
		switch nw.Type {
		case "dynamic":
			dynamic[name] = nw
		case "manual", "":
			manual[name] = nw
		default:
			return nil, nil, fmt.Errorf("unexpected network type: %s", nw.Type)
		}
	}

	return dynamic, manual, nil
}

func (u *Ubuntu) dynamicInterfaces(networkComponents VirtualGuestNetworkComponents, dynamic Networks) ([]Interface, error) {
	if len(dynamic) != 1 {
		return nil, errors.New("virtual guests must have exactly one dynamic network")
	}

	nw := dynamic.First()
	privateComponent := networkComponents.PrimaryBackendNetworkComponent
	publicComponent := networkComponents.PrimaryNetworkComponent

	subnet, err := privateComponent.NetworkVLAN.Subnets.containing(privateComponent.PrimaryIPAddress)
	if err != nil {
		err = fmt.Errorf("%s: privateComponent: %#v", err, privateComponent)
		return nil, err
	}

	privateInterface := Interface{
		Name:           fmt.Sprintf("%s%d", privateComponent.Name, privateComponent.Port),
		Auto:           true,
		AllowHotplug:   true,
		Address:        privateComponent.PrimaryIPAddress,
		Netmask:        subnet.Netmask,
		Gateway:        subnet.Gateway,
		DefaultGateway: (publicComponent.PrimaryIPAddress == "" && nw.HasDefaultGateway()),
		Routes:         SoftlayerPrivateRoutes(subnet.Gateway),
	}
	interfaces := []Interface{privateInterface}

	if publicComponent.PrimaryIPAddress != "" {
		for _, s := range publicComponent.NetworkVLAN.Subnets {
			if s.contains(publicComponent.PrimaryIPAddress) {
				subnet = s
				break
			}
		}
		publicInterface := Interface{
			Name:           fmt.Sprintf("%s%d", publicComponent.Name, publicComponent.Port),
			Auto:           true,
			AllowHotplug:   true,
			Address:        publicComponent.PrimaryIPAddress,
			Netmask:        subnet.Netmask,
			Gateway:        subnet.Gateway,
			DefaultGateway: nw.HasDefaultGateway(),
		}
		interfaces = append(interfaces, publicInterface)
	}

	return interfaces, nil
}

func (u *Ubuntu) manualInterfaces(networkComponents VirtualGuestNetworkComponents, networks Networks) ([]Interface, error) {
	privateComponent := networkComponents.PrimaryBackendNetworkComponent
	publicComponent := networkComponents.PrimaryNetworkComponent

	interfaces := []Interface{}
	for networkName, nw := range networks {
		if subnet, err := privateComponent.NetworkVLAN.Subnets.containing(nw.IP); err == nil {
			intf := Interface{
				Name:           fmt.Sprintf("%s%d:%s", privateComponent.Name, privateComponent.Port, networkName),
				Auto:           true,
				AllowHotplug:   true,
				Address:        nw.IP,
				Netmask:        subnet.Netmask,
				Gateway:        subnet.Gateway,
				DefaultGateway: nw.HasDefaultGateway(),
				Routes:         SoftlayerPrivateRoutes(subnet.Gateway),
			}

			interfaces = append(interfaces, intf)
			continue
		}

		if subnet, err := publicComponent.NetworkVLAN.Subnets.containing(nw.IP); err == nil {
			intf := Interface{
				Name:           fmt.Sprintf("%s%d:%s", publicComponent.Name, publicComponent.Port, networkName),
				Auto:           true,
				AllowHotplug:   true,
				Address:        nw.IP,
				Netmask:        subnet.Netmask,
				Gateway:        subnet.Gateway,
				DefaultGateway: nw.HasDefaultGateway(),
			}

			interfaces = append(interfaces, intf)
			continue
		}

		return nil, errors.New("manual subnet not found")
	}

	return interfaces, nil
}

func (i Interfaces) Configuration() ([]byte, error) {
	buf := &bytes.Buffer{}

	t := template.Must(template.New("network-interfaces").Parse(ETC_NETWORK_INTERFACES_TEMPLATE))
	err := t.Execute(buf, i)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

const ETC_NETWORK_INTERFACES_TEMPLATE = `# Generated by softlayer-cpi
auto lo
iface lo inet loopback
{{ range . -}}
# {{ .Name }}
{{- if .Auto }}
auto {{ .Name }}
{{- end }}
{{- if .AllowHotplug }}
allow-hotplug {{ .Name }}
{{- end }}
iface {{ .Name }} inet static
    address {{ .Address }}
    netmask {{ .Netmask }}
    {{- if .DefaultGateway }}
    gateway {{ .Gateway }}
		{{- end }}
    {{- range $route := .Routes }}
    post-up route add -net {{ $route.Network }} netmask {{ $route.Netmask }} gw {{ $route.Gateway }}
    {{- end }}
{{- if .DNS }}
    dns-nameservers{{ range .DNS }} {{ . }}{{ end }}
{{- end }}
{{ end }}`
