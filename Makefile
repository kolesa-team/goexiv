.PHONY: test test-concurrent-3 test-concurrent-10 test-concurrent-100

test:
	go test -v ./...

test-concurrent-3:
	go test -run Test_GetBytes_Goroutine -count=3 -v

test-concurrent-10:
	go test -run Test_GetBytes_Goroutine -count=10 -v

test-concurrent-100:
	go test -run Test_GetBytes_Goroutine -count=100 -v
