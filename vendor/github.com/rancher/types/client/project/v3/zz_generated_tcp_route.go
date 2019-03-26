package client

const (
	TCPRouteType       = "tcpRoute"
	TCPRouteFieldMatch = "match"
	TCPRouteFieldRoute = "route"
)

type TCPRoute struct {
	Match []L4MatchAttributes `json:"match,omitempty" yaml:"match,omitempty"`
	Route []DestinationWeight `json:"route,omitempty" yaml:"route,omitempty"`
}
