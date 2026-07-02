package lightweigit

import (
	"context"
	"errors"
)

// // // // // // // // // // // // // // // //

// Send delivers v into out or gives up when ctx is canceled, so a stream
// blocked on a consumer that stopped reading never hangs forever.
func Send[T any](ctx context.Context, out chan<- T, v T) error {
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case out <- v:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// StreamPages drives an offset-based paginated listing (per_page/page style)
// and adaptively halves the page size whenever a page response exceeds the
// GetJSON body cap (ErrResponseTooLarge).
//
// The stream contract is append-only, so already-emitted items must never
// repeat and none may be skipped. The loop therefore tracks the absolute
// offset of the next item to deliver instead of a raw page counter: after a
// shrink, the page to fetch is offset/perPage+1 and the first offset%perPage
// items of that page are duplicates of what was already sent, so they are
// dropped. With an even page size s this reduces to the classic remap of old
// page p onto new pages (2p-1, 2p) of size s/2 with nothing to drop; the
// offset form also stays exact for odd sizes produced by limit clamping.
//
// Shrinking stops at perPage == 1: a single item bigger than the cap is a
// hard error and is returned as-is.
//
// emit delivers one item to the consumer; a non-nil error (typically
// ctx.Err() from Send) aborts the stream and is returned as-is.
func StreamPages[T any](ctx context.Context, perPage, limit int, fetch func(perPage, page int) ([]T, error), emit func(T) error) error {
	if fetch == nil || emit == nil {
		return errors.New("StreamPages: nil fetch or emit")
	}
	if perPage <= 0 {
		return errors.New("StreamPages: perPage must be positive")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if limit > 0 && limit < perPage {
		perPage = limit
	}

	sent := 0
	offset := 0
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		page := offset/perPage + 1
		skip := offset % perPage

		items, err := fetch(perPage, page)
		if err != nil {
			if errors.Is(err, ErrResponseTooLarge) && perPage > 1 {
				perPage /= 2
				continue
			}
			return err
		}
		// A page shorter than the duplicate prefix carries nothing new.
		if len(items) <= skip {
			return nil
		}

		lastPage := len(items) < perPage
		for _, it := range items[skip:] {
			if limit > 0 && sent >= limit {
				return nil
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}

			if err := emit(it); err != nil {
				return err
			}
			sent++
			offset++
		}

		if lastPage {
			return nil
		}
	}
}
