package protocol

//Type represents the API protocol type
type Type uint32

//Defines the API protocol types
const (
	Amalgam8 Type = 1 << iota // Amalgam8 protocol
	Eureka                    // Eureka protocol
)

// NameOf returns the name of the given protocol type value
func NameOf(protocol Type) string {
	switch protocol {
	case Amalgam8:
		return "Amalgam8"
	case Eureka:
		return "Eureka"
	default:
		return "Unknown"
	}
}
