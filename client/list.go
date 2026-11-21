package client

import (
	"context"
	"reflect"

	"go.mws.cloud/go-sdk/mws/page"
)

// listPageSize is the pageSize query sent on every paginated list. Generated
// requests leave the field empty, and then each service picks its own default,
// which is small enough that a large collection takes many round trips. A
// service may return fewer rows than asked, or cap the value; the pager still
// follows the next-page token.
const listPageSize = 1000

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
	for item, err := range page.NewPager(WithPageSize(req), do).All(ctx) {
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

// WithPageSize fills an empty PageSize with listPageSize. A request that
// already set the field, or that has no such field, is returned unchanged.
//
// The pager interface only carries the page token. Every generated MWS list
// request stores the query value in an exported *int named PageSize, so this
// one helper is how the default reaches every table, including the project
// list that does not go through List.
func WithPageSize[Req any](req Req) Req {
	v := reflect.ValueOf(&req).Elem()
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return req
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return req
	}
	field := v.FieldByName("PageSize")
	if !field.IsValid() || !field.CanSet() || field.Kind() != reflect.Pointer || !field.IsNil() {
		return req
	}
	if field.Type().Elem().Kind() != reflect.Int {
		return req
	}
	n := listPageSize
	field.Set(reflect.ValueOf(&n))
	return req
}
