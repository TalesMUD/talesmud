package sshkeys

import "testing"

func FuzzParseAuthorizedKey(f *testing.F) {
	f.Add("")
	f.Add("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA comment")
	f.Add("ssh-dss AAAAB3NzaC1kc3MAAACB comment")
	f.Add("-----BEGIN PRIVATE KEY-----\nMIIE\n-----END PRIVATE KEY-----")
	f.Add("ssh-rsa AAAA\nssh-ed25519 AAAA")
	f.Add("ecdsa-sha2-nistp256 AAAA")
	f.Add("not a key")
	f.Fuzz(func(t *testing.T, line string) {
		_, _, _ = ParseAuthorizedKey(line)
	})
}
