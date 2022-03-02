package keystore

import (
	"bytes"
	"crypto/aes"
	"encoding/json"
	"fmt"

	web3 "github.com/umbracle/ethgo"
)

// EncryptV3 encrypts data in v3 format
func EncryptV3(content []byte, password string, customScrypt ...int) ([]byte, error) {

	// default scrypt values
	scryptN, scryptP := 1<<18, 1

	if len(customScrypt) >= 1 {
		scryptN = customScrypt[0]
	}
	if len(customScrypt) >= 2 {
		scryptP = customScrypt[1]
	}

	iv := getRand(aes.BlockSize)

	scrypt := scryptParams{
		N:     scryptN,
		R:     8,
		P:     scryptP,
		Dklen: 32,
		Salt:  hexString(getRand(32)),
	}
	kdf, err := scrypt.Key([]byte(password))
	if err != nil {
		return nil, err
	}

	cipherText, err := aesCTR(kdf[:16], content, iv)
	if err != nil {
		return nil, err
	}

	// generate mac
	mac := web3.Keccak256(kdf[16:32], cipherText)

	v3 := &v3Encoding{
		Version: 3,
		Crypto: &cryptoEncoding{
			Cipher:     "aes-128-ctr",
			CipherText: hexString(cipherText),
			CipherParams: struct{ IV hexString }{
				IV: hexString(iv),
			},
			KDF:       "scrypt",
			KDFParams: scrypt,
			Mac:       hexString(mac),
		},
	}

	encrypted, err := v3.Marshal()
	if err != nil {
		return nil, err
	}
	return encrypted, nil
}

// DecryptV3 decodes bytes in the v3 keystore format
func DecryptV3(content []byte, password string) ([]byte, error) {
	encoding := v3Encoding{}
	if err := encoding.Unmarshal(content); err != nil {
		return nil, err
	}
	if encoding.Version != 3 {
		return nil, fmt.Errorf("only version 3 supported")
	}
	if encoding.Crypto.Cipher != "aes-128-ctr" {
		return nil, fmt.Errorf("cipher %s not supported", encoding.Crypto.Cipher)
	}

	// decode the kdf
	kdf, err := applyKdf(encoding.Crypto.KDF, []byte(password), encoding.Crypto.KDFParamsRaw)
	if err != nil {
		return nil, err
	}

	// validate mac
	mac := web3.Keccak256(kdf[16:32], encoding.Crypto.CipherText)
	if !bytes.Equal(mac, encoding.Crypto.Mac) {
		return nil, fmt.Errorf("incorrect mac")
	}

	dst, err := aesCTR(kdf[:16], encoding.Crypto.CipherText, encoding.Crypto.CipherParams.IV)
	if err != nil {
		return nil, err
	}
	return dst, nil
}

type v3Encoding struct {
	ID      string          `json:"id"`
	Version int64           `json:"version"`
	Crypto  *cryptoEncoding `json:"crypto"`
}

func (j *v3Encoding) Marshal() ([]byte, error) {
	params, err := json.Marshal(j.Crypto.KDFParams)
	if err != nil {
		return nil, err
	}
	j.Crypto.KDFParamsRaw = json.RawMessage(params)
	return json.Marshal(j)
}

func (j *v3Encoding) Unmarshal(data []byte) error {
	return json.Unmarshal(data, j)
}

type cryptoEncoding struct {
	Cipher       string `json:"cipher"`
	CipherParams struct {
		IV hexString
	} `json:"cipherparams"`
	CipherText   hexString `json:"ciphertext"`
	KDF          string    `json:"kdf"`
	KDFParams    interface{}
	KDFParamsRaw json.RawMessage `json:"kdfparams"`
	Mac          hexString       `json:"mac"`
}
