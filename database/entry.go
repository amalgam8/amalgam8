package database

// Entry generic Cloudant entry
// To use this interface with the Cloudant DB, the following fields need to
// be present in the struct:
//    ID string `json:"_id"`
//    Rev string `json:"_rev,omitempty"`
type Entry interface {
	IDRev() (string, string)
	GetIV() string
	SetIV(iv string)
	SetRev()
}

// AllDocs generic interface representing bulk read objects
type AllDocs interface {
	GetEntries() []Entry
}
