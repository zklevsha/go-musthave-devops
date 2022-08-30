package rsaencrypt

import (
	"log"
	"os"
	"testing"
)

func setUp(dirName string) {
	// create dir for keys
	err := os.Mkdir(dirName, os.ModePerm)
	if err != nil {
		log.Fatalf("Cant create dir %s: %s", dirName, err.Error())
	}
}

func tearDown(dirName string) {
	err := os.RemoveAll(dirName)
	if err != nil {
		log.Fatalf("Cant delete dir %s: %s", dirName, err.Error())
	}
}

func TestRSA(t *testing.T) {
	name := "testRSA"
	testString := "this is very secret string"

	// creating dir for keys
	dirName := "/tmp/" + name
	setUp(dirName)
	defer tearDown(dirName)

	t.Run(name, func(t *testing.T) {

		// Generating keys
		err := Generate(dirName)
		if err != nil {
			t.Errorf("Cant generate keys: %s", err.Error())
		}

		// Loading Public
		publicKey, err := LoadPublicKey(dirName + "/public.pem")
		if err != nil {
			t.Errorf("Cant load public key: %s", err.Error())
		}

		// Loading Private
		privateKey, err := LoadPrivateKey(dirName + "/private.pem")
		if err != nil {
			t.Errorf("Cant load private key: %s", err.Error())
		}

		// Encrypting string
		encString, err := Encrypt(publicKey, []byte(testString), []byte(name))
		if err != nil {
			t.Errorf("Failed to encrypt string: %s", err.Error())
		}

		// Decrypting string
		decString, err := Decrypt(privateKey, encString, []byte(name))
		if err != nil {
			t.Errorf("Failed to encrypt string: %s", err.Error())
		}

		if string(decString) != testString {
			t.Errorf("Decrypted string does not match the original: "+
				"have: %s, want: %s", string(decString), testString)
		}
	})

}
