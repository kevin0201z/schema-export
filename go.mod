module github.com/schema-export/schema-export

go 1.25.0

require (
	dm v0.0.0-00010101000000-000000000000
	github.com/sijms/go-ora/v2 v2.9.0
	github.com/spf13/cobra v1.10.2
)

require (
	github.com/golang/snappy v0.0.4 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace dm => ./internal/dm-go-driver/dm
