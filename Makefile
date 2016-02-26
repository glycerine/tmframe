install:
	go install
	go install ./cmd/tfcat
	go install ./cmd/tfmerge
	go install ./cmd/tfdedup
	go install ./cmd/tfindex
