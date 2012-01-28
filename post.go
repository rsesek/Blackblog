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
	"bufio"
	"bytes"
	"crypto/md5"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// A Post contains the metadata about a post.
type Post struct {
	// The full path to the Markdown file containing this post.
	Filename string

	// The title of the post, from the metadata.
	Title string

	// The URL fragment of the post, from the metadata.
	URLFragment string

	// The date the post was published, from the metadata.
	Date string

	// The MD5 checksum of the file's contents.
	checksum []byte
}

// NewPostFromPath creates a new Post object from the file at the given path
// with the metadata updated.
func NewPostFromPath(path string) (*Post, os.Error) {
	p := &Post{Filename: path}
	p.UpdateMetadata()
	if len(p.checksum) < 1 {
		return nil, os.NewError("Could not checksum blog post")
	}
	return p, nil
}

// Open returns an opened file handle for the Post.
func (p *Post) Open() (*os.File, os.Error) {
	return os.Open(p.Filename)
}

// IsUpToDate checks that the in-memory data about the Post matches that of the
// file on disk.
func (p *Post) IsUpToDate() bool {
	file, err := p.Open()
	defer file.Close()
	if err != nil {
		return false
	}
	return p.isUpToDateInternal(file)
}

func (p *Post) isUpToDateInternal(file *os.File) bool {
	fd, err := p.Open()
	defer fd.Close()
	if err != nil {
		return false
	}

	return bytes.Equal(p.checksum, computeChecksum(fd))
}

func computeChecksum(file *os.File) []byte {
	digest := md5.New()
	io.Copy(digest, file)
	return digest.Sum()
}

// GetContents returns the Markdown content of a post, excluding metadata.
func (p *Post) GetContents() ([]byte, os.Error) {
	file, err := p.Open()
	defer file.Close()
	if err != nil {
		return nil, err
	}

	var result []byte
	wasPrefix := false
	reader := bufio.NewReader(file)
	for {
		line, isPrefix, err := reader.ReadLine()

		// If an error occurred, return the error except for EOF.
		if err != nil {
			if err == os.EOF {
				err = nil
			}
			return result, err
		}

		// Skip lines that are for metadata.
		if wasPrefix || len(line) >= 2 && string(line[0:2]) == "~~" {
			wasPrefix = isPrefix
			continue
		}

		if !isPrefix {
			line = append(line, '\n')
		}

		// Store the rest of the file.
		result = append(result, line...)
	}
	return result, nil
}

// UpdateMetadata re-reads the file on disk and updates the in-memory metadata.
func (p *Post) UpdateMetadata() {
	file, err := p.Open()
	defer file.Close()
	if err != nil || p.isUpToDateInternal(file) {
		return
	}
	p.checksum = computeChecksum(file)

	file.Seek(0, 0)
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err == os.EOF || len(line) < 2 || line[0:2] != "~~" {
			break
		}
		p.parseMetadataLine(line)
	}
}

func (p *Post) parseMetadataLine(line string) {
	if len(line) < 2 || line[0:2] != "~~" {
		return
	}
	line = line[2:]

	pieces := strings.SplitN(line, ":", 2)
	if len(pieces) != 2 {
		return
	}

	val := strings.TrimSpace(pieces[1])

	switch strings.ToLower(strings.TrimSpace(pieces[0])) {
	case "title": p.Title = val
	case "url": p.URLFragment = val
	case "date": p.Date = val
	}
}

// CreateURL constructs the URL of a post based on its metadata.
func (p* Post) CreateURL() string {
	// First, create the file's basename.
	basename := p.URLFragment
	if basename == "" && p.Title != "" {
		basename = strings.Replace(p.Title, " ", "_", -1)
	} else if basename == "" {
		basename = path.Base(p.Filename)
		ext := path.Ext(basename)
		basename = basename[:len(basename) - len(ext)]
	}

	basename = strings.ToLower(basename)
	url := basename + ".html"

	// Next, try and get the date of the post to include subdirectories.
	time := parseDate(p.Date)
	if time != nil {
		url = path.Join(strconv.Itoa64(time.Year), strconv.Itoa(time.Month), url)
	}

	return url
}

func parseDate(input string) *time.Time {
	if input == "" {
		return nil
	}

	t, err := time.Parse("_2 January 2006", input)
	if err == nil {
		return t
	}

	t, err = time.Parse("January _2, 2006", input)
	if err == nil {
		return t
	}

	t, err = time.Parse("January _2 2006", input)
	if err == nil {
		return t
	}

	return nil
}