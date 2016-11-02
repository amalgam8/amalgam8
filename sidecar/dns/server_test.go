package dns

import (
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"strconv"
	"testing"
	"time"

	"sort"

	"github.com/amalgam8/amalgam8/registry/client"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	server   *Server
	config   Config
	myClient *mySimpleServiceDiscovery
}

/********************* Mock client ***************************/

type mySimpleServiceDiscovery struct {
	services []*client.ServiceInstance
}

// ListServices queries the registry for the list of services for which instances are currently registered.
func (m *mySimpleServiceDiscovery) ListServices() ([]string, error) {
	servicesNames := []string{}
	for _, service := range m.services {
		servicesNames = append(servicesNames, service.ServiceName)
	}
	return servicesNames, nil
}

// ListInstances queries the registry for the list of service instances currently registered.
// The given InstanceFilter can be used to filter the returned instances as well as the fields returned for each.
func (m *mySimpleServiceDiscovery) ListInstances(filter client.InstanceFilter) ([]*client.ServiceInstance, error) {

	servicesToReturn := []*client.ServiceInstance{}
	count := 0
	for _, service := range m.services {
		if service.ServiceName == filter.ServiceName {
			for _, tag := range filter.Tags {
				for _, serviceTag := range service.Tags {
					if tag == serviceTag {
						count++
					}
				}
			}
			if count == len(filter.Tags) {
				servicesToReturn = append(servicesToReturn, service)

			}
		}
	}
	return servicesToReturn, nil
}

// ListServiceInstances queries the registry for the list of service instances with status 'UP' currently
// registered for the given service.
func (m *mySimpleServiceDiscovery) ListServiceInstances(serviceName string) ([]*client.ServiceInstance, error) {

	servicesToReturn := []*client.ServiceInstance{}
	for _, service := range m.services {
		if service.ServiceName == serviceName {
			servicesToReturn = append(servicesToReturn, service)

		}

	}
	return servicesToReturn, nil
}

// create the dns server with a client , initialize registry with service instances.
func (suite *TestSuite) SetupTest() {
	var err error
	suite.myClient = new(mySimpleServiceDiscovery)

	rand.Seed(int64(time.Now().Nanosecond()))
	port := rand.Intn(9000-8000) + 8000

	suite.config = Config{
		DiscoveryClient: suite.myClient,
		Port:            uint16(port),
		Domain:          "amalgam8",
	}

	suite.server, err = NewServer(suite.config)
	suite.NoError(err)

	url1 := url.URL{
		Scheme: "http",
		Host:   "amalgam8",
		Path:   "/shopping/cart",
	}
	url2 := url.URL{
		Scheme: "http",
		Host:   "amalgam8",
		Path:   "/Orders",
	}
	suite.myClient.services = append(suite.myClient.services, &client.ServiceInstance{ServiceName: "shoppingCart",
		ID: "1", Endpoint: client.NewHTTPEndpoint(url1)})

	suite.myClient.services = append(suite.myClient.services, &client.ServiceInstance{ServiceName: "shoppingCart",
		ID: "2", Endpoint: client.NewTCPEndpoint("127.0.0.5", 5050)})

	suite.myClient.services = append(suite.myClient.services, &client.ServiceInstance{Tags: []string{"first", "second"},
		ServiceName: "shoppingCart", ID: "3", Endpoint: client.NewTCPEndpoint("127.0.0.4", 3050)})

	suite.myClient.services = append(suite.myClient.services, &client.ServiceInstance{ServiceName: "Orders",
		ID: "4", Endpoint: client.NewTCPEndpoint("127.0.0.10", 3050)})

	suite.myClient.services = append(suite.myClient.services, &client.ServiceInstance{ServiceName: "Orders",
		ID: "6", Endpoint: client.NewHTTPEndpoint(url2)})

	suite.myClient.services = append(suite.myClient.services, &client.ServiceInstance{ServiceName: "Orders",
		ID: "7", Endpoint: client.NewTCPEndpoint("132.68.5.6", 1010)})

	suite.myClient.services = append(suite.myClient.services, &client.ServiceInstance{ServiceName: "Reviews",
		ID: "8", Endpoint: client.NewTCPEndpoint("132.68.5.6", 1010)})

	suite.myClient.services = append(suite.myClient.services, &client.ServiceInstance{ServiceName: "httpService",
		ID: "9", Endpoint: client.NewHTTPEndpoint(url1)})

	go suite.server.ListenAndServe()
	time.Sleep((200) * time.Millisecond)

}

func (suite *TestSuite) TearDownTest() {
	suite.server.Shutdown()
}

func (suite *TestSuite) TestShoppingCartNoTags() {
	r, err := suite.doDNSQuery("shoppingCart.amalgam8.", dns.TypeA)

	suite.NoError(err)
	suite.Len(r.Answer, 2, "Should be two records for shoppingCart")
	suite.Equal(dns.RcodeSuccess, r.Rcode)

	sort.Sort(ByIP(r.Answer))

	suite.IsType(&dns.A{}, r.Answer[0])
	suite.IsType(&dns.A{}, r.Answer[1])

	suite.Equal(net.ParseIP("127.0.0.5").To4(), r.Answer[0].(*dns.A).A.To4())
	suite.Equal(net.ParseIP("127.0.0.4").To4(), r.Answer[1].(*dns.A).A.To4())
}

func (suite *TestSuite) TestUnregisteredServices() {
	r, err := suite.doDNSQuery("unregisterd.amalgam8.", dns.TypeA)

	suite.Equal(dns.RcodeNameError, r.Rcode)
	suite.NoError(err)
	suite.Empty(r.Answer, "No records for service unregistred")

	r, err = suite.doDNSQuery("httpService.service.amalgam8.", dns.TypeA)

	suite.Equal(dns.RcodeNameError, r.Rcode)
	suite.NoError(err)
	suite.Empty(r.Answer, "No records for service unregistred")
}

func (suite *TestSuite) TestEmptyRequest() {
	r, err := suite.doDNSQuery("amalgam8.", dns.TypeA)

	suite.Equal(dns.RcodeNameError, r.Rcode)
	suite.NoError(err)
	suite.Empty(r.Answer, "No records for serive unregistred")
}

func (suite *TestSuite) TestRequestsWithTags() {
	r, err := suite.doDNSQuery("first.second.shoppingCart.amalgam8.", dns.TypeA)

	suite.NoError(err)
	suite.Equal(dns.RcodeSuccess, r.Rcode)
	suite.Len(r.Answer, 1, "Should be 1 record for shoppingCart")

	suite.IsType(&dns.A{}, r.Answer[0])
	suite.Equal(net.IPv4(127, 0, 0, 4).To4(), r.Answer[0].(*dns.A).A.To4())

	r, err = suite.doDNSQuery("tag.Reviews.amalgam8.", dns.TypeA)

	suite.Equal(dns.RcodeNameError, r.Rcode)
	suite.NoError(err)
	suite.Empty(r.Answer, "No records for service unregistred")
}

func (suite *TestSuite) TestRequestsWithSubTags() {
	r, err := suite.doDNSQuery("seconds.shoppingCart.amalgam8.", dns.TypeA)

	suite.Equal(dns.RcodeNameError, r.Rcode, "Expected Error code : "+strconv.Itoa(dns.RcodeNameError)+" Got :"+strconv.Itoa(r.Rcode))
	suite.NoError(err)
	suite.Empty(r.Answer, "Should be No records for serive unregistred")
}

