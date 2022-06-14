package did

import (
	"crypto/rand"
	"os"
	"testing"

	"github.com/tcfw/didem/pkg/did"
)

func TestNewFileStore(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	f, err := NewFileStore(home + "/.didem/identity.yaml")
	if err != nil {
		t.Fatal(err)
	}

	id, err := did.GenerateEd25519Identity(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	if err := f.Add(id); err != nil {
		t.Fatal(err)
	}
}
