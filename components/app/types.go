package app

import "github.com/containernetworking/cni/pkg/types"

// Net cni configuration
type Net struct {
	Version string `json:"cniVersion"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Master  string `json:"master"`
	Mode    string `json:"mode"`
	IPAM    IPAM   `json:"ipam"`
}

// IPAM ip address mangagement
type IPAM struct {
	Type  string    `json:"type"`
	Range RangeSet  `json:"range"`
	DNS   types.DNS `json:"dns"`
}

// RangeSet ...
type RangeSet []Range

// Range ...
type Range struct {
	Dev    string `json:"dev"`
	Subnet string `json:"subnet"`
	Start  string `json:"start"`
	End    string `json:"end"`
	GW     string `json:"gw"`
}

// ValueSet ...
// type ValueSet []Value

// Value ...
// type Value struct {
// 	Start string `json:"start"`
// 	End   string `json:"end"`
// 	GW    string `json:"gw"`
// }
