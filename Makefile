.PHONY: test coverage

test:
	go test -v -race ./...

coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Abre coverage.html en el navegador para ver el reporte."
