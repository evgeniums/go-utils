package utils_test

import (
	"testing"

	"github.com/evgeniums/go-backend-helpers/utils"
)

type testHashDoc struct {
	Field1 string
	Field2 int
	Field3 string
	Field4 int
}

func TestHash(t *testing.T) {

	obj := &testHashDoc{"field one", 1, "field two", 20}
	h := utils.Hash(obj, "Field1", "Field4")
	sampleH := "feb0dee1f6115a2803766166052e5617d7c0d8f9941313ee69a68f52e1125f3d"
	if sampleH != h {
		t.Fatalf("Values mistmatch: expected %v, got %v", sampleH, h)
	}

	h = utils.Hash(obj)
	sampleH = "29eca917f9ec1a2e9222294753750c9ff34131a414bc3a99ca5fbfd35e2d7eef"
	if sampleH != h {
		t.Fatalf("Values mistmatch: expected %v, got %v", sampleH, h)
	}
}

func TestHmac(t *testing.T) {

	s := "secret--12345"
	obj := &testHashDoc{"field one", 1, "field two", 20}

	h := utils.Hmac(s, obj, "Field1", "Field4")
	sampleH := "12b4fcc97c8b5268bd79d519bd59181089d3be6cf565a1fd8975769a8170bd40"
	if sampleH != h {
		t.Fatalf("Values mistmatch: expected %v, got %v", sampleH, h)
	}

	h = utils.Hmac(s, obj)
	sampleH = "cb540821e8807b96de61f578654216d8366c54e7fb3d2e5da5a6bc004f28628f"
	if sampleH != h {
		t.Fatalf("Values mistmatch: expected %v, got %v", sampleH, h)
	}
}

func TestHashValue(t *testing.T) {
	h := utils.HashValue("just string")
	sampleH := "bdc371165c24879978c2fe005f2e80449e5121fb4464505aefcd267b8dc916fd"
	if sampleH != h {
		t.Fatalf("Values mistmatch: expected %v, got %v", sampleH, h)
	}

	h = utils.HashValue(12345678)
	sampleH = "ef797c8118f02dfb649607dd5d3f8c7623048c9c063d532cc95c5ed7a898a64f"
	if sampleH != h {
		t.Fatalf("Values mistmatch: expected %v, got %v", sampleH, h)
	}
}

func TestHmacValue(t *testing.T) {

	s := "secret--12345"

	h := utils.HmacValue(s, "just string")
	sampleH := "4e2ecc4cb8ec910af2304fa3c46b16c3d2588cf04f8ba674e76ef844a9d263f0"
	if sampleH != h {
		t.Fatalf("Values mistmatch: expected %v, got %v", sampleH, h)
	}

	h = utils.HmacValue(s, 12345678)
	sampleH = "320b7d5adfe494ecbac3b7a2c9d5f279d86113bf199ef585922fdd61cfa8b4b7"
	if sampleH != h {
		t.Fatalf("Values mistmatch: expected %v, got %v", sampleH, h)
	}
}

func TestHashVals(t *testing.T) {

	r := utils.HashValsHex("12345678", "eeaabbcc", "helloworld")
	expected := "131387112f3394ec8ecd96972c394a76a025283f0a660bc134bacef3e62b613e"
	if r != expected {
		t.Fatalf("Invalid hash: expected %v, got %v", expected, r)
	}
}
