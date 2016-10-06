package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"time"
)

// Icons can be found at http://fontawesome.io/icons/

func main() {
	serviceInstancesByContainerID = make(map[string]serviceInstance)
	go hostCommunication()
	go getAmalgam8ContainerIds()

	hostname, _ := os.Hostname()
	var (
		addr   = flag.String("addr", "/var/run/scope/plugins/a8routing.sock", "unix socket to listen for connections on")
		hostID = flag.String("hostname", hostname, "hostname of the host running this plugin")
	)
	flag.Parse()

	log.Printf("Starting on %s...\n", *hostID)

	os.Remove(*addr)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		os.Remove(*addr)
		os.Exit(0)
	}()

	listener, err := net.Listen("unix", *addr)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		listener.Close()
		os.Remove(*addr)
	}()

	log.Printf("Listening on: unix://%s", *addr)

	plugin := &Plugin{HostID: *hostID}
	http.HandleFunc("/report", plugin.Report)
	http.HandleFunc("/control", plugin.Control)
	if err := http.Serve(listener, nil); err != nil {
		log.Printf("error: %v", err)
	}
}

// Plugin groups the methods a plugin needs
type Plugin struct {
	HostID string

	lock       sync.Mutex
	routesEnabled bool
}

type request struct {
	NodeID  string
	Control string
}

type response struct {
	ShortcutReport *report `json:"shortcutReport,omitempty"`
}

type report struct {
	Container    topology
	Plugins []pluginSpec
}

type topology struct {
	Nodes           map[string]node           `json:"nodes"`
	MetricTemplates map[string]metricTemplate `json:"metric_templates"`
	Controls        map[string]control        `json:"controls"`
}

type node struct {
	Metrics        map[string]metric       `json:"metrics"`
	LatestControls map[string]controlEntry `json:"latestControls,omitempty"`
	AdjacencyList []string `json:"adjacency",omitempty`
	Rank string `json:rank,omitempty`
}

type metric struct {
	Samples []sample `json:"samples,omitempty"`
	Min     float64  `json:"min"`
	Max     float64  `json:"max"`
}

type sample struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

type controlEntry struct {
	Timestamp time.Time   `json:"timestamp"`
	Value     controlData `json:"value"`
}

type controlData struct {
	Dead bool `json:"dead"`
}

type metricTemplate struct {
	ID       string  `json:"id"`
	Label    string  `json:"label,omitempty"`
	Format   string  `json:"format,omitempty"`
	Priority float64 `json:"priority,omitempty"`
}

type control struct {
	ID    string `json:"id"`
	Human string `json:"human"`
	Icon  string `json:"icon"`
	Rank  int    `json:"rank"`
}

type pluginSpec struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Interfaces  []string `json:"interfaces"`
	APIVersion  string   `json:"api_version,omitempty"`
}

func (p *Plugin) makeReport() (*report, error) {
	// Add the container IDs to the map of nodes in the report
	m := make(map[string]node)
	for _, v := range serviceInstancesByContainerID {
		key := v.ContainerID + ";<container>"

		metrics, err := p.routingPercentage(v)
		if metrics == nil {
			//log.Println("skipping")
			continue
		}
		if err != nil {
			return nil, err
		}

		p.routesEnabled = true

		m[key] = node { 
			Metrics:        metrics,
			LatestControls: p.latestControls(),
			AdjacencyList: []string{""},
			Rank: "8",
		}
	}

	//[]string{"93bacc03b31df77768b47d9a5458946e22ddcc077468822fa0efa5da419ecb76;<container>"},

	rpt := &report{
		Container: topology{
			Nodes: m,
			MetricTemplates: p.metricTemplates(),
			Controls:        p.controls(),
		},
		Plugins: []pluginSpec{
			{
				ID:          "a8routing",
				Label:       "a8routing",
				Description: "Adds routing to different versions of a microservice",
				Interfaces:  []string{"reporter", "controller"},
				APIVersion:  "1",
			},
		},
	}
	return rpt, nil
}

