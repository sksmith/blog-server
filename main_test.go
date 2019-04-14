package main

import (
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
	data := []byte(`<!--META--
author: Sean K Smith
--END-->
# My First Blog Post
`)
	expMetaTxt := `author: Sean K Smith`
	expBodyTxt := `# My First Blog Post`
	meta, body, err := splitMetadata(data)

	if err != nil {
		t.Fatalf("parseMetadata(body) got unexpected error: %v", err)
	}

	if meta == nil {
		t.Errorf("parseMetadata(body) got nil expected %s", expMetaTxt)
	} else if string(meta) != expMetaTxt {
		t.Errorf("parseMetadata(body) got \"%s\" expected \"%s\"", string(meta), expMetaTxt)
	}

	if body == nil {
		t.Errorf("parseMetadata(body) got nil expected %s", expBodyTxt)
	} else if string(body) != expBodyTxt {
		t.Errorf("parseMetadata(body) got %s expected %s", string(body), expBodyTxt)
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

func BechmarkSplitMetadata(b *testing.B) {
	data := []byte(metaStartTag + "some garbo text <=|= there is no meta in this text")
	for i := 0; i < b.N; i++ {
		splitMetadata(data)
	}
}
