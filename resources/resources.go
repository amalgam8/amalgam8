package resources

import (
	"strconv"
	"time"
)

// BasicEntry TODO
type BasicEntry struct {
	ID      string  `json:"_id"`
	Rev     string  `json:"_rev,omitempty"`
	IV      string  `json:"iv"`
	Version float64 `json:"version"`
}

// IDRev TODO
func (e *BasicEntry) IDRev() (string, string) {
	return e.ID, e.Rev
}

// SetRev
func (e *BasicEntry) SetRev() {
	if e.Rev == "" {
		e.Rev = "0"
	}
	i, _ := strconv.Atoi(e.Rev)
	i += 1
	e.Rev = strconv.Itoa(i)
}

// SetIV TODO
func (e *BasicEntry) SetIV(iv string) {
	e.IV = iv
}

// GetIV TODO
func (e *BasicEntry) GetIV() string {
	return e.IV
}

// MetaData service instance metadata
type MetaData struct {
	Version string
}

// ServiceCatalog TODO
type ServiceCatalog struct {
	BasicEntry
	Services   []Service
	LastUpdate time.Time
}

// Service TODO
type Service struct {
	Name      string
	Endpoints []Endpoint
}

// Endpoint TODO
type Endpoint struct {
	Type     string
	Value    string
	Metadata MetaData
}

// ByService TODO
type ByService []Service

// Len TODO
func (a ByService) Len() int {
	return len(a)
}

// Swap TODO
func (a ByService) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less TODO
func (a ByService) Less(i, j int) bool {
	return a[i].Name < a[j].Name
}

// ByEndpoint TODO
type ByEndpoint []Endpoint

// Len TODO
func (a ByEndpoint) Len() int {
	return len(a)
}

// Swap TODO
func (a ByEndpoint) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less TODO
func (a ByEndpoint) Less(i, j int) bool {
	if a[i].Value == a[j].Value {
		if a[i].Type == a[j].Type {
			return a[i].Metadata.Version < a[j].Metadata.Version
		}
		return a[i].Type < a[j].Type
	}
	return a[i].Value < a[j].Value
}

// Registry TODO
type Registry struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

// Kafka TODO
type Kafka struct {
	APIKey   string   `json:"api_key"`
	AdminURL string   `json:"admin_url"`
	RestURL  string   `json:"rest_url"`
	Brokers  []string `json:"brokers"`
	User     string   `json:"user"`
	Password string   `json:"password"`
	SASL     bool     `json:"sasl"`
}

// Credentials TODO
type Credentials struct {
	Kafka    Kafka    `json:"kafka"`
	Registry Registry `json:"registry"`
}

// ProxyConfig TODO
type ProxyConfig struct {
	BasicEntry
	LoadBalance       string      `json:"load_balance"`
	Port              int         `json:"port"`
	ReqTrackingHeader string      `json:"req_tracking_header"` // TODO: name?
	Filters           Filters     `json:"filters"`
	Credentials       Credentials `json:"credentials"`
}

// Filters TODO
type Filters struct {
	Rules    []Rule    `json:"rules"`
	Versions []Version `json:"versions"`
}

// Rule TODO
type Rule struct {
	Source           string  `json:"source"`
	Destination      string  `json:"destination"`
	Header           string  `json:"header"`
	Pattern          string  `json:"pattern"`
	Delay            float64 `json:"delay"`
	DelayProbability float64 `json:"delay_probability"`
	AbortProbability float64 `json:"abort_probability"`
	ReturnCode       int     `json:"return_code"`
}

// Version TODO
type Version struct {
	Service   string `json:"service"`
	Default   string `json:"default"`
	Selectors string `json:"selectors"`
}
