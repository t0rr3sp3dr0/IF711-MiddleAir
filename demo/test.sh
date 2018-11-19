#!/bin/bash
go test -bench=. -benchmem -benchtime 1h -cpuprofile cpu.prof -memprofile mem.prof -vet off
