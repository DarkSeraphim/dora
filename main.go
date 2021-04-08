package main

/*

Copyright (c) 2021 Jeroen van Dongen

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
documentation files (the "Software"), to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO
THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
IN THE SOFTWARE.

*/

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/abbot/go-http-auth"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/kelseyhightower/envconfig"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var Version string

// FileSystem custom file system handler
type FileSystem struct {
	fs http.FileSystem
}

type RepoHandler struct {
	RepoURL string
	HookSecret string
	DeployKey []byte
	DeployKeyPass string
	Branch string
	CloneDir string
}
// Open opens file
func (fs FileSystem) Open(path string) (http.File, error) {
	f, err := fs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil { return nil, err }
	if s.IsDir() {
		index := strings.TrimSuffix(path, "/") + "/index.html"
		if _, err := fs.fs.Open(index); err != nil {
			return nil, err
		}
	}

	return f, nil
}

func (rh *RepoHandler) Clone() error {
	sshKeys, err := ssh.NewPublicKeys("git", rh.DeployKey, rh.DeployKeyPass)
	if err != nil { return err }

	r, err := git.PlainClone(rh.CloneDir, false, &git.CloneOptions{
		// The intended use of a GitHub personal access token is in replace of your password
		// because access tokens can easily be revoked.
		// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
		Auth:     sshKeys,
		URL:      rh.RepoURL,
		Progress: os.Stdout,
	})
	if err != nil { return err }
	w, err := r.Worktree()
	if err != nil { return err }
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(rh.Branch),
	})
	if err != nil { return err }
	return nil
}

func (rh *RepoHandler) Pull() error {
	r, err := git.PlainOpen(rh.CloneDir)
	if err != nil { return err }
	w, err := r.Worktree()
	if err != nil { return err }

	sshKeys, err := ssh.NewPublicKeys("git", rh.DeployKey, rh.DeployKeyPass)
	if err != nil { return err }

	err = w.Pull(&git.PullOptions{
		Auth:     sshKeys,
		RemoteName: "origin",
	})
	if err != nil { return err }
	return nil
}

func (rh *RepoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//if r.Method == "GET" {
	//	err := rh.Pull()
	//	if err != nil {
	//		if err != git.NoErrAlreadyUpToDate {
	//			http.Error(w, err.Error(), http.StatusInternalServerError)
	//			return
	//		}
	//	}
	//	http.Redirect(w, r, "/", http.StatusFound)
	//	return
	//}

	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusUnsupportedMediaType)
		return
	}
	if r.Header.Get("X-GitHub-Event") != "push" {
		return
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if rh.HookSecret != "" {
		log.Println("Checking hook secret")
		mac := hmac.New(sha256.New, []byte(rh.HookSecret))
		mac.Write(body)
		expectedMAC := mac.Sum(nil)
		hexMac := strings.TrimPrefix(r.Header.Get("X-Hub-Signature-256"), "sha256=")
		messageMAC, err := hex.DecodeString(hexMac)
		if err != nil {
			log.Printf("error decoding message mac: %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if !hmac.Equal(messageMAC, expectedMAC) {
			log.Printf("invalid mac\n")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
	}

	err = rh.Pull()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func main() {
	var config struct {
		RepoURL       string `required:"true" split_words:"true"`
		Branch        string `default:"main"`
		DeployKey     string `required:"true" split_words:"true"`
		DeployKeyPass string `split_words:"true"`
		HookEndpoint string `default:"/__pull" split_words:"true"`
		HookSecret string `split_words:"true"`
		DocumentRoot  string `default:"public" split_words:"true"`
		MountPoint string `default:"/" split_words:"true"`
		CloneDir      string `default:"/var/www" split_words:"true"`
		Port          int    `default:"8080"`
		BasicAuth 	map[string]string

	}
	envconfig.MustProcess("", &config)

	log.Printf("Dora %s starting ...\n", Version)

	log.Printf("Preparing repo %s:%s\n", config.RepoURL, config.Branch)
	deployKey, err := base64.StdEncoding.DecodeString(config.DeployKey)
	if err != nil {
		panic(fmt.Errorf("error decoding deploy key: %w", err))
	}
	repoHandler := RepoHandler{
		RepoURL: config.RepoURL,
		DeployKeyPass: config.DeployKeyPass,
		DeployKey: deployKey,
		Branch: config.Branch,
		CloneDir: config.CloneDir,
		HookSecret: config.HookSecret,
	}
	err = repoHandler.Clone()
	if err != nil {
		panic(fmt.Errorf("error cloning repo: %w", err))
	}


	fileServer := http.FileServer(FileSystem{http.Dir(filepath.Join(config.CloneDir, config.DocumentRoot))})

	if len(config.BasicAuth) > 0 {
		log.Println("Enabling basic authentication")

		htpasswd := map[string]string{}
		for user, pass := range config.BasicAuth {
			htpasswd[user] = string(auth.MD5Crypt([]byte(pass), []byte(auth.RandomKey()), []byte("$1$")))
			log.Printf("Adding user %s\n", user)
		}
		authenticator := auth.NewBasicAuthenticator("authentication required", func(user, realm string) string {
			passwd, ok := htpasswd[user]
			if !ok { return "" }
			log.Printf("authenticated request for user %s\n", user)
			return passwd
		})
		fileServer = auth.JustCheck(authenticator, fileServer.ServeHTTP)
	}

	http.Handle(config.HookEndpoint, &repoHandler)
	http.Handle(config.MountPoint, fileServer)

	log.Printf("Serving %s on HTTP port: %s\n", config.DocumentRoot, config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.Port), nil))
}
