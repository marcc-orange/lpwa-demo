package datavenue

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type Client struct {
	URL     string
	OAPIKey string
	ISSKey  string
	Client  *http.Client
}

func (c *Client) RetreiveStream(datasourceID, streamID string) (*Stream, error) {
	URL := c.URL + "/datasources/" + datasourceID + "/streams/" + streamID

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-ISS-Key", c.ISSKey)
	req.Header.Add("X-OAPI-Key", c.OAPIKey)

	response, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, errors.New("bad response: " + response.Status)
	}

	stream := &Stream{}
	err = json.NewDecoder(response.Body).Decode(stream)
	if err != nil {
		return nil, err
	}

	return stream, nil
}

func (c *Client) AppendValues(datasourceID, streamID string, values []*Value) error {
	URL := c.URL + "/datasources/" + datasourceID + "/streams/" + streamID + "/values"

	json, err := json.Marshal(values)
	if err != nil {
		return err
	}

	log.Println("POST", URL)
	log.Println("val:", string(json))

	req, err := http.NewRequest("POST", URL, bytes.NewReader(json))
	if err != nil {
		return err
	}
	req.Header.Add("X-ISS-Key", c.ISSKey)
	req.Header.Add("X-OAPI-Key", c.OAPIKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		var b bytes.Buffer
		b.ReadFrom(resp.Body)
		return errors.New("bad response: " + resp.Status + ":" + b.String())
	}

	return nil
}
