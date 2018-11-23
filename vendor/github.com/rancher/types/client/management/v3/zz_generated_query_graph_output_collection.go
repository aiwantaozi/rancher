package client

const (
	QueryGraphOutputCollectionType      = "queryGraphOutputCollection"
	QueryGraphOutputCollectionFieldData = "data"
	QueryGraphOutputCollectionFieldType = "type"
)

type QueryGraphOutputCollection struct {
	Data []QueryGraph `json:"data,omitempty" yaml:"data,omitempty"`
	Type string       `json:"type,omitempty" yaml:"type,omitempty"`
}
