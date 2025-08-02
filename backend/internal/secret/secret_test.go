package secret_test

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Nivl/trakt-netflix/internal/secret"
	"github.com/sethvargo/go-envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Used to test envconfig
type lookuper struct {
	env map[string]string
}

func (l *lookuper) Lookup(key string) (string, bool) {
	v, ok := l.env[key]
	return v, ok
}

const (
	mySecret = "secret string"
	redacted = "**REDACTED**"
)

func TestNewSecret(t *testing.T) {
	t.Parallel()

	t.Run("Get() should return the data", func(t *testing.T) {
		t.Parallel()

		s := secret.NewSecret(mySecret)
		assert.Equal(t, mySecret, s.Get())
	})

	t.Run("Stringer should return redacted data", func(t *testing.T) {
		t.Parallel()

		s := secret.NewSecret(mySecret)
		assert.Equal(t, redacted, s.String())
	})

	t.Run("Stringer should return redacted data", func(t *testing.T) {
		t.Parallel()

		s := secret.NewSecret(mySecret)
		assert.Equal(t, redacted, fmt.Sprintf("%#v", s))
	})

	t.Run("binary", func(t *testing.T) {
		t.Parallel()

		t.Run("should unmarshal the data", func(t *testing.T) {
			buf := new(bytes.Buffer)
			for _, c := range mySecret {
				err := binary.Write(buf, binary.LittleEndian, c)
				require.NoError(t, err)
			}
			binarySecret := buf.Bytes()
			require.Len(t, binarySecret, len(mySecret)*4)

			data := new(secret.Secret)
			err := data.UnmarshalBinary(binarySecret)
			require.NoError(t, err)
			assert.Len(t, data.Get(), len(mySecret))
			assert.Equal(t, mySecret, data.Get())
		})

		t.Run("should redact the data when marshalling", func(t *testing.T) {
			s := secret.NewSecret(mySecret)
			out, err := s.MarshalBinary()
			require.NoError(t, err)

			// We just unmarshal it to check, so we don't have
			// to write a complex method that decodes the data.
			// It should create a secret with **REDACTED** as secret
			data := new(secret.Secret)
			err = data.UnmarshalBinary(out)
			require.NoError(t, err)

			assert.Equal(t, redacted, s.String())
		})
	})

	t.Run("JSON", func(t *testing.T) {
		t.Parallel()

		type jsonStruct struct {
			Secret secret.Secret `json:"secret"`
		}

		t.Run("should unmarshal the data", func(t *testing.T) {
			t.Parallel()

			data := new(jsonStruct)
			err := json.Unmarshal([]byte(`{"secret":"`+mySecret+`"}`), &data)
			require.NoError(t, err)
			assert.Equal(t, mySecret, data.Secret.Get())
		})

		t.Run("should not trim too many quotes when unmarshalling the data", func(t *testing.T) {
			t.Parallel()

			data := new(jsonStruct)
			err := json.Unmarshal([]byte(`{"secret":"\"\""}`), data)
			require.NoError(t, err)
			assert.Equal(t, `""`, data.Secret.Get())
		})

		t.Run("should redact the data when marshalling", func(t *testing.T) {
			t.Parallel()

			s := secret.NewSecret(mySecret)
			data := jsonStruct{
				Secret: s,
			}

			out, err := json.Marshal(data)
			require.NoError(t, err)

			assert.Contains(t, string(out), redacted)
			assert.NotContains(t, string(out), mySecret)
		})
	})

	t.Run("should work with sethvargo/go-envconfig", func(t *testing.T) {
		t.Parallel()

		type envStruct struct {
			Secret secret.Secret `env:"SECRET"`
		}
		finder := &lookuper{
			env: map[string]string{
				"SECRET": mySecret,
			},
		}

		var cfg envStruct
		err := envconfig.ProcessWith(t.Context(), &envconfig.Config{
			Target:   &cfg,
			Lookuper: finder,
		})
		require.NoError(t, err)
		assert.Equal(t, mySecret, cfg.Secret.Get())
	})
}
