install:
	GO15VENDOREXPERIMENT=1 go install
	GO15VENDOREXPERIMENT=1 go install ./cmd/tfcat
	GO15VENDOREXPERIMENT=1 go install ./cmd/tfmerge
	GO15VENDOREXPERIMENT=1 go install ./cmd/tfdedup
	GO15VENDOREXPERIMENT=1 go install ./cmd/tfindex
	GO15VENDOREXPERIMENT=1 go install ./cmd/tfsort
	GO15VENDOREXPERIMENT=1 go install ./cmd/tfgrep
