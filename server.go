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
	"flag"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	serverPollWait = flag.Int("server-poll-time", 30, "The time in seconds that the server waits before polling the directory for changes.")
)

type blogServer struct {
	blog *Blog

	mu    *sync.RWMutex
	posts PostList
	r     *render
}

// StartBlogServer runs the program's web server given the blog located
// at |blogRoot|.
func StartBlogServer(blog *Blog) error {
	server := &blogServer{
		blog: blog,
		mu:   new(sync.RWMutex),
	}

	err := server.buildPosts()
	if err != nil {
		return err
	}
	go server.pollPostChanges()

	if blog.StaticFilesDir() != "" {
		http.Handle(StaticFilesDir, http.StripPrefix(StaticFilesDir, http.FileServer(http.Dir(blog.StaticFilesDir()))))
	}

	http.Handle("/", server)

	fmt.Printf("Starting blog server on port %d\n", blog.Port())
	return http.ListenAndServe(fmt.Sprintf(":%d", blog.Port()), nil)
}

func (b *blogServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	url := strings.Trim(req.URL.Path, "/")

	b.mu.RLock()
	defer b.mu.RUnlock()

	if url == "" {
		b.serveNode(rw, req, b.r)
		return
	}

	parts := strings.Split(url, "/")
	node := b.r
	for _, part := range parts {
		if child, ok := node.object.(renderTree)[part]; ok {
			node = child
		} else {
			http.NotFound(rw, req)
			return
		}
	}

	b.serveNode(rw, req, node)
}

func (b *blogServer) serveNode(rw http.ResponseWriter, req *http.Request, render *render) {
	switch render.t {
	case renderTypePost:
		post := render.object.(*Post)
		data, err := post.GetContents()
		var content []byte
		if err == nil {
			content, err = RenderPost(post, data, CreatePageParams(b.blog, render))
		}
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			fmt.Fprint(rw, err.Error())
			return
		}
		rw.Write(content)
	case renderTypeRedirect:
		http.Redirect(rw, req, render.object.(string), http.StatusMovedPermanently)
	case renderTypeDirectory:
		// The root element should generate a post list.
		if render.parent == nil {
			index, err := CreateIndex(b.posts, b.blog)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(rw, err.Error())
				return
			}
			rw.Write(index)
		} else {
			// Otherwise, render the index.html node.
			render = render.object.(renderTree)["index.html"]
			b.serveNode(rw, req, render)
		}
	default:
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, "Unknown render: %v", render)
	}
}

func (b *blogServer) pollPostChanges() {
	for {
		time.Sleep(time.Duration(*serverPollWait) * time.Second)
		if err := b.buildPosts(); err != nil {
			panic(err.Error())
		}
	}
}

func (b *blogServer) buildPosts() (err error) {
	newPosts, err := GetPostsInDirectory(b.blog.GetPostsDir())
	if err != nil {
		return
	}

	b.mu.RLock()
	rebuild := len(newPosts) != len(b.posts)
	if !rebuild {
		for _, p := range b.posts {
			if !p.IsUpToDate() {
				rebuild = true
				break
			}
		}
	}
	b.mu.RUnlock()

	if rebuild {
		b.mu.Lock()
		defer b.mu.Unlock()

		b.posts = newPosts
		b.r, err = createRenderTree(b.posts)
		if err != nil {
			return
		}
	}
	return nil
}