func (suite *TestSuite) TestRequestsSRVNoTags() {
	r, err := suite.doDNSQuery("_shoppingCart._tcp.amalgam8.", dns.TypeSRV)

	suite.NoError(err)
	suite.Len(r.Answer, 2, "Should be 2 tcp records for shoppingCart")
	suite.Equal(dns.RcodeSuccess, r.Rcode)

	sort.Sort(ByPort(r.Answer))
	sort.Sort(ByIP(r.Extra))

	suite.IsType(&dns.SRV{}, r.Answer[0])
	suite.IsType(&dns.SRV{}, r.Answer[1])

	target1 := fmt.Sprintf("%s.shoppingCart.amalgam8.", suite.myClient.services[1].ID)
	target2 := fmt.Sprintf("%s.shoppingCart.amalgam8.", suite.myClient.services[2].ID)

	suite.Equal(target1, r.Answer[0].(*dns.SRV).Target, "Wrong target for SRV record")
	suite.Equal(target2, r.Answer[1].(*dns.SRV).Target, "Wrong target for SRV record")
	suite.EqualValues(5050, r.Answer[0].(*dns.SRV).Port, "Wrong port for SRV record")
	suite.EqualValues(3050, r.Answer[1].(*dns.SRV).Port, "Wrong port for SRV record")

	suite.IsType(&dns.A{}, r.Extra[0])
	suite.IsType(&dns.A{}, r.Extra[1])
	suite.Equal(net.IPv4(127, 0, 0, 5).To4(), r.Extra[0].(*dns.A).A.To4())
	suite.Equal(net.IPv4(127, 0, 0, 4).To4(), r.Extra[1].(*dns.A).A.To4())
}

func (suite *TestSuite) TestRequestsSRVWithTag() {
	r, err := suite.doDNSQuery("_shoppingCart._first.amalgam8.", dns.TypeSRV)

	suite.NoError(err)
	suite.Len(r.Answer, 1, "Should be 1 records for with tag first")
	suite.Equal(dns.RcodeSuccess, r.Rcode)

	target := fmt.Sprintf("%s.shoppingCart.amalgam8.", suite.myClient.services[2].ID)

	suite.IsType(&dns.SRV{}, r.Answer[0])
	suite.Equal(target, r.Answer[0].(*dns.SRV).Target, "Wrong target name in SRV record")
	suite.EqualValues(3050, r.Answer[0].(*dns.SRV).Port, "Wrong port in SRV record")

	suite.IsType(&dns.A{}, r.Extra[0])
	suite.Equal(net.ParseIP("127.0.0.4").To4(), r.Extra[0].(*dns.A).A.To4())
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestTestSuite(t *testing.T) {

	suite.Run(t, new(TestSuite))
}

type ByPort []dns.RR

func (rrs ByPort) Len() int {
	return len(rrs)
}

func (rrs ByPort) Less(i, j int) bool {
	srvI, okI := rrs[i].(*dns.SRV)
	srvJ, okJ := rrs[j].(*dns.SRV)

	var portI, portJ uint16
	if okI {
		portI = srvI.Port
	}
	if okJ {
		portJ = srvJ.Port
	}
	return portI > portJ
}

func (rrs ByPort) Swap(i, j int) {
	rrs[i], rrs[j] = rrs[j], rrs[i]
}

type ByIP []dns.RR

func (rrs ByIP) Len() int {
	return len(rrs)
}

func (rrs ByIP) Less(i, j int) bool {
	aI, okI := rrs[i].(*dns.A)
	aJ, okJ := rrs[j].(*dns.A)

	var ipI, ipJ net.IP
	if okI {
		ipI = aI.A
	}
	if okJ {
		ipJ = aJ.A
	}
	return ipI.String() > ipJ.String()
}

func (rrs ByIP) Swap(i, j int) {
	rrs[i], rrs[j] = rrs[j], rrs[i]
}

func (suite *TestSuite) doDNSQuery(question string, questionType uint16) (*dns.Msg, error) {
	s := "127.0.0.1:" + strconv.Itoa(int(suite.config.Port))

	m := &dns.Msg{}
	m.SetQuestion(question, questionType)

	return dns.Exchange(m, s)
}
