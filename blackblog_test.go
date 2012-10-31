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
	"sort"
	"testing"
)

func TestGetPostsInDir(t *testing.T) {
	posts, err := GetPostsInDirectory("./tests")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(posts) != 4 {
		t.Errorf("Expecting %d posts, only got %d", 4, len(posts))
	}

	for _, post := range posts {
		if post.Title == "" || post.Filename == "" {
			t.Errorf("Missing title or filename in post %v", post)
		}
	}
}

func TestSortedPosts(t *testing.T) {
	posts := PostList{
		&Post{URLFragment: "alpha"},
		&Post{URLFragment: "b_test", Date: "6 January 2012"},
		&Post{URLFragment: "c_test", Date: "18 January 2012"},
		&Post{URLFragment: "test", Date: "7 February 2011"},
	}
	order := []string{
		"2011/2/test.html",
		"2012/1/b_test.html",
		"2012/1/c_test.html",
		"alpha.html",
	}

	sort.Sort(posts)

	for i, expected := range order {
		if expected != posts[i].CreateURL() {
			t.Errorf("Sorted order mismatch. At %d, expected '%s', got '%s'", i, expected, posts[i].CreateURL())
		}
	}
}
