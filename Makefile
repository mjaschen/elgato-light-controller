.PHONY: install
install: elgato-light-controller
	install $< ~/bin/elc

elgato-light-controller: *.go go.*
	go build
