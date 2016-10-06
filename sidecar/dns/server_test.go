package dns

import (
	"testing"
	"github.com/stretchr/testify/suite"
	"github.com/amalgam8/amalgam8/registry/client"
	"net/url"

	//"github.com/stretchr/testify/assert"
	//"sync"
	"github.com/miekg/dns"
	"math/rand"
	"time"

	"net"

	"sync"
	"strconv"
	"fmt"
)


type TestSuite struct {
	suite.Suite
	server *Server
	myClient client.Client
	wg sync.WaitGroup

}

/******************** Helper functions ***********************/

/*************************************************************/

func (suite *TestSuite ) createNewClient() (client.Client, error) {
	conf := client.Config{URL: "http://172.17.0.02:8080"}
	return client.New(conf)
}

func (suite *TestSuite) createDNSServer() (*Server, error){
	suite.myClient, _ = suite.createNewClient()
	rand.Seed(int64(time.Now().Nanosecond()))
	port := rand.Intn(9000 - 8000) + 8000
	fmt.Println("Port generateed: ", port)
	dnsConfig := Config{
		DiscoveryClient: suite.myClient ,
		Port:            uint16(port),
		Domain:          "amalgam8",
	}
	return NewServer(dnsConfig)

}

/***************************************************************/

// create the dns server with a client , initialize registry with service instances.
func (suite *TestSuite) SetupTest() {
	var err error
	suite.server, err = suite.createDNSServer()
	suite.Nil(err, "Error should be nil")

	url1 := url.URL{
		Scheme:  "http",
		Host:    "amalgam8",
		Path:    "/shopping/cart",
	}
	url2 := url.URL{
		Scheme:  "http",
		Host:    "amalgam8",
		Path:    "/Orders",
	}
	_, err = ((suite.myClient)).Register(&client.ServiceInstance{ServiceName:"shoppingCart", ID: "1", Endpoint: client.NewHTTPEndpoint(url1)})
	suite.Nil(err, "Error should be nil")
	_, err = ((suite.myClient)).Register(&client.ServiceInstance{ServiceName:"shoppingCart", ID: "2", Endpoint: client.NewTCPEndpoint("127.0.0.5", 5050)})
	suite.Nil(err, "Error should be nil")
	_, err = ((suite.myClient)).Register(&client.ServiceInstance{Tags: []string{"first","second"}, ServiceName:"shoppingCart", ID: "3", Endpoint: client.NewTCPEndpoint("127.0.0.4", 5050)})
	suite.Nil(err, "Error should be nil")
	_, err = ((suite.myClient)).Register(&client.ServiceInstance{ServiceName:"Orders", ID: "4", Endpoint: client.NewTCPEndpoint("127.0.0.10", 3050)})
	suite.Nil(err, "Error should be nil")
	_, err = ((suite.myClient)).Register(&client.ServiceInstance{ServiceName:"Orders", ID: "6", Endpoint: client.NewHTTPEndpoint(url2)})
	suite.Nil(err, "Error should be nil")
	_, err = ((suite.myClient)).Register(&client.ServiceInstance{ServiceName:"Orders", ID: "7", Endpoint: client.NewTCPEndpoint("132.68.5.6", 1010)})
	suite.Nil(err, "Error should be nil")
	_, err = ((suite.myClient)).Register(&client.ServiceInstance{ServiceName:"Reviews", ID: "8", Endpoint: client.NewTCPEndpoint("132.68.5.6", 1010)})
	suite.Nil(err, "Error should be nil")
	_, err = ((suite.myClient)).Register(&client.ServiceInstance{ServiceName:"httpService", ID: "9", Endpoint: client.NewHTTPEndpoint(url1)})
	suite.Nil(err, "Error should be nil")
	suite.wg.Add(1)
	go suite.server.ListenAndServe()
	time.Sleep((10)*time.Second)
	//suite.Nil(err, "Error should be nil")

}

func (suite *TestSuite) TearDownTest() {
	suite.server.Shutdown()

}

// All methods that begin with "Test" are run as tests within a
// suite.

func (suite *TestSuite) TestShoppingCartNoTags() {
	target := "shoppingCart.amalgam8"
	server := "127.0.0.1:"

	c := dns.Client{}
	m := dns.Msg{}
	m.SetQuestion(target+".", dns.TypeA)
	r, _, err := c.Exchange(&m, server+strconv.Itoa(int(suite.server.config.Port)))
	suite.Nil(err)
	suite.Equal(2, len(r.Answer), "Should be two records for shoppingCart")

	suite.Equal(dns.RcodeSuccess,r.Rcode)

	for _, ans := range r.Answer {
		Arecord := ans.(*dns.A)

		a:=net.IPv4(127,0,0,4)
		b:=net.IPv4(127,0,0,5)
		suite.True(Arecord.A.Equal(a) || Arecord.A.Equal(b))


	}

}

func (suite *TestSuite) TestUnregisteredServices() {

	target := "unregisterd.amalgam8"
	server := "127.0.0.1:"

	c := dns.Client{}
	m := dns.Msg{}
	m.SetQuestion(target+".", dns.TypeA)
	r, _, err := c.Exchange(&m, server+strconv.Itoa(int(suite.server.config.Port)))
	suite.Equal(dns.RcodeNameError,r.Rcode)
	suite.Nil(err)
	suite.Equal(0, len(r.Answer), "No records for serive unregistred")

	target = "httpService.amalgam8"
	m.SetQuestion(target+".", dns.TypeA)
	r, _, err = c.Exchange(&m, server+strconv.Itoa(int(suite.server.config.Port)))
	suite.Equal(dns.RcodeNameError,r.Rcode)
	suite.Nil(err)
	suite.Equal(0, len(r.Answer), "No records for serive unregistred")
}


func (suite *TestSuite) TestEmptyRequest()  {
	target := "amalgam8"
	server := "127.0.0.1:"
	c := dns.Client{}
	m := dns.Msg{}
	m.SetQuestion(target+".", dns.TypeA)
	r, _, err := c.Exchange(&m, server+strconv.Itoa(int(suite.server.config.Port)))
	suite.Equal(dns.RcodeNameError,r.Rcode)
	suite.Nil(err)
	suite.Equal(0, len(r.Answer), "No records for serive unregistred")
}

func (suite *TestSuite) TestRequestsWithTags() {
	target := "first.second.shoppingCart.amalgam8"
	server := "127.0.0.1:"

	c := dns.Client{}
	m := dns.Msg{}
	m.SetQuestion(target+".", dns.TypeA)
	r, _, err := c.Exchange(&m, server+strconv.Itoa(int(suite.server.config.Port)))
	suite.Nil(err)
	suite.Equal(1, len(r.Answer), "Should be 1 record for shoppingCart")

	suite.Equal(dns.RcodeSuccess,r.Rcode)

	for _, ans := range r.Answer {
		Arecord := ans.(*dns.A)

		a:=net.IPv4(127,0,0,4)
		suite.True(Arecord.A.Equal(a))


	}


	target = "tag.Reviews.amalgam8"

	m.SetQuestion(target+".", dns.TypeA)
	r, _, err = c.Exchange(&m, server+strconv.Itoa(int(suite.server.config.Port)))
	suite.Equal(dns.RcodeNameError,r.Rcode)
	suite.Nil(err)
	suite.Equal(0, len(r.Answer), "No records for serive unregistred")

}

func (suite *TestSuite) TestRequestsWithSubTags() {
	target := "seconds.shoppingCart.amalgam8"
	server := "127.0.0.1:"

	c := dns.Client{}
	m := dns.Msg{}
	m.SetQuestion(target+".", dns.TypeA)
	r, _, err := c.Exchange(&m, server+strconv.Itoa(int(suite.server.config.Port)))
	suite.Equal(dns.RcodeNameError,r.Rcode, "Expected Error code : " + strconv.Itoa(dns.RcodeNameError) + " Got :" + strconv.Itoa(r.Rcode))
	suite.Nil(err)
	suite.Equal(0, len(r.Answer), "Should be No records for serive unregistred")
	//fmt.Println(r.Answer)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestTestSuite(t *testing.T) {

	suite.Run(t, new(TestSuite))
}