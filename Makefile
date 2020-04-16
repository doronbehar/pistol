pistol:
	go build ./cmd/pistol
all: pistol

install:
	go install ./cmd/pistol

.PHONY: pistol
