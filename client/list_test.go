package client

import (
	"errors"
	"testing"
)

func TestFetchDrainsAndSkips(t *testing.T) {
	t.Run("items", func(t *testing.T) {
		var got []any
		res := make(chan any, 4)
		err := Fetch(func() ([]string, error) {
			return []string{"a", "b"}, nil
		}, res, nil)
		if err != nil {
			t.Fatal(err)
		}
		close(res)
		for v := range res {
			got = append(got, v)
		}
		if len(got) != 2 || got[0] != "a" || got[1] != "b" {
			t.Errorf("got %v", got)
		}
	})

	t.Run("skip", func(t *testing.T) {
		res := make(chan any, 1)
		err := Fetch(func() ([]string, error) {
			return nil, errors.New("denied")
		}, res, func(error) bool { return true })
		if err != nil {
			t.Fatalf("skip should swallow the error, got %v", err)
		}
	})

	t.Run("fatal", func(t *testing.T) {
		res := make(chan any, 1)
		want := errors.New("boom")
		err := Fetch(func() ([]string, error) {
			return nil, want
		}, res, nil)
		if !errors.Is(err, want) {
			t.Fatalf("got %v, want %v", err, want)
		}
	})
}
