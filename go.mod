module github.com/Bytonomics/multipay-adapter

go 1.26

toolchain go1.26.0

require (
	github.com/bojanz/currency v1.3.0
	github.com/cashfree/cashfree-pg/v6 v6.0.5
	github.com/razorpay/razorpay-go v1.4.1
)

require (
	github.com/cashfree/cashfree-pg/v3 v3.2.14 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
)

require (
	github.com/cockroachdb/apd/v3 v3.2.1 // indirect; transitive dep of bojanz/currency (decimal arithmetic, NOT a database)
	github.com/getsentry/sentry-go v0.29.1 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)
