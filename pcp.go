package nat

import(
  "net"
  "time"

  "github.com/sashahilton00/go-pcp"
)

var (
	_ NAT = (*pcpClient)(nil)
)

type pcpClient struct {
  c *pcp.Client
}

func discoverPCP() <- chan NAT {
  res := make(chan NAT, 1)
  var client *pcp.Client
	client, err := pcp.NewClient()
	if err == nil {
    //Currently connection checking logic is missing upstream. Thus PCP is always returned as an option even where not present.
		res <- &pcpClient{client}
	}
  return res
}

func (p *pcpClient) GetDeviceAddress() (addr net.IP, err error) {
  return p.c.GetGatewayAddress()
}

func (p *pcpClient) GetExternalAddress() (addr net.IP, err error) {
  return p.c.GetExternalAddress()
}

func (p *pcpClient) GetInternalAddress() (addr net.IP, err error) {
  return p.c.GetInternalAddress()
}

func (p *pcpClient) AddPortMapping(protocol string, internalPort int, description string, timeout time.Duration) (mappedExternalPort int, err error) {
  proto := stringToProtocol(protocol)
  externalPort := randomPort()
  timeoutInSeconds := int(timeout / time.Second)
  lifetime := uint32(timeoutInSeconds)
  err = p.c.AddPortMapping(proto, uint16(internalPort), uint16(externalPort), nil, lifetime)
  if err != nil {
    return 0, err
  }
  return externalPort, nil
}

func (p *pcpClient) DeletePortMapping(protocol string, internalPort int) (err error) {
  err = p.c.DeletePortMapping(uint16(internalPort))
  return
}

func (p *pcpClient) Type() string {
  return "PCP"
}

func stringToProtocol(s string) pcp.Protocol {
  var p pcp.Protocol
  switch s {
  case "tcp":
    p = 6
  case "udp":
    p = 17
  default:
    p = 0
  }
  return p
}
