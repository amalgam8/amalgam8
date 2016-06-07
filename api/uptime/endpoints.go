package uptime

// URL returns URL path used for uptime checks
func URL() string {
	return uptimePath
}

// HealthyURL returns URL path used for healthy checks
func HealthyURL() string {
	return healthyPath
}

const (
	uptimePath  = "/uptime"
	healthyPath = "/"
)
