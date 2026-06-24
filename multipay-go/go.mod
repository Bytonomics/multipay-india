module github.com/Bytonomics/multipay-india/multipay-go

go 1.26

toolchain go1.26.0

require (
	github.com/SmrutAI/pedantigo v1.1.4
	github.com/bojanz/currency v1.3.0
	github.com/cashfree/cashfree-pg/v6 v6.0.5
	github.com/razorpay/razorpay-go v1.4.1
)

require (
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cashfree/cashfree-pg/v3 v3.2.14 // indirect
	github.com/invopop/jsonschema v0.13.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/cockroachdb/apd/v3 v3.2.1 // indirect; transitive dep of bojanz/currency (decimal arithmetic, NOT a database)
	github.com/getsentry/sentry-go v0.29.1 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)
