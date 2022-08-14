package utils_test

import (
	"path/filepath"
	"testing"

	"github.com/evgeniums/go-backend-helpers/utils"
)

var (
	rsaPubKey  = filepath.Join(filepath.Dir(scriptPath), "test_assets/pubkey.pem")
	rsaPrivKey = filepath.Join(filepath.Dir(scriptPath), "test_assets/privkey.pem")
)

func TestRsa(t *testing.T) {

	privKey, err := utils.LoadRsaPrivateKey("", rsaPrivKey, "")
	if err != nil {
		t.Fatalf("Failed to load private key: %s", err)
	}

	pubKey, err := utils.LoadRsaPublicKey(rsaPubKey)
	if err != nil {
		t.Fatalf("Failed to load public key: %s", err)
	}

	content := "Hello world!"

	signature, err := utils.RsaSign(privKey, content)
	if err != nil {
		t.Fatalf("Failed to sign data: %s", err)
	}

	err = utils.RsaVerify(pubKey, content, signature)
	if err != nil {
		t.Fatalf("Failed to verify data: %s", err)
	}
}
