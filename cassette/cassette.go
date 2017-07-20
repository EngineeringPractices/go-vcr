// Copyright (c) 2015 Marin Atanasov Nikolov <dnaeon@gmail.com>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer
//    in this position and unchanged.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE AUTHOR(S) ``AS IS'' AND ANY EXPRESS OR
// IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
// OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
// IN NO EVENT SHALL THE AUTHOR(S) BE LIABLE FOR ANY DIRECT, INDIRECT,
// INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
// NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
// THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package cassette

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
)

// Cassette format versions
const (
	cassetteFormatV1 = 1
)

var (
	// ErrInteractionNotFound indicates that a requested
	// interaction was not found in the cassette file
	ErrInteractionNotFound = errors.New("Requested interaction not found")
)

// Request represents a client request as recorded in the
// cassette file
type Request struct {
	// Body of request
	Body string `json:"body"`

	// Form values
	Form url.Values `json:"form"`

	// Request headers
	Headers http.Header `json:"headers"`

	// Request URL
	URL string `json:"url"`

	// Request method
	Method string `json:"method"`
}

// Response represents a server response as recorded in the
// cassette file
type Response struct {
	// Body of response
	Body string `json:"body"`

	// Response headers
	Headers http.Header `json:"headers"`

	// Response status message
	Status string `json:"status"`

	// Response status code
	Code int `json:"code"`
}

// Interaction type contains a pair of request/response for a
// single HTTP interaction between a client and a server
type Interaction struct {
	Request  `json:"request"`
	Response `json:"response"`
}

// Matcher function returns true when the actual request matches
// a single HTTP interaction's request according to the function's
// own criteria.
type Matcher func(*http.Request, Request) bool

// Default Matcher is used when a custom matcher is not defined
// and compares only the method and URL.
func DefaultMatcher(r *http.Request, i Request) bool {
	return r.Method == i.Method && r.URL.String() == i.URL
}

// Cassette type
type Cassette struct {
	// Name of the cassette
	Name string `json:"-"`

	// File name of the cassette as written on disk
	File string `json:"-"`

	// Cassette format version
	Version int `json:"version"`

	sync.RWMutex
	// Interactions between client and server
	Interactions []*Interaction `json:"interactions"`

	// Matches actual request with interaction requests.
	Matcher Matcher `json:"-"`
}

// New creates a new empty cassette
func New(name string) *Cassette {
	c := &Cassette{
		Name:         name,
		File:         fmt.Sprintf("%s.json", name),
		Version:      cassetteFormatV1,
		Interactions: make([]*Interaction, 0),
		Matcher:      DefaultMatcher,
	}

	return c
}

// Load reads a cassette file from disk
func Load(name string) (*Cassette, error) {
	c := New(name)
	data, err := ioutil.ReadFile(c.File)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &c)

	return c, err
}

// AddInteraction appends a new interaction to the cassette
func (c *Cassette) AddInteraction(i *Interaction) {
	c.Lock()
	c.Interactions = append(c.Interactions, i)
	c.Unlock()
}

// GetInteraction retrieves a recorded request/response interaction
func (c *Cassette) GetInteraction(r *http.Request) (*Interaction, error) {
	c.RLock()
	defer c.RUnlock()
	for _, i := range c.Interactions {
		if c.Matcher(r, i.Request) {
			return i, nil
		}
	}

	return nil, ErrInteractionNotFound
}

// Save writes the cassette data on disk for future re-use
func (c *Cassette) Save() error {
	c.RLock()
	defer c.RUnlock()
	// Save cassette file only if there were any interactions made
	if len(c.Interactions) == 0 {
		return nil
	}

	// Create directory for cassette if missing
	cassetteDir := filepath.Dir(c.File)
	if _, err := os.Stat(cassetteDir); os.IsNotExist(err) {
		if err = os.MkdirAll(cassetteDir, 0755); err != nil {
			return err
		}
	}

	// Marshal to YAML and save interactions
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	f, err := os.Create(c.File)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return nil
}
