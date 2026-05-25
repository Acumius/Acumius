package trust

import (
	"crypto/ed25519"
	"testing"
)

func TestIdentityAndDID(t *testing.T) {
	pub, priv, err := GenerateKeypair()
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}

	if len(pub) != ed25519.PublicKeySize {
		t.Errorf("expected public key size %d, got %d", ed25519.PublicKeySize, len(pub))
	}

	if len(priv) != ed25519.PrivateKeySize {
		t.Errorf("expected private key size %d, got %d", ed25519.PrivateKeySize, len(priv))
	}

	did := GenerateDID(pub)
	if len(did) == 0 {
		t.Fatal("generated DID is empty")
	}

	parsedPub, err := ParseDID(did)
	if err != nil {
		t.Fatalf("failed to parse generated DID: %v", err)
	}

	for i := range pub {
		if pub[i] != parsedPub[i] {
			t.Errorf("mismatch at byte %d: expected %d, got %d", i, pub[i], parsedPub[i])
		}
	}
}

func TestInvalidDIDs(t *testing.T) {
	tests := []struct {
		name    string
		did     string
		wantErr bool
	}{
		{
			name:    "empty DID",
			did:     "",
			wantErr: true,
		},
		{
			name:    "wrong prefix",
			did:     "did:other:12345",
			wantErr: true,
		},
		{
			name:    "empty key part",
			did:     "did:acumius:",
			wantErr: true,
		},
		{
			name:    "invalid base58 characters",
			did:     "did:acumius:invalidIO0O", // O and I are invalid in base58
			wantErr: true,
		},
		{
			name:    "invalid public key size",
			did:     "did:acumius:123", // too short
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseDID(tc.did)
			if (err != nil) != tc.wantErr {
				t.Errorf("expected error presence: %v, got error: %v", tc.wantErr, err)
			}
		})
	}
}
