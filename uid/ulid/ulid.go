package ulid

import (
	ulidPkg "github.com/oklog/ulid/v2"
	"io"
)

type Ulid struct {
	Entropy io.Reader
}

func New(entropy io.Reader) *Ulid {
	return &Ulid{
		Entropy: entropy,
	}
}

// TODO for test we do not want now
func (u *Ulid) Create() string {
	return ulidPkg.MustNew(ulidPkg.Now(), u.Entropy).String()
}

func (u *Ulid) Validate(uid string) error {
	_, err := ulidPkg.ParseStrict(uid)
	if err != nil {
		return err
	}

	return nil
}
