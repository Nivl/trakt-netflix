// Package secret contains a type that can be used to store secrets.
// Secrets are protected against accidental leaks by returning the string
// "**REDACTED**" when the secret is printed or marshalled.
// In case of a bypass, a pointer address will be printed instead.
// This package is not meant to be secure against attacks, but rather
// to prevent accidental leaks (such as a secret being sent to the logger).
package secret

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// Secret represents a secret string that is protected against accidental leaks
type Secret struct {
	secret *string
}

// NewSecret stores the provided string in a Secret and returns it
func NewSecret(secret string) Secret {
	return Secret{
		secret: &secret,
	}
}

// Get returns the secret
func (s Secret) Get() string {
	if s.secret == nil {
		panic("secret is not set")
	}
	return *s.secret
}

// EnvDecode implements the envconfig.Decoder interface
func (s *Secret) EnvDecode(val string) error {
	s.secret = &val
	return nil
}

// String implements the fmt.Stringer interface (%s).
// It returns "**REDACTED**"
func (s Secret) String() string {
	return "**REDACTED**"
}

// GoString implements the fmt.GoStringer interface (%#v).
// It returns "**REDACTED**"
func (s Secret) GoString() string {
	return "**REDACTED**"
}

// MarshalJSON implements the json.Marshaler interface.
// It returns "**REDACTED**"
func (s Secret) MarshalJSON() ([]byte, error) {
	return []byte(`"**REDACTED**"`), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (s *Secret) UnmarshalJSON(data []byte) error {
	cpy, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	s.secret = &cpy
	return nil
}

// GobEncode implements the gob.GobEncoder interface.
func (s Secret) GobEncode() ([]byte, error) {
	return s.MarshalBinary()
}

// GobDecode implements the gob.GobDecoder interface.
func (s *Secret) GobDecode(data []byte) error {
	return s.UnmarshalBinary(data)
}

// MarshalText implements the encoding.TextMarshaler interface
// It returns "**REDACTED**"
func (s Secret) MarshalText() ([]byte, error) {
	return []byte(`**REDACTED**`), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (s *Secret) UnmarshalText(data []byte) error {
	cpy := string(data)
	s.secret = &cpy
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
// It returns "**REDACTED**"
func (s Secret) MarshalBinary() ([]byte, error) {
	redacted := "**REDACTED**"

	buf := new(bytes.Buffer)
	for _, c := range redacted {
		err := binary.Write(buf, binary.LittleEndian, c)
		if err != nil {
			return nil, fmt.Errorf("binary.Write failed: %w", err)
		}
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (s *Secret) UnmarshalBinary(data []byte) error {
	reader := bytes.NewBuffer(data)
	out := new(bytes.Buffer)
	var char rune

	var err error
	charSize := 4
	nbChars := len(data) / charSize
	for range nbChars {
		err = binary.Read(reader, binary.LittleEndian, &char)
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("binary.Read failed: %w", err)
		}
		out.WriteRune(char)
	}

	cpy := out.String()
	s.secret = &cpy
	return nil
}
