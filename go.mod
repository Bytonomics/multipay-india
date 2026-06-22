module github.com/Bytonomics/multipay-adapter

go 1.26

toolchain go1.26.0

require (
	github.com/bojanz/currency v1.3.0
	github.com/cashfree/cashfree_pg v0.3.0
	github.com/razorpay/razorpay-go v1.4.1
)

require (
	github.com/cockroachdb/apd/v3 v3.2.1 // indirect; transitive dep of bojanz/currency (decimal arithmetic, NOT a database)
	github.com/getsentry/sentry-go v0.18.0 // indirect
	golang.org/x/sys v0.0.0-20220928140112-f11e5e49a4ec // indirect
	golang.org/x/text v0.3.7 // indirect
)
