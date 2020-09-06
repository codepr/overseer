// BSD 2-Clause License
//
// Copyright (c) 2020, Andrea Giacomo Baldan
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package internal contains all utilities and inner library features, not
// meant to be exported outside for the client
package internal

import (
	"time"
)

type URL = string

// ServerStatus defines the current state of a monitored server, URL to
// identify it, alive status, response time of the last request along with the
// status code and content
type ServerStatus struct {
	Url             URL           `json:"url"`
	Alive           bool          `json:"alive"`
	ResponseTime    time.Duration `json:"response_time"`
	ResponseStatus  int           `json:"response_status"`
	ResponseContent string        `json:"response_content"`
}

// Stats holds the collected stats for each server ready to be dispatched to a
// front-end client
type Stats struct {
	Url             URL           `json:"url"`
	Alive           bool          `json:"alive"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
	Availability    float64       `json:"availability"`
	StatusCodes     map[int]int   `json:"status_codes"`
}
