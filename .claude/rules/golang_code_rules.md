# Go Code Rules for multipay-india

## Rule 1: No unhandled errors
Every error must be either returned (wrapped with `%w`) or logged; never use `_ = err`.

❌ Wrong: `_, err := doSomething()` (error ignored)
✅ Correct: `_, err := doSomething(); if err != nil { return fmt.Errorf("context: %w", err) }`

## Rule 2: Close HTTP response bodies
Every HTTP response from an SDK call must be closed immediately via defer.

❌ Wrong: `data, _, err := client.Do(ctx)` (response body leaks)
✅ Correct: `resp, _, err := client.Do(ctx); defer func() { if resp != nil && resp.Body != nil { resp.Body.Close() } }()`
