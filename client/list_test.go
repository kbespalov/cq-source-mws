package client

import (
	"context"
	"errors"
	"testing"
)

type pageRequest struct {
	PageSize  *int
	PageToken *string
}

func (r pageRequest) WithPageToken(token *string) pageRequest {
	r.PageToken = token
	return r
}

type barePageRequest struct {
	PageToken *string
}

func (r barePageRequest) WithPageToken(token *string) barePageRequest {
	r.PageToken = token
	return r
}

type pageResponse struct {
	items []string
	next  *string
}

func (r pageResponse) GetNextPageToken() *string { return r.next }
func (r pageResponse) GetItems() []string        { return r.items }

func TestWithPageSize(t *testing.T) {
	got := WithPageSize(pageRequest{})
	if got.PageSize == nil || *got.PageSize != listPageSize {
		t.Fatalf("PageSize = %v, want %d", got.PageSize, listPageSize)
	}

	custom := 25
	kept := WithPageSize(pageRequest{PageSize: &custom})
	if kept.PageSize == nil || *kept.PageSize != 25 {
		t.Fatalf("explicit PageSize = %v, want 25", kept.PageSize)
	}

	bare := WithPageSize(barePageRequest{})
	if bare.PageToken != nil {
		t.Fatalf("request without PageSize was changed: %+v", bare)
	}
}

func TestListSendsPageSizeOnEveryPage(t *testing.T) {
	var sizes []*int
	res := make(chan any, 4)
	next := "next"
	err := List(context.Background(), pageRequest{}, func(_ context.Context, req pageRequest) (pageResponse, error) {
		sizes = append(sizes, req.PageSize)
		if req.PageToken == nil {
			return pageResponse{items: []string{"a"}, next: &next}, nil
		}
		return pageResponse{items: []string{"b"}}, nil
	}, res, nil)
	if err != nil {
		t.Fatal(err)
	}
	close(res)
	var got []any
	for v := range res {
		got = append(got, v)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("items = %v", got)
	}
	if len(sizes) != 2 {
		t.Fatalf("pages = %d, want 2", len(sizes))
	}
	for i, size := range sizes {
		if size == nil || *size != listPageSize {
			t.Errorf("page %d size = %v, want %d", i, size, listPageSize)
		}
	}
}

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
