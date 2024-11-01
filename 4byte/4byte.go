package fourbyte

import (
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
)

const (
	fourByteURL = "https://www.4byte.directory"
)

// Resolve resolves a method/event signature
func Resolve(str string) (string, error) {
	return get("/api/v1/signatures/?hex_signature=" + str)
}

// ResolveBytes resolves a method/event signature in bytes
func ResolveBytes(b []byte) (string, error) {
	return Resolve(hex.EncodeToString(b))
}

func get(path string) (string, error) {
	req, err := http.Get(fourByteURL + path)
	if err != nil {
		return "", err
	}
	defer req.Body.Close()

	data, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		Results []signatureResult
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}

	if len(result.Results) == 0 {
		return "", nil
	}
	return result.Results[0].TextSignature, nil
}

type signatureResult struct {
	TextSignature string `json:"text_signature"`
}
