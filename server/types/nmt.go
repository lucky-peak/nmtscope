package types

type NMTResponse[T any] struct {
	Data T `json:"data"`
}

type NMTHeader struct {
	Name      string `json:"name"`
	Reserved  int    `json:"reserved"`
	Committed int    `json:"committed"`
}

type NMTEntry struct {
	NMTHeader
}

type NMTReport struct {
	PID        int        `json:"pid"`
	Created    int64      `json:"created"`
	NMTEntries []NMTEntry `json:"nmt_entries"`
}
