# Copyright 2018 The zipper team Authors
# This file is part of the z0 library.
#
# The z0 library is free software: you can redistribute it and/or modify
# it under the terms of the GNU Lesser General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# The z0 library is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
# GNU Lesser General Public License for more details.
#
# You should have received a copy of the GNU Lesser General Public License
# along with the z0 library. If not, see <http://www.gnu.org/licenses/>.

TEST = $(shell go list ./...)
all:
	@go install ./cmd/z0
	go build -o z0 ./cmd/z0/main.go

run:
	@./z0
stop:
clear:
test:
	@echo $(TEST)
	go test $(TEST)

.PHONY: all run stop clear test
