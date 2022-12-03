package sinoname

import (
	"reflect"
	"testing"
)

func TestTokenizeDefault(t *testing.T) {
	type testCase struct {
		in  string
		out []string
	}

	for _, tCase := range []testCase{
		{"", []string{""}},
		{"lowercase", []string{"lowercase"}},
		{"Class", []string{"Class"}},
		{"MyClass", []string{"My", "Class"}},
		{"MyC", []string{"My", "C"}},
		{"HTML", []string{"HTML"}},
		{"PDFLoader", []string{"PDF", "Loader"}},
		{"AString", []string{"A", "String"}},
		{"SimpleXMLParser", []string{"Simple", "XML", "Parser"}},
		{"vimRPCPlugin", []string{"vim", "RPC", "Plugin"}},
		{"GL11Version", []string{"GL", "11", "Version"}},
		{"Alpha1Testing", []string{"Alpha", "1Testing"}},
		{"Alpha12Testing", []string{"Alpha", "12", "Testing"}},
		{"99Bottles", []string{"99", "Bottles"}},
		{"8ottles", []string{"8ottles"}},
		{"May5", []string{"May5"}},
		{"BFG9000", []string{"BFG", "9000"}},
		{"BöseÜberraschung", []string{"Böse", "Überraschung"}},
		{"Two  spaces", []string{"Two", "spaces"}},
		{"BadUTF8\xe2\xe2\xa1", []string{"BadUTF8\xe2\xe2\xa1"}},
	} {
		if v := tokenizeDefault(tCase.in); !reflect.DeepEqual(tCase.out, v) {
			t.Fatalf("%#v != %#v", tCase.out, v)
		}
	}
}

func TestStripNumbersASCII(t *testing.T) {
	in := "123L᠑mb3ls34"

	letters, numbers := stripNumbersASCII(in)
	if letters != "L᠑mbls" || numbers != "123334" {
		t.Fatalf("123L᠑mb3ls34 -> L᠑mbls, 123334 but got %v, %v", letters, numbers)
	}
}

func TestStripNumbersUnicode(t *testing.T) {
	in := "123L᠑mb3ls34"

	letters, numbers := StripNumbersUnicode(in)
	if letters != "Lmbls" || numbers != "123᠑334" {
		t.Fatalf("123L᠑mb3ls34 -> Lmbls, 123᠑334 but got %v, %v", letters, numbers)
	}
}
