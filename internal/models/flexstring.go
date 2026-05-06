package models

import (
	"bytes"
	"encoding/json"
)

// FlexString acepta JSON que venga como string o como número y siempre
// lo serializa de vuelta como string. El Alteon a veces devuelve campos
// como "TotalBw" unas veces entre comillas y otras como número según el
// servicio — este tipo maneja ambos sin romper el parseo.
type FlexString string

func (f *FlexString) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*f = ""
		return nil
	}

	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*f = FlexString(s)
		return nil
	}

	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexString(n.String())
		return nil
	}

	*f = FlexString(data)
	return nil
}

func (f FlexString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(f))
}

func (f FlexString) String() string { return string(f) }
