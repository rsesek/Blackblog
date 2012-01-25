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
	"bytes"
	"fmt"
	"flag"
	"os"
	"path"
	"sort"
	"strings"
	"template"

	"github.com/russross/blackfriday"
)

var (
	flagSource = flag.String("root", "", "The root directory of all Markdown posts")
	flagPort = flag.String("port", "", "The port to bind to for running the standalone HTTP server")
	flagDest = flag.String("dest", "", "The output directory for running in comiple mode")
)

func main() {
	flag.Parse()

	if *flagSource == "" {
		fmt.Fprintf(os.Stderr, "No -root blog directory specified\n")
		os.Exit(1)
	}

	if *flagPort == "" && *flagDest == "" {
		fmt.Fprintf(os.Stderr, "No -port or -dest flag specified\n")
		os.Exit(2)
	}

	if *flagPort != "" {
		fmt.Fprintf(os.Stderr, "** SERVER NOT IMPLEMENTED **\n")
		os.Exit(-1)
	}

	posts := GetPostsInDirectory(*flagSource)
	postMap, sortList := SortPosts(posts)
	for _, url := range sortList {
		post := postMap[url]
		filePath := path.Join(*flagDest, url)

		source, err := post.GetContents()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading post at %s: %v\n", post.Filename, err)
			continue
		}

		html := RenderPost(post, source)
		makeParentDirIfNecessary(filePath)

		fd, err := os.Create(filePath)
		defer fd.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			continue
		}
		fd.Write(html)
	}
}

// GetPostsInDirectory recursively examines the directory at the path and finds
// any Markdown (.md) files and returns the corresponding Post objects.
func GetPostsInDirectory(dirPath string) []*Post {
	fd, err := os.Open(dirPath)
	defer fd.Close()
	if err != nil {
		return nil
	}

	files, err := fd.Readdir(-1)
	if err != nil {
		return nil
	}

	var results []*Post
	for _, file := range files {
		filePath := path.Join(dirPath, file.Name)
		if file.IsDirectory() {
			subfiles := GetPostsInDirectory(filePath)
			if subfiles != nil {
				results = append(results, subfiles...)
			}
		} else if strings.HasSuffix(file.Name, ".md") {
			if post, err := NewPostFromPath(filePath); post != nil && err == nil {
				results = append(results, post)
			}
		}
	}

	return results
}

// PostURLMap keys Post objects by their final URL.
type PostURLMap map[string]*Post

// SortPosts creates URLs for each Post and returns a map that links the URL to
// the post and a slice of the URLs in sorted order.
func SortPosts(posts []*Post) (postMap PostURLMap, sorted []string) {
	postMap = make(PostURLMap, len(posts))
	for _, post := range posts {
		url := post.CreateURL()
		postMap[url] = post
		sorted = append(sorted, url)
	}
	sort.Strings(sorted)
	return
}

var kPostHeader = []string{
	"<div id=\"header\">",
	"	<h1 id=\"title\">{Title}</h1>",
	"	<h2 id=\"date\">{Date}</h2>",
	"</div>",
	"\n",
}

// RenderPost runs the input source through the blackfriday library.
func RenderPost(post *Post, input []byte) []byte {
	tpl := template.New(nil)
	tpl.Parse(strings.Join(kPostHeader, "\n"))

	buf := bytes.NewBuffer([]byte{})
	tpl.Execute(buf, post)

	buf.Write(blackfriday.Markdown(
		input,
		blackfriday.HtmlRenderer(
			blackfriday.HTML_USE_SMARTYPANTS |
				blackfriday.HTML_SMARTYPANTS_LATEX_DASHES,
			"",
			""),
		0))

	return buf.Bytes()
}

func makeParentDirIfNecessary(dir string) {
	parent, _ := path.Split(dir)
	os.MkdirAll(parent, 0755)
}
