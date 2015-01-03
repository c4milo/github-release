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
	UploadURL  string `json:"upload_url,omitempty"`
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
	gh-release <user/repo> <tag> <branch> <description> <files>

<files> can be specified using glob patterns.
`

func main() {
	if len(os.Args) < 6 {
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

	filepaths, err := filepath.Glob(os.Args[5])
	if err != nil {
		log.Fatalf("Error: Invalid glob pattern: %s\n", os.Args[3])
	}

	tag := os.Args[2]
	branch := os.Args[3]
	desc := os.Args[4]

	CreateRelease(tag, branch, desc, filepaths)
	log.Println("Done")
}

// Creates a Github Release, attaching the given files as release assets
// If a release already exist, up in Github, this function will attempt to attach the given files to it
func CreateRelease(tag, branch, desc string, filepaths []string) {
	endpoint := fmt.Sprintf("%s/releases", GithubAPIEndpoint)

	release := Release{
		TagName:    tag,
		Name:       tag,
		Prerelease: false,
		Draft:      false,
		Branch:     branch,
		Body:       desc,
	}

	releaseData, err := json.Marshal(release)
	if err != nil {
		log.Fatalln(err)
	}

	releaseBuffer := bytes.NewBuffer(releaseData)

	data, err := doRequest("POST", endpoint, "application/json", releaseBuffer, int64(releaseBuffer.Len()))
	if err != nil {
		log.Fatalln(err)
	}

	// Gets the release Upload URL from the returned JSON data
	err = json.Unmarshal(data, &release)
	if err != nil {
		log.Fatalln(err)
	}

	// Upload URL comes like this https://uploads.github.com/repos/octocat/Hello-World/releases/1/assets{?name}
	// So we need to remove the {?name} part
	uploadURL := strings.Split(release.UploadURL, "{")[0]

	var wg sync.WaitGroup
	for i := range filepaths {
		wg.Add(1)
		func(index int) {
			file, err := os.Open(filepaths[i])
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

			body, err := doRequest("POST", uploadURL+"?name="+file.Name(), "application/octet-stream", file, size)
			if err != nil {
				log.Printf("Error: %s\n", err.Error())
			}

			if DEBUG {
				log.Println("========= UPLOAD RESPONSE ===========")
				log.Println(string(body[:]))
			}

			wg.Done()
		}(i)
	}
	wg.Wait()
}

func fileSize(file *os.File) (int64, error) {
	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

// Sends HTTP request to Github API
func doRequest(method, url, contentType string, reqBody io.Reader, bodySize int64) ([]byte, error) {
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Github returned an error:\n Code: %s. \n Body: %s", resp.Status, respBody)
	}

	return respBody, nil
}
