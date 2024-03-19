package background

import "log/slog"

// Go starts a new goroutine to execute the provided function concurrently.
func Go(fn func()) {
	// Starts a new anonymous goroutine.
	go func() {
		// Defer a function call to recover from any panics that might occur during execution.
		defer func() {
			// Recover from any panics and handle them.
			if r := recover(); r != nil {
				// Log an error message using slog.Error if a panic is recovered.
				slog.Error("background goroutine panic", "recover", r)
			}
		}()
		// Execute the provided function.
		fn()
	}()
}
