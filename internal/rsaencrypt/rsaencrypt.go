package rsaencrypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	// reading file
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		e := fmt.Errorf("cant read file: {%s}: %s", path, err.Error())
		return nil, e
	}
	// decoding pem
	data, _ := pem.Decode(bytes)
	// loading private key
	private, err := x509.ParsePKCS1PrivateKey(data.Bytes)

	return private, err

}

func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	// reading file
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		e := fmt.Errorf("cant read file: {%s}: %s", path, err.Error())
		return nil, e
	}
	// decoding pem
	data, _ := pem.Decode(bytes)
	// loading private key
	public, err := x509.ParsePKCS1PublicKey(data.Bytes)
	return public, err
}

func Encrypt(public *rsa.PublicKey, message []byte, label []byte) ([]byte, error) {
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, public, message, label)
	return ciphertext, err
}

func Decrypt(private *rsa.PrivateKey, message []byte, label []byte) ([]byte, error) {
	ciphertext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, private, message, label)
	return ciphertext, err
}

func Generate(outDir string) error {
	// generate key
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Cannot generate RSA keyn")
		os.Exit(1)
	}
	publickey := &privatekey.PublicKey

	// dump private key to file
	var privateKeyBytes []byte = x509.MarshalPKCS1PrivateKey(privatekey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	privatePemFile, err := os.Create(path.Join(outDir, "private.pem"))
	if err != nil {
		e := fmt.Errorf("error when create private.pem: %s n", err)
		return e
	}
	defer privatePemFile.Close()

	err = pem.Encode(privatePemFile, privateKeyBlock)
	if err != nil {
		e := fmt.Errorf("error when encode private pem: %s n", err)
		return e
	}

	// dump public key to file
	publicKeyBytes := x509.MarshalPKCS1PublicKey(publickey)

	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicPemFile, err := os.Create(path.Join(outDir, "public.pem"))
	if err != nil {
		e := fmt.Errorf("error when create public.pem: %s n", err)
		return e
	}
	defer publicPemFile.Close()

	err = pem.Encode(publicPemFile, publicKeyBlock)
	if err != nil {
		e := fmt.Errorf("error when encode public pem: %s n", err)
		return e
	}
	return nil

}
