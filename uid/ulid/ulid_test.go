package ulid_test

import (
	crand "crypto/rand"
	"fmt"
	ulidPkg "github.com/oklog/ulid/v2"
	"math/rand"
	"testing"
	"time"

	"github.com/revelaction/go-srs/uid/ulid"
	"lukechampine.com/frand"
)

// String Representation
//
//       01AN4Z07BY      79KA1307SR9X4MV3
//      |----------|    |----------------|
//       Timestamp           Entropy
//        10 chars           16 chars
//         48bits             80bits
//         base32             base32
func ExampleFixedEntropy() {

	// fixed  entropy value
	ti := time.Unix(1000000, 0)

	// The returned Monotonic io.Reader isn't safe for concurrent use.
	entropy := ulidPkg.Monotonic(rand.New(rand.NewSource(ti.UnixNano())), 0)

	uid := ulid.New(entropy)

	u := uid.Create()
	entropyPart := u[len(u)-16:]
	fmt.Printf("%s\n", entropyPart)

	u = uid.Create()
	entropyPart = u[len(u)-16:]
	fmt.Printf("%s\n", entropyPart)

	u = uid.Create()
	entropyPart = u[len(u)-16:]
	fmt.Printf("%s\n", entropyPart)

	//Output:
	//MQJHBF4QX1EFD6Y3
	//MQJHBF4QX3A9X93K
	//MQJHBF4QX73HGXVC
}

// https://github.com/lukechampine/frand
// frand is a fast-key-erasure CSPRNG in userspace. Its output is sourced from the keystream of a ChaCha cipher
// At init time, a "master" generator is created using an initial key sourced
// from crypto/rand. New generators source their initial keys from this master
// generator. This means the frand package only reads system entropy once, at
// startup.
// RNG is concurrent safe it can be given to many ulid objects.
//
// go test ./... -bench=. -run=Bench
func BenchmarkFrand(b *testing.B) {

	rng := frand.New()
	uid := ulid.New(rng)

	for i := 0; i < b.N; i++ {
		_ = uid.Create()
	}
}

func BenchmarkMonotonic(b *testing.B) {

	// fixed  entropy value
	ti := time.Unix(1000000, 0)
	// The returned Monotonic io.Reader isn't safe for concurrent use.
	entropy := ulidPkg.Monotonic(rand.New(rand.NewSource(ti.UnixNano())), 0)

	uid := ulid.New(entropy)

	for i := 0; i < b.N; i++ {
		_ = uid.Create()
	}
}

func BenchmarkCryptoRand(b *testing.B) {

	// The returned Monotonic io.Reader isn't safe for concurrent use.
	entropy := ulidPkg.Monotonic(crand.Reader, 0)

	uid := ulid.New(entropy)

	for i := 0; i < b.N; i++ {
		_ = uid.Create()
	}
}

func TestValidate(t *testing.T) {

	ti := time.Unix(1000000, 0)
	entropy := ulidPkg.Monotonic(rand.New(rand.NewSource(ti.UnixNano())), 0)
	uid := ulid.New(entropy)

	type test struct {
		input string
		want  []string
	}

	tests := []struct {
		input string
		want  bool
	}{
		{input: "01ET2F9NGWM6R96PJYCZ7N3XG3", want: true},
		// ULIDs with the wrong data size.
		{input: "01ET2F9NGWM6R96PJYCZ7N3XG", want: false},
		// invalid Base32 encodings
		{input: "01ET2F9NGWM6R96PJYCZ7N3XGO", want: false},
		// first character is larger than 7, thereby exceeding the valid bit depth of 128
		{input: "81ET2F9NGWM6R96PJYCZ7N3XG3", want: false},
	}

	for _, tc := range tests {
		err := uid.Validate(tc.input)
		t.Logf("☑  validate error : %v", err)

		if (err == nil) != tc.want {
			t.Errorf("\nfor id %s want %t error", tc.input, tc.want)
		}
	}
}
