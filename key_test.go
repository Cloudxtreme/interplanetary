package interplanetary

import "testing"

func TestKeyFromString(t *testing.T) {
	maybe := "Qmf7UC9uXXTxhmYHPJaBsDureMsth3wJzCg4kSTzPV5WBn"
	k, err := parseKey(maybe)
	if err != nil {
		t.Fatal(err)
	}
	if k.String() != maybe {
		t.Fatal("transformation changed the key")
	}
}
