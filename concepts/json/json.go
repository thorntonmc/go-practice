package json

import (
	"bytes"
	"encoding/json"
	"log"
)

type testJSON struct {
	Name string `json:"name"`
}

func marshal(b []byte, tj *testJSON) error {
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()

	err := d.Decode(tj)
	if err != nil {
		log.Println("got an err ", err)
		return err
	}

	log.Println("no error")

	return nil
}
