//
// Blackblog
// Copyright 2012 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"testing"
)

func TestGetContents(t *testing.T) {
	post, _ := NewPostFromPath("./tests/simple_post.md")
	s, _ := post.GetContents()
	expected := "\nThis is a simple post!\n\nWith two lines.\n"
	if string(s) != expected {
		t.Errorf("simple_post contents incorrect. Expected '%s', got '%s'", expected, s)
	}
}

type parseMetadata struct {
	input string
	post Post
}

func TestParseMetadataLine(t *testing.T) {
	var results = []parseMetadata{
		{"~~ Title: This is a title", Post{Title: "This is a title"}},
		{"~~ Title: ~Weird Data~", Post{Title: "~Weird Data~"}},
		{"~~ Unknown: Field", Post{}},
		{"~~ uRl: foo_bar.html", Post{URLFragment: "foo_bar.html"}},
		{"~~ Date: 12/13/1344", Post{Date: "12/13/1344"}},
		{"~~Date: 13 January 2012     ", Post{Date: "13 January 2012"}},
	}

	for _, r := range results {
		var p Post
		p.parseMetadataLine(r.input)
		rp := r.post
		if p.Title != rp.Title || p.Date != rp.Date || p.URLFragment != rp.URLFragment {
			t.Errorf("Parse error for input '%s', expected '%v', got '%v'", r.input, r.post, p)
		}
	}
}

func TestFullMetadata(t *testing.T) {
	post, err := NewPostFromPath("./tests/simple_post.md")
	if err != nil {
		t.Errorf("Error reading post: %v", err)
	}

	expected := "Simple Post"
	if post.Title != expected {
		t.Errorf("post.Title mismatch, expected '%s', got '%s'", expected, post.Title)
	}

	expected = "simple_post"
	if post.URLFragment != expected {
		t.Errorf("post.URLFragment mismatch, expected '%s', got '%s'", expected, post.URLFragment)
	}

	expected = "24 Jan 2012"
	if post.Date != expected {
		t.Errorf("post.Date mismatch, expected '%s', got '%s'", expected, post.Date)
	}
}

func TestIsOutOfDate(t *testing.T) {
	post, err := NewPostFromPath("./tests/update_test.md")
	if err != nil {
		t.Errorf("Error reading post: %v", err)
	}

	if !post.IsUpToDate() {
		t.Errorf("Post %s is unexpectedly out of date", post.Filename)
	}

	post.Filename = "./tests/update_test_out_of_date.md"
	if post.IsUpToDate() {
		t.Errorf("Post %s is unexpectedly up-to-date", post.Filename)
	}
}