// Get the routing weight for service+tags that match this service instance's tag
func (p *Plugin) routingPercentage(service serviceInstance) (map[string]metric, error) {
	id, _ := p.metricIDAndName()
	value := 0.0

	cmdArgs := []string{"-H 'Accept: application/json'","http://localhost:31200/v1/rules/routes/" + service.Name}
	nu, errrr := exec.Command("curl", cmdArgs...).Output()
	if errrr != nil {
		log.Fatal(errrr)
	}
	a := string(nu)
	var rules RulesList
	json.Unmarshal([]byte(a), &rules)

	/*
	if len(rules.Rules) > 0 {
		log.Println(rules)
	}*/	

	//log.Println(service.Name)
	//log.Println(service)

	for _, rule := range rules.Rules {
		if rule.Destination != service.Name {
			continue
		}
		if len(rule.Route.Backends) == 1 {
			//log.Println("only 1 backend")
			value = 100.0
		} else {
			for _, backend := range rule.Route.Backends {
				weight := backend.Weight
				for _, s := range backend.Tags {
					for _, tag := range service.Tags {
						if s == tag {
							//log.Println("tag:",s)
							value = weight*100
							//log.Println("tags are equal")
							//log.Println(weight)
							if weight == 0 {
								//log.Println("found tag, but weight is 0 or not present")
								return nil, nil
							}
							metrics := map[string]metric{
								id: {
									Samples: []sample{
										{
											Date:  time.Now(),
											Value: value,
										},
									},
									Min: 0,
									Max: 100,
								},
							}
							return metrics, nil
						}
					}
				}
			}
		}
	}
	//log.Println(service)
	//log.Println("no matched tags?")
	return nil, nil	
}

func (p *Plugin) latestControls() map[string]controlEntry {
	ts := time.Now()
	ctrls := map[string]controlEntry{}
	for _, details := range p.allControlDetails() {
		ctrls[details.id] = controlEntry{
			Timestamp: ts,
			Value: controlData{
				Dead: details.dead,
			},
		}
	}
	return ctrls
}

func (p *Plugin) metricTemplates() map[string]metricTemplate {
	id, name := p.metricIDAndName()
	return map[string]metricTemplate{
		id: {
			ID:       id,
			Label:    name,
			Format:   "percent",
			Priority: 0.1,
		},
	}
}

// Define the controls in the topology report @ /report
func (p *Plugin) controls() map[string]control {
	ctrls := map[string]control{}
	for _, details := range p.allControlDetails() {
		ctrls[details.id] = control{
			ID:    details.id,
			Human: details.human,
			Icon:  details.icon,
			Rank:  1,
		}
	}
	return ctrls
}

/**********
*
* Structs for handling the Amalgam8 service list API responses
*
***********/

// Value is typically the IP address of the service instance
type serviceEndpoint struct {
	Type string `json:type`
	Value string `json:value`
}

type serviceInstance struct {
	Id string `json:"id"`
	Name string `json:"service_name"`
	Endpoint serviceEndpoint `json:endpoint`
	Tags []string `json:tags`
	ContainerID string `json:"containerid,omitempty"`
}

type serviceDetails struct {
	Name string `json:"service_name"`
	Instances []serviceInstance `json:instances`
}

type serviceListResponse struct {
	Services []string `json:services`
}

