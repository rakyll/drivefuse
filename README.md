drivefuse
=
	go build -v -ldflags -linkmode=external main.go
	./main -mountpoint=<path> [-datadir]