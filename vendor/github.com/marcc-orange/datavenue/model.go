package datavenue

type Stream struct {
	ID        string   `json:"id"`
	LastValue AnyValue `json:"lastValue"`
}

type Values struct {
	Values []*Value `json:"values"`
}

type Value struct {
	Value    AnyValue               `json:"value"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type AnyValue interface{}