// Report is called by scope when a new report is needed. It is part of the
// "reporter" interface, which all plugins must implement.
func (p *Plugin) Report(w http.ResponseWriter, r *http.Request) {
	p.lock.Lock()
	defer p.lock.Unlock()
	log.Println(r.URL.String())
	rpt, err := p.makeReport()
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	raw, err := json.Marshal(*rpt)
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//log.Println(string(raw))
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

type idAddressPair struct {
	ID string `json:id`
	IP string `json:ip`
}

var latestHostServerResponse string

func hostCommunication() {
	log.Println("hostCommunication")
	for {
		time.Sleep(2 * time.Second)
		c, err := net.Dial("unix", "/var/run/dockerConnection/hostconnection.sock")
		if err != nil {
			continue;
		}
		// send to socket
		log.Println("sending to server")
		fmt.Fprintf(c, "hi" + "\n")
		// listen for reply
		message, _ := bufio.NewReader(c).ReadString('\n')
		//log.Println("Message from server: " + message)
		log.Println("Received update from host server")

		// set  this to be the latest response
		latestHostServerResponse = message
	}
} 

// map of service instances, with the IP addresses as the keys
//var serviceInstances []serviceInstance
// map of service instances, with the container ID as the key
var serviceInstancesByContainerID map[string]serviceInstance

func updateAmalgam8ServiceInstances() map[string]serviceInstance{
	// amalgam8
	cmdArgs := []string{"-H 'Accept: application/json'","http://localhost:31300/api/v1/services"}
	o, errrr := exec.Command("curl", cmdArgs...).Output()
	if errrr != nil {
		log.Fatal(errrr)
	}
	s := string(o)
	m := make(map[string]serviceInstance) // IP addresses are the keys to the map of instances
	var svcResponse serviceListResponse
	json.Unmarshal([]byte(s), &svcResponse)
	
	// Get the IP address of each service
	for _, serviceName := range svcResponse.Services {
		var svcDetails serviceDetails
		//log.Println(serviceName)
		cmdArgs = []string{"-H 'Accept: application/json'","http://localhost:31300/api/v1/services/" + serviceName}
		nu, errrr := exec.Command("curl", cmdArgs...).Output()
		if errrr != nil {
			log.Fatal(errrr)
		}
		a := string(nu)
		json.Unmarshal([]byte(a), &svcDetails)

		// Add each instance of the service to the list
		for _, instance := range svcDetails.Instances {
			ip := strings.Split(instance.Endpoint.Value, ":")
			m[ip[0]] = instance
		}
	}
	log.Println("Updated Amalgam8 Service Instances without Container IDs")
	return m
}

func getAllContainerIdAddressPairs(serverJsonString string) []idAddressPair {
	var pairs []idAddressPair
	if len(serverJsonString) == 0 {
		return pairs
	}
	json.Unmarshal([]byte(serverJsonString), &pairs)
	return pairs
}

// Look at the Amalgam8 IP addresses and use this list to filter the list of ID/IP pairs from hostCommunication
func getAmalgam8ContainerIds() {
	for {
		log.Println("getAmalgam8ContainerIds")
		time.Sleep(3 * time.Second)

		// Get the IP addresses of Amalgam8 containers
		//addressMap := findAmalgam8Addresses()
		m := make(map[string]serviceInstance) // containerIDs are the keys to this map of instances

		containerIDAddressPairs := getAllContainerIdAddressPairs(latestHostServerResponse)

		// map of service instances by IP address
		serviceInstances := updateAmalgam8ServiceInstances()

		time.Sleep(1 * time.Second)

		// Add the pairs with Amalgam8 IP addresses to our collection
		for _, pa := range containerIDAddressPairs {
			// Check to see if this container is in the collection of Amalgam8 services
			if _, ok := serviceInstances[pa.IP]; ok { 
				var tmp = serviceInstances[pa.IP]
				tmp.ContainerID = pa.ID
				serviceInstances[pa.IP] = tmp
				m[pa.ID] = tmp
				//log.Println(serviceInstances[pa.IP])
			}
		}
		serviceInstancesByContainerID = m
		log.Println("Added Container IDs to Amalgam8 services")
	}
}

/********************************
*
* For handling /v1/rules/routes/{service_name}
*
********************************/

type Backend struct {
	Weight float64 `json:"weight"`
	Tags []string `json:"tags"`
}

type Route struct {
	Backends []Backend `json:"backends"`
}

type Rule struct {
	Id string `json:"id"`
	Priority int `json:"priority"`
	Destination string `json:"destination"`
	Route Route `json:"route"`
}

type RulesList struct {
	Rules []Rule `json:"rules"`
}

// Control is called by scope when a control is activated. It is part
// of the "controller" interface.
func (p *Plugin) Control(w http.ResponseWriter, r *http.Request) {
	p.lock.Lock()
	defer p.lock.Unlock()
	log.Println(r.URL.String())
	xreq := request{}
	err := json.NewDecoder(r.Body).Decode(&xreq)
	if err != nil {
		log.Printf("Bad request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Print out the node & control values
	log.Println("node ID: " + xreq.NodeID)
	log.Println("control ID: " + xreq.Control)

	switch {
		case xreq.Control == "clearroutes": clearRoutes(xreq.NodeID)
	}

	rpt, err := p.makeReport()
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := response{ShortcutReport: rpt}
	raw, err := json.Marshal(res)
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

func (p *Plugin) metricIDAndName() (string, string) {
	return "routeamount", "Route Setting"
}

type controlDetails struct {
	id    string
	human string
	icon  string
	dead  bool
}

func clearRoutes(NodeId string) {
	idParts := strings.Split(NodeId, ";")
	name := serviceInstancesByContainerID[idParts[0]].Name
	log.Println("Clearing routes for " + name)
	req, _ := http.NewRequest("DELETE", "http://localhost:31200/v1/rules/routes/" + name, nil)
	_, _ = http.DefaultClient.Do(req)
}

func (p *Plugin) allControlDetails() []controlDetails {
	return []controlDetails{
		{
			id:    "decrease",
			human: "Decrease Routing by 10%",
			icon:  "fa-angle-down",
			dead:  !p.routesEnabled,
		},
		{
			id:    "increase",
			human: "Increase Routing by 10%",
			icon:  "fa-angle-up",
			dead:  !p.routesEnabled,
		},
		{
			id:    "clearroutes",
			human: "Delete route rules for this service",
			icon:  "fa-hand-paper-o ",
			dead:  !p.routesEnabled,
		},
		{
			id:    "enableroutes",
			human: "Enable routes for this service",
			icon:  "fa-rocket",
			dead:  p.routesEnabled,
		},
	}
}

func (p *Plugin) controlDetails() (string, string, string) {
	for _, details := range p.allControlDetails() {
		if !details.dead {
			return details.id, details.human, details.icon
		}
	}
	return "", "", ""
}