package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Fatal errors handler
func HandleFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func GetRawData(link string) string {
	u, err := url.Parse(link)
	HandleFatal(err)
	video_id := u.Query()["v"][0]
	v := url.Values{}
	v.Set("video_id", video_id)
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://www.youtube.com/get_video_info?%s", v.Encode()), nil)
	HandleFatal(err)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/34.0.1847.116 Chrome/34.0.1847.116 Safari/537.36")
	resp, err := client.Do(req)
	HandleFatal(err)
	defer resp.Body.Close()
	raw_data, err := ioutil.ReadAll(resp.Body)
	HandleFatal(err)

	return string(raw_data)
}

func GetFileMeta(link string) *http.Response {
	// TODO: Change request headers
	client := &http.Client{}
	resp, err := client.Head(link)
	HandleFatal(err)
	defer resp.Body.Close()

	return resp
}

// Function downloads and saves video file
func Download(link string, filename string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", link, nil)
	HandleFatal(err)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/34.0.1847.116 Chrome/34.0.1847.116 Safari/537.36")
	resp, err := client.Do(req)
	HandleFatal(err)
	defer resp.Body.Close()
	// Creating output file
	out, err := os.Create(filename)
	HandleFatal(err)
	fmt.Printf("Downloading \"%s\"...", filename)
	n, err := io.Copy(out, resp.Body)
	HandleFatal(err)
	fmt.Printf("Done. %d bytes copied\n", n)

}

func main() {
	var itag_argument int
	flag.IntVar(&itag_argument, "itag", 0, `itag values: 
	itag=5 - FLV 320 x 240
	itag=34 - FLV 640 x 360
	itag=35 - FLV 854 x 480
	itag=18 - MP4 640 x 360
	itag=22 - MP4 1280 x 720
	itag=37 - MP4 1920 x 1080
	itag=38 - MP4 4096 x 1714
	itag=43 - WEBM 640 x 360
	itag=44 - WEBM 854 x 480
	itag=45 - WEBM 1280 x 720`)

	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("YouTube address missing...")
	}
	youtubeURL := flag.Arg(0)

	raw_data := GetRawData(youtubeURL)
	var (
		data  string
		title string
	)
	lst := strings.Split(raw_data, "&")
	for _, elem := range lst {
		if strings.HasPrefix(elem, "title=") {
			title = strings.TrimPrefix(elem, "title=")
		}
		if strings.Contains(elem, "url_encoded_fmt_stream_map") {
			data = elem
		}
	}

	q, err := url.ParseQuery(data)
	HandleFatal(err)
	encoded_stream_map := q["url_encoded_fmt_stream_map"][0]

	urls := []string{}
	x := strings.Split(encoded_stream_map, "&")
	for _, val := range x {
		if strings.HasPrefix(val, "url=") {
			urls = append(urls, strings.TrimPrefix(val, "url="))
		}
	}

	fmt.Printf("Title: %s\n\n", title)

	var (
		s           string
		directLink  *url.URL
		videoFormat string
		ext         string
	)

	if itag_argument == 0 {
		fmt.Println("itag argument missing. Only listing available formats.")
	}

	for _, u := range urls {
		s, _ = url.QueryUnescape(u)
		directLink, _ = url.Parse(s)
		// Debugging
		// fmt.Printf("%+v\n", directLink.String())
		itag := directLink.Query()["itag"][0]
		switch itag {
		case "5":
			videoFormat = "FLV 320 x 240"
			ext = ".flv"
		case "17":
			videoFormat = "3GP video"
			ext = ".3gp"
		case "18":
			videoFormat = "MP4 640 x 360"
			ext = ".mp4"
		case "22":
			videoFormat = "MP4 1280 x 720"
			ext = ".mp4"
		case "34":
			videoFormat = "FLV 640 x 360"
			ext = ".flv"
		case "35":
			videoFormat = "FLV 854 x 480"
			ext = ".flv"
		case "37":
			videoFormat = "MP4 1920 x 1080"
			ext = ".mp4"
		case "38":
			videoFormat = "MP4 4096 x 1714"
			ext = ".mp4"
		case "43":
			videoFormat = "WEBM 640 x 360"
			ext = ".webm"
		case "44":
			videoFormat = "WEBM 854 x 480"
			ext = ".webm"
		case "45":
			videoFormat = "WEBM 1280 x 720"
			ext = ".webm"
		default:
			videoFormat = "Unknown quality"
			ext = ""
		}

		directLink.Query().Set("title", title)
		metaData := *GetFileMeta(directLink.String())

		if itag_argument == 0 || itag == strconv.Itoa(itag_argument) {
			fmt.Printf("itag = %s (%s)\n", itag, videoFormat)
			fmt.Printf("File size: %v bytes \n\n", metaData.Header.Get("Content-Length"))
		}

		if itag == strconv.Itoa(itag_argument) {
			filename, err := url.QueryUnescape(title)
			HandleFatal(err)
			Download(directLink.String(), filename+ext)
		}

	}

}
