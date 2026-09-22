package util

import "testing"

func TestInitViperJWT(t *testing.T) {
	v := InitViper("../conf", "jwt", YAML)
	if v.GetString("secret") == "" {
		t.Fatal("jwt secret should be configured")
	}
}
