package passwordhash

import (
	"strings"
	"testing"
)

var fastTestParameters = Parameters{MemoryInKibibytes: 1024, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32}

func TestHashAndVerifyRoundTrip(t *testing.T) {
	encodedHash, err := Hash("Mining4Crystal!Moon", fastTestParameters)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if !strings.HasPrefix(encodedHash, "$argon2id$v=19$m=1024,t=1,p=1$") {
		t.Fatalf("unexpected encoding %q", encodedHash)
	}

	isMatch, err := Verify("Mining4Crystal!Moon", encodedHash)
	if err != nil || !isMatch {
		t.Fatalf("expected the correct password to match, got %v, %v", isMatch, err)
	}
	isMatch, err = Verify("mining4crystal!moon", encodedHash)
	if err != nil || isMatch {
		t.Fatalf("expected a wrong password to be rejected, got %v, %v", isMatch, err)
	}
}

func TestHashUsesUniqueSalts(t *testing.T) {
	firstHash, _ := Hash("same-password-123", fastTestParameters)
	secondHash, _ := Hash("same-password-123", fastTestParameters)
	if firstHash == secondHash {
		t.Fatal("two hashes of the same password must differ")
	}
}

func TestDefaultParametersFollowOwaspMinimum(t *testing.T) {
	if DefaultParameters.MemoryInKibibytes < 19*1024 || DefaultParameters.Iterations < 2 {
		t.Fatalf("default parameters %+v are below the OWASP argon2id minimum", DefaultParameters)
	}
}

func TestVerifyRejectsMalformedHashes(t *testing.T) {
	for _, malformedHash := range []string{"", "plain-text", "$argon2i$v=19$m=1,t=1,p=1$c2FsdA$a2V5", "$argon2id$v=19$m=x$c2FsdA$a2V5"} {
		if _, err := Verify("anything", malformedHash); err == nil {
			t.Fatalf("expected an error for %q", malformedHash)
		}
	}
}
