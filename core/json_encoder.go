package core

import "encoding/json"

type JsonEncoder struct {
}

// decodes the byte array into Application specific Message format
func (je *JsonEncoder) Decode(data []byte) (*Message, error) {

	var res Message

	if err := json.Unmarshal(data, &res); err != nil {
		AppLogger.logger.Errorln("error occurred while unmarshalling bytes to Message: ", err)
		return nil, err
	}
	return &res, nil
}

// encodes the application specific Message format into byte array
func (je *JsonEncoder) Encode(data Message) []byte {

	bytes, err := json.Marshal(data)
	if err != nil {
		AppLogger.logger.Errorln("error occurred while marshalling Message to bytes: ", err)
		return nil
	}
	return bytes
}
