package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"
)

var (
	downloadedFiles []string
)

const (
	ItunesEndPoint = "https://itunes.apple.com/search/"
)

type ITunesRequestParams struct {
	Term    string
	Country string
	Lang    string
	Media   string
	Limit   string
}

type Music struct {
	ArtistName     string `json:"artistName"`
	TrackName      string `json:"trackName"`
	PreviewURL     string `json:"previewUrl"`
	CollectionName string `json:"CollectionName"`
	TrackViewURL   string `json:"TrackViewURL"`
}

type ITunesResponse struct {
	Count  int     `json:"resultCount"`
	Musics []Music `json:"results"`
}

func SearchMusic(v ITunesRequestParams) <-chan ITunesResponse {

	resultChan := make(chan ITunesResponse)

	params := url.Values{}
	params.Add("term", v.Term)
	params.Add("country", v.Country)
	params.Add("lang", v.Lang)
	params.Add("media", v.Media)
	params.Add("limit", v.Limit)

	go func() {
		fmt.Println("* provided courtesy of iTunes *")
		fmt.Println()
		response, err := http.Get(ItunesEndPoint + "?" + params.Encode())

		if err != nil {
			log.Fatal(err)
			os.Exit(1)

		} else {
			defer response.Body.Close()

			contents, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}

			var data ITunesResponse
			json.Unmarshal([]byte(contents), &data)

			resultChan <- data
		}

	}()

	return resultChan
}
func Play(fileName string, rate string) {
	defer os.Remove(fileName)

	out, _ := exec.Command("afplay", fileName, "--rate", rate, "-q", "1").CombinedOutput()
	fmt.Print(string(out))
}

func Info(music Music, num int, total int) {
	fmt.Printf("* (%d/%d)\n", num, total)
	fmt.Printf("♪ %s - %s / %s\n", music.TrackName, music.ArtistName, music.CollectionName)
	fmt.Printf("# %s\n", music.TrackViewURL)
	fmt.Println()
}

func Download(url string) <-chan string {

	fileNameChan := make(chan string)

	go func(url string) {
		response, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		defer response.Body.Close()

		file, err := ioutil.TempFile(os.TempDir(), "bgm_tmp_")
		downloadedFiles = append(downloadedFiles, file.Name())

		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		io.Copy(file, response.Body)

		fileNameChan <- file.Name()

	}(url)

	return fileNameChan

}

func ShuffleMusic(a *[]Music) {

	musics := *a

	rand.Seed(time.Now().UnixNano())
	for i := range musics {
		j := rand.Intn(i + 1)
		musics[i], musics[j] = musics[j], musics[i]
	}
}

func SweepFiles() {
	for _, file := range downloadedFiles {
		os.Remove(file)
	}
}
