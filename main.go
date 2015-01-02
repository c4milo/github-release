package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var (
	GithubToken       string
	GithubUser        string
	GithubRepo        string
	GithubAPIEndpoint string
	DEBUG             bool
)

type Release struct {
	TagName    string `json:"tag_name"`
	Branch     string `json:"target_commitish"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

func init() {
	log.SetFlags(0)

	DEBUG, _ = strconv.ParseBool(os.Getenv("DEBUG"))

	GithubToken = os.Getenv("GITHUB_TOKEN")
	GithubUser = os.Getenv("GITHUB_USER")
	GithubRepo = os.Getenv("GITHUB_REPO")
	GithubAPIEndpoint = os.Getenv("GITHUB_API")
}

var usage string = `Github release tool.
Usage:
	gh-release <user/repo> <tag> <files>

<files> can be specified using glob patterns.
`

func main() {
	if len(os.Args) < 4 {
		log.Fatal(usage)
	}

	userRepo := strings.Split(os.Args[1], "/")
	if len(userRepo) != 2 {
		log.Printf("Error: Invalid format used for username and repository: %s\n\n", os.Args[1])
		log.Fatal(usage)
	}

	if GithubToken == "" {
		log.Fatal(`Error: GITHUB_TOKEN environment variable is not set.
																																																Please refer to https://help.github.com/articles/creating-an-access-token-for-command-line-use/ for more help\n`)
	}

	GithubUser = userRepo[0]
	GithubRepo = userRepo[1]
	GithubAPIEndpoint = fmt.Sprintf("https://api.github.com/repos/%s/%s", GithubUser, GithubRepo)

	filepaths, err := filepath.Glob(os.Args[3])
	if err != nil {
		log.Fatalf("Error: Invalid glob pattern: %s\n", os.Args[3])
	}

	tag := os.Args[2]

	CreateRelease(tag, filepaths)
	log.Println("Done")
}

// Creates a Github Release, attaching the given files as release assets
// If a release already exist up in Github, this function will attempt to attach the given files
func CreateRelease(tag string, filepaths []string) {
	endpoint := fmt.Sprintf("%s/releases", GithubAPIEndpoint)

	release := Release{
		TagName:    tag,
		Name:       "Release " + tag,
		Prerelease: false,
		Draft:      false,
		Branch:     "master",
		Body:       "Changelog (TODO)",
	}

	releaseData, err := json.Marshal(release)
	if err != nil {
		log.Fatalln(err)
	}

	releaseBuffer := bytes.NewBuffer(releaseData)

	err = doRequest("POST", endpoint, "application/json", releaseBuffer, int64(releaseBuffer.Len()))
	if err != nil {
		log.Fatalln(err)
	}

	var wg sync.WaitGroup
	for i := range filepaths {
		wg.Add(1)
		func(index int) {
			file, err := os.Open(filepaths[i]) // For read access.
			if err != nil {
				log.Printf("Error: %s\n", err.Error())
				return
			}
			defer file.Close()

			size, err := fileSize(file)
			if err != nil {
				log.Printf("Error: %s\n", err.Error())
				return
			}

			err = doRequest("POST", endpoint, "application/octet-stream", file, size)
			if err != nil {
				log.Printf("Error: %s\n", err.Error())
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

// Calculates md5, sha1 and sha512 hashes for each given file
func HashFiles(files []string) (map[string]string, error) {
	return nil, nil
}

// Generates release changelog by comparing the tag provided, against the latest tag pushed up to Github
func changelog(tag string) string {
	endpoint := fmt.Sprintf("%s/commits", GithubAPIEndpoint)
	log.Println(endpoint)
	return ""
}

func fileSize(file *os.File) (int64, error) {
	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

// Sends HTTP request to Github API
func doRequest(method, url, contentType string, body io.Reader, bodySize int64) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", GithubToken))
	req.Header.Set("Content-type", contentType)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.ContentLength = bodySize

	if DEBUG {
		log.Println("================ REQUEST DUMP ==================")
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			log.Println(err.Error())
		}
		log.Println(string(dump[:]))
	}

	resp, err := http.DefaultClient.Do(req)

	if DEBUG {
		log.Println("================ RESPONSE DUMP ==================")
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Println(err.Error())
		}
		log.Println(string(dump[:]))
	}

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("Github returned an error:\n Code: %s. \n Body: %s", resp.Status, body)
	}

	return nil
}
