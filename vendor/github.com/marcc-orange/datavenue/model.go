package datavenue

type Stream struct {
	ID        string   `json:"id"`
	LastValue AnyValue `json:"lastValue"`
}

type Notification struct {
	DatasourceID string   `json:"datasourceID"`
	StreamID     string   `json:"streamID"`
	Values       []*Value `json:"values"`
}

type Value struct {
	Value    AnyValue               `json:"value"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type AnyValue interface{}
