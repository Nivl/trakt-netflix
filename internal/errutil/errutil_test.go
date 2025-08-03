package errutil_test

import (
	"errors"
	"testing"

	"github.com/Nivl/trakt-netflix/internal/errutil"
	"github.com/stretchr/testify/assert"
)

type closer struct {
	actualClose func() error
}

func (c closer) Close() error {
	return c.actualClose()
}

func TestCheck(t *testing.T) {
	t.Parallel()

	t.Run("Should call close and set the error", func(t *testing.T) {
		t.Parallel()

		errToReturn := errors.New("expected error")
		expectedErrMsg := "wrapper: " + errToReturn.Error()

		closed := false
		var err error
		c := closer{
			actualClose: func() error {
				closed = true
				return errToReturn
			},
		}

		errutil.RunAndSetError(c.Close, &err, "wrapper")
		assert.True(t, closed, "Close() should have been called")
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("Should call close and NOT set the error", func(t *testing.T) {
		t.Parallel()

		errToReturn := errors.New("expected error")

		closed := false
		err := errToReturn
		c := closer{
			actualClose: func() error {
				closed = true
				return errors.New("unexpected error")
			},
		}

		errutil.RunAndSetError(c.Close, &err, "wrapper")
		assert.True(t, closed, "Close() should have been called")
		assert.Equal(t, errToReturn, err)
	})

	t.Run("Should call close and set the error with a wrapped message", func(t *testing.T) {
		t.Parallel()

		errToReturn := errors.New("expected error")
		expectedErrMsg := "wrapper: " + errToReturn.Error()

		closed := false
		var err error
		c := closer{
			actualClose: func() error {
				closed = true
				return errToReturn
			},
		}

		errutil.RunAndSetError(c.Close, &err, "wrapper")
		assert.True(t, closed, "Close() should have been called")
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("Should do nothing", func(t *testing.T) {
		t.Parallel()

		closed := false
		var err error
		c := closer{
			actualClose: func() error {
				closed = true
				return nil
			},
		}

		errutil.RunAndSetError(c.Close, &err, "shouldn't show this")
		assert.True(t, closed, "Close() should have been called")
		assert.NoError(t, err)
	})
}
