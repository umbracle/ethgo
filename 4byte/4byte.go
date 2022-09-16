package fourbyte

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const (
	fourByteURL = "https://www.4byte.directory"
)

// Resolve resolves a method/event signature
func Resolve(str string) ([]string, error) {
	return get("/api/v1/signatures/?hex_signature=" + str + "&ordering=created_at")
}

func get(path string) ([]string, error) {
	req, err := http.Get(fourByteURL + path)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Results []struct {
			TextSignature string `json:"text_signature"`
		}
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if len(result.Results) == 0 {
		return nil, nil
	}
	signatures := make([]string, 0)
	for _, r := range result.Results {
		signatures = append(signatures, r.TextSignature)
	}
	return signatures, nil
}
