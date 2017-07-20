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

package vcr_test

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/flynn/go-vcr/cassette"
	"github.com/flynn/go-vcr/recorder"
)

var recording = os.Getenv("RECORDING") != ""

func newRecorder(name string) (*recorder.Recorder, error) {
	mode := recorder.ModeReplaying
	if recording {
		mode = recorder.ModeRecording
	}
	return recorder.NewAsMode(name, mode, nil)
}

func TestSimple(t *testing.T) {
	// Start our recorder
	r, err := newRecorder("fixtures/golang-org")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Stop() // Make sure recorder is stopped once done with it

	// Create an HTTP client and inject our transport
	client := &http.Client{
		Transport: r, // Inject as transport!
	}

	u := "http://golang.org/"
	resp, err := client.Get(u)
	if err != nil {
		t.Fatalf("Failed to get url %s: %s", u, err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("Failed to read response body: %s", err)
	}

	wantTitle := "<title>The Go Programming Language</title>"
	bodyContent := string(body)

	if !strings.Contains(bodyContent, wantTitle) {
		t.Errorf("Title %s not found in response", wantTitle)
	}

	if !recording {
		_, err = client.Get(u)
		urlErr, ok := err.(*url.Error)
		if !ok || urlErr.Err != cassette.ErrInteractionNotFound {
			t.Errorf("expected ErrInteractionNotFound but didn't get it")
		}
	}
}
