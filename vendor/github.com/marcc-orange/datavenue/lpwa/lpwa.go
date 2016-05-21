package lpwa

import (
	"encoding/hex"

	"github.com/marcc-orange/datavenue"
)

// SendDownlinkData sends data with current fCnt
func SendDownlinkData(c *datavenue.Client, datasourceID, downlinkStreamID string, value []byte, fCnt, port uint32, confirmed bool) error {

	return c.AppendValues(datasourceID, downlinkStreamID, []*datavenue.Value{
		&datavenue.Value{
			Value: hex.EncodeToString(value),
			Metadata: map[string]interface{}{
				"fcnt":      fCnt,
				"port":      port,
				"confirmed": confirmed,
			},
		}})
}

// RetrieveDownlinkFCnt returns the current frame counter for downlink
func RetrieveDownlinkFCnt(c *datavenue.Client, datasourceID, downlinkFCntStreamID string) (uint32, error) {

	stream, err := c.RetreiveStream(datasourceID, downlinkFCntStreamID)
	if err != nil {
		return 0, err
	}

	return uint32(stream.LastValue.(float64)) + 1, nil
}

func RetrieveCommandsStates(c *datavenue.Client, datasourceID, commandStreamID string) (map[uint16]string, error) {

	values, err := c.RetreiveValues(datasourceID, commandStreamID)
	if err != nil {
		return nil, err
	}

	commands := map[uint16]string{}
	for _, value := range values {
		fcnt := uint16(value.Metadata["fcnt"].(float64))
		if status, ok := value.Metadata["status"]; ok {
			status := status.(string)
			commands[fcnt] = status
		}
	}

	return commands, nil
}
