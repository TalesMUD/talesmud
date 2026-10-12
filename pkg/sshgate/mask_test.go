package sshgate

import "testing"

func TestMaskFingerprintAndCode(t *testing.T) {
	got := MaskFingerprint("SHA256:abcdefghijklmnopqrstuvwxyz012345")
	if got != "SHA256:abcd…2345" {
		t.Fatalf("fp %q", got)
	}
	if MaskFingerprint("nope") != "SHA256:…" {
		t.Fatal("non fingerprint")
	}
	if MaskUserCode("BCDF-GHJK") != "BC**-****" {
		t.Fatalf("code %q", MaskUserCode("BCDF-GHJK"))
	}
	if MaskUserCode("") != "**-****" {
		t.Fatal("empty code")
	}
}
