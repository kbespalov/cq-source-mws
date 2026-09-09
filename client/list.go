package client

import (
	"context"

	"go.mws.cloud/go-sdk/mws/page"
)

// List drains a paginated list call into the channel the resolver was handed.
//
// Every table in this plugin fetches the same way, because every MWS list
// endpoint is paginated the same way, so the loop lives here instead of once
// per table. skip decides whether an error is worth failing the sync over; see
// Client.Skip.
func List[Token ~string, T any, Req page.Request[Req], Resp page.Response[Token, T]](
	ctx context.Context,
	req Req,
	do page.Do[Token, T, Req, Resp],
	res chan<- any,
	skip func(error) bool,
) error {
	for item, err := range page.NewPager(req, do).All(ctx) {
		if err != nil {
			if skip != nil && skip(err) {
				return nil
			}
			return err
		}
		res <- item
	}
	return nil
}

// Fetch is List for the handful of endpoints that return everything at once.
// A project has a few Kubernetes release channels and a few versions in each,
// so those listings take no page token and there is nothing to drain.
func Fetch[T any](call func() ([]T, error), res chan<- any, skip func(error) bool) error {
	items, err := call()
	if err != nil {
		if skip != nil && skip(err) {
			return nil
		}
		return err
	}
	for _, item := range items {
		res <- item
	}
	return nil
}
