package client

const (
	DestinationWeightType             = "destinationWeight"
	DestinationWeightFieldDestination = "destination"
	DestinationWeightFieldWeight      = "weight"
)

type DestinationWeight struct {
	Destination *Destination `json:"destination,omitempty" yaml:"destination,omitempty"`
	Weight      int64        `json:"weight,omitempty" yaml:"weight,omitempty"`
}
