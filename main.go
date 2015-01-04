package main

import (
	"bytes"
	"encoding/json"
	"flag"
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
	Version           string
	Debug             bool
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

var verFlag bool

func init() {
	log.SetFlags(0)

	Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))

	GithubToken = os.Getenv("GITHUB_TOKEN")
	GithubUser = os.Getenv("GITHUB_USER")
	GithubRepo = os.Getenv("GITHUB_REPO")
	GithubAPIEndpoint = os.Getenv("GITHUB_API")

	flag.BoolVar(&verFlag, "version", false, "-version")
	flag.Parse()
}

var usage string = `Github command line release tool.
Usage:
	gh-release <user/repo> <tag> <branch> <description> <files>

	<files> can be specified using glob patterns.

Options:
	-version: Displays version
`

func main() {
	if verFlag {
		log.Println(Version)
		return
	}

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

	if Debug {
		log.Println("Glob pattern received: ")
		log.Println(os.Args[5])
	}

	filepaths, err := filepath.Glob(os.Args[5])
	if err != nil {
		log.Fatalf("Error: Invalid glob pattern: %s\n", os.Args[5])
	}

	if Debug {
		log.Println("Expanded glob pattern: ")
		log.Printf("%v\n", filepaths)
	}

	tag := os.Args[2]
	branch := os.Args[3]
	desc := os.Args[4]

	CreateRelease(tag, branch, desc, filepaths)
	log.Println("Done")
}

func uploadFile(uploadURL, path string) {
	file, err := os.Open(path)
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

	log.Printf("Uploading %s...\n", file.Name())
	body, err := doRequest("POST", uploadURL+"?name="+file.Name(), "application/octet-stream", file, size)
	if err != nil {
		log.Printf("Error: %s\n", err.Error())
	}

	if Debug {
		log.Println("========= UPLOAD RESPONSE ===========")
		log.Println(string(body[:]))
	}
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
			uploadFile(uploadURL, filepaths[index])
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

	if Debug {
		log.Println("================ REQUEST DUMP ==================")
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			log.Println(err.Error())
		}
		log.Println(string(dump[:]))
	}

	resp, err := http.DefaultClient.Do(req)

	if Debug {
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
