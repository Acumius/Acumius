package trust

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"strings"

	"github.com/mr-tron/base58/base58"
)

// GenerateKeypair generates a new Ed25519 keypair.
func GenerateKeypair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return pub, priv, nil
}

// GenerateDID formats a public key into the did:acumius:<base58_pubkey> DID format.
func GenerateDID(pubKey ed25519.PublicKey) string {
	encoded := base58.Encode(pubKey)
	return "did:acumius:" + encoded
}

// ParseDID parses a did:acumius:<base58_pubkey> DID format back into an Ed25519 public key.
func ParseDID(did string) (ed25519.PublicKey, error) {
	if !strings.HasPrefix(did, "did:acumius:") {
		return nil, errors.New("invalid DID prefix: must start with did:acumius:")
	}

	encoded := strings.TrimPrefix(did, "did:acumius:")
	if len(encoded) == 0 {
		return nil, errors.New("invalid DID: empty base58 key")
	}

	pubKeyBytes, err := base58.Decode(encoded)
	if err != nil {
		return nil, err
	}

	if len(pubKeyBytes) != ed25519.PublicKeySize {
		return nil, errors.New("invalid public key size")
	}

	return ed25519.PublicKey(pubKeyBytes), nil
}
