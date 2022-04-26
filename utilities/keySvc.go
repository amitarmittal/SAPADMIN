package utils

import (
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
)

const (
	hypexSalt = "3309d355-0e71-4e1f-86ac-535056ce76c6" // salt for hashing signature from and to hypex operator
)

func GetNewKeys() (string, string, error) {
	// Generating RSA Private Key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Println("Error generating Key Pair : ", err.Error())
		return "", "", fmt.Errorf("Error generating Key Pair")
	}
	priKeyStr := ExportRsaPrivateKeyAsPemStr(privateKey)
	// Extract Public Key from RSA Private Key
	publicKey := privateKey.PublicKey
	pubKeyStr, err := ExportRsaPublicKeyAsPemStr(&publicKey)
	if err != nil {
		log.Println("Error exporting public key : ", err.Error())
		return "", "", fmt.Errorf("Error exporting public key")
	}
	return priKeyStr, pubKeyStr, nil
}

func GenerateKey() (rsa.PrivateKey, rsa.PublicKey) {
	// Generating RSA Private Key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		log.Println("Error generating Key Pair : ", err.Error())
		os.Exit(1)
	}

	// Extract Public Key from RSA Private Key
	publicKey := privateKey.PublicKey

	fmt.Println("Private Key (2048) :  ", *privateKey)
	fmt.Println("Public Key (2048) ", publicKey)

	return *privateKey, publicKey
}

func ExportRsaPrivateKeyAsPemStr(privkey *rsa.PrivateKey) string {
	privkey_bytes := x509.MarshalPKCS1PrivateKey(privkey)
	privkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privkey_bytes,
		},
	)
	return string(privkey_pem)
}

func ParseRsaPrivateKeyFromPemStr(privPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

func ExportRsaPublicKeyAsPemStr(pubkey *rsa.PublicKey) (string, error) {
	pubkey_bytes, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		return "", err
	}
	pubkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubkey_bytes,
		},
	)

	return string(pubkey_pem), nil
}

func ParseRsaPublicKeyFromPemStr(pubPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		break // fall through
	}
	return nil, errors.New("Key type is not RSA")
}

func CreateSignature(payload string, privateKey rsa.PrivateKey) (string, error) {
	rng := rand.Reader
	hashed := sha256.Sum256([]byte(payload))
	signature, err := rsa.SignPKCS1v15(rng, &privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		log.Println("CreateSignature: *** Signature creation failed with error - ***", err.Error())
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func VerifySignature(signature string, payload string, publicKey rsa.PublicKey) bool {
	sig, _ := base64.StdEncoding.DecodeString(signature)
	hashed := sha256.Sum256([]byte(payload))
	err := rsa.VerifyPKCS1v15(&publicKey, crypto.SHA256, hashed[:], sig)
	if err != nil {
		fmt.Printf("Error from verification: %s\n", err.Error())
		log.Println("VerifySignature: *** Signature verification Failed with error - ***", err.Error())
		return false
	}

	return true
}

func VerifySignature2(signature string, payload string, publicKey rsa.PublicKey) bool {
	sig, _ := base64.StdEncoding.DecodeString(signature)
	hashed := []byte(payload)
	err := rsa.VerifyPKCS1v15(&publicKey, crypto.SHA256, hashed[:], sig)
	if err != nil {
		fmt.Printf("Error from verification: %s\n", err.Error())
		log.Println("VerifySignature: *** Signature verification Failed with error - ***", err.Error())
		return false
	}

	return true
}

func CreateHashAsSignature(payload string) string {
	hashed := sha256.Sum256([]byte(payload))
	return fmt.Sprintf("%X", hashed)
}

func VerifyHashAsSignature(signature string, payload string) bool {
	hashed := sha256.Sum256([]byte(payload))
	return fmt.Sprintf("%X", hashed) == signature
}

func CreateHashAsSignatureWithSalt(payload string) (string, error) {
	mac := hmac.New(sha256.New, []byte(hypexSalt))
	_, err := mac.Write([]byte(payload))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%X", mac.Sum(nil)), nil
}

func VerifyHashAsSignatureWithSalt(sign, payload string) (bool, error) {
	mac := hmac.New(sha256.New, []byte(hypexSalt))
	_, err := mac.Write([]byte(payload))
	if err != nil {
		return false, err
	}
	return fmt.Sprintf("%X", mac.Sum(nil)) == sign, nil
}
