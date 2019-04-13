package main

import (
	"strings"
	"testing"
)

func TestSplitMetadataEmpty(t *testing.T) {
	_, _, err := splitMetadata(nil)
	if err == nil {
		t.Error("expected error, not none")
	}
	_, _, err = splitMetadata([]byte(""))
	if err == nil {
		t.Error("expected error, got none")
	}
}

func TestSplitMetadata(t *testing.T) {
	const expMetaTxt = "some meta text"
	const expBodyTxt = "just some leftover text"

	var b strings.Builder
	b.WriteString(metaStartTag)
	b.WriteString(expMetaTxt)
	b.WriteString(metaEndTag)
	b.WriteString(expBodyTxt)

	data := []byte(b.String())
	meta, body, err := splitMetadata(data)

	if err != nil {
		t.Fatalf("parseMetadata(\"%s\") got unexpected error: %v", string(data), err)
	}

	if meta == nil {
		t.Errorf("parseMetadata(\"%s\") got nil expected %s", string(data), expMetaTxt)
	} else if string(meta) != expMetaTxt {
		t.Errorf("parseMetadata(\"%s\") got %s expected %s", string(data), string(meta), expMetaTxt)
	}

	if body == nil {
		t.Errorf("parseMetadata(\"%s\") got nil expected %s", string(data), expBodyTxt)
	} else if string(body) != expBodyTxt {
		t.Errorf("parseMetadata(\"%s\") got %s expected %s", string(data), string(body), expBodyTxt)
	}
}

func TestSplitMetadataNoMeta(t *testing.T) {
	const expBodyTxt = metaStartTag + "some garbo text <=|= there is no meta in this text"

	data := []byte(expBodyTxt)
	meta, body, err := splitMetadata(data)

	if err != nil {
		t.Fatalf("parseMetadata(\"%s\") got unexpected error: %v", string(data), err)
	}

	if meta == nil {
		t.Errorf("parseMetadata(\"%s\") got nil expected \"%s\"", string(data), "")
	} else if string(meta) != "" {
		t.Errorf("parseMetadata(\"%s\") got \"%s\" expected \"%s\"", string(data), string(meta), "")
	}

	if body == nil {
		t.Errorf("parseMetadata(\"%s\") got nil expected %s", string(data), expBodyTxt)
	} else if string(body) != expBodyTxt {
		t.Errorf("parseMetadata(\"%s\") got %s expected %s", string(data), string(body), expBodyTxt)
	}
}

func TestParseMetadata(t *testing.T) {

}
