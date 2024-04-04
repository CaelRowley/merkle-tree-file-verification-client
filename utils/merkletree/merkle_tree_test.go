package merkletree

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"testing"
)

func TestBuildTree(t *testing.T) {
	leafData := []string{
		"hash1",
		"hash2",
		"hash3",
		"hash4",
		"hash5",
	}

	var leafHashes [][]byte

	for _, leaf := range leafData {
		hash := sha256.Sum256([]byte(leaf))
		leafHashes = append(leafHashes, hash[:])
	}

	root := BuildTree(leafHashes)
	got := hex.EncodeToString(root.Hash[:])
	want := "1726c9d7c9f5585c6657edb9f5de6ee2f14c447d2fb80c9083a2572857702912"

	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestNewNode(t *testing.T) {
	left := &Node{Hash: []byte("leftHash")}
	right := &Node{Hash: []byte("rightHash")}

	got := newNode([]byte("someHash"), left, right)
	want := &Node{
		Hash:  []byte("someHash"),
		Left:  left,
		Right: right,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestHashPair(t *testing.T) {
	left := sha256.Sum256([]byte("hash1"))
	right := sha256.Sum256([]byte("hash2"))
	hashedPair := hashPair(left[:], right[:])

	got := hex.EncodeToString(hashedPair)
	want := "e6a8cc2a789a8e72fced42d013d87acb0c29f83e6d7716ab2bd92ee74f54a2da"

	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
