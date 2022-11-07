go build -o bookings.exe ./cmd/web/
go test -coverprofile=coverage.out && go tool cover -html=coverage.out
bookings.exe