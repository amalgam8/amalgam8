// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	userCookie          = "user"
	requestIDHeader     = "X-Request-ID"
	enableRatingsEnvVar = "ENABLE_RATINGS"
	starColorEnvVar     = "STAR_COLOR"
	defaultStarColor    = "black"
)

// Globals
var (
	proxyURL      string
	enableRatings bool
	starColor     string
)

type Review struct {
	Text   string  `json:"text,omitempty"`
	Rating *Rating `json:"rating,omitempty"`
}

type Rating struct {
	Stars int    `json:"stars,omitempty"`
	Color string `json:"color,omitempty"`
}

var Reviews = map[string]*Review{
	"reviewer1": {
		Text: "An extremely entertaining play by Shakespeare. The slapstick humour is refreshing!",
	},
	"reviewer2": {
		Text: "Absolutely fun and entertaining. The play lacks thematic depth when compared to other plays by Shakespeare.",
	},
}

func main() {
	if len(os.Args) < 3 {
		log.Printf("Usage: %s <port> <proxy url>", os.Args[0])
		os.Exit(-1)
	}

	port := os.Args[1]
	proxyURL = os.Args[2]

	enableRatings = os.Getenv(enableRatingsEnvVar) == "true"
	starColor = os.Getenv(starColorEnvVar)
	if starColor == "" {
		starColor = defaultStarColor
	}

	http.HandleFunc("/reviews", reviewsHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func reviewsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var ratings map[string]*Rating
	if enableRatings {
		ratings = getRatings(getForwardHeaders(r))
	} else {
		ratings = map[string]*Rating{}
	}

	reviews := make(map[string]*Review, len(Reviews))
	for k, v := range Reviews {
		reviews[k] = &Review{
			Text:   v.Text,
			Rating: ratings[k],
		}
	}

	bytes, _ := json.Marshal(reviews)
	w.Write(bytes)
}

func getRatings(forwardHeaders http.Header) map[string]*Rating {
	timeout := 2500 * time.Millisecond
	if starColor == defaultStarColor {
		timeout = 10 * time.Second
	}

	ratings := map[string]*Rating{}

	bytes, err := doRequest("/ratings/ratings", forwardHeaders, timeout)
	if err != nil {
		log.Printf("Error getting ratings: %v", err)
		return ratings
	}
	json.Unmarshal(bytes, &ratings)

	for _, v := range ratings {
		v.Color = starColor
	}

	return ratings
}

func doRequest(path string, forwardHeaders http.Header, timeout time.Duration) ([]byte, error) {
	client := http.Client{}
	client.Timeout = timeout

	req, _ := http.NewRequest("GET", proxyURL+path, nil)
	req.Header = forwardHeaders

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received unexpected status code %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

func getForwardHeaders(r *http.Request) http.Header {
	fwdReq, _ := http.NewRequest("GET", "dummy", nil)

	cookie, err := r.Cookie(userCookie)
	if err != http.ErrNoCookie {
		fwdReq.AddCookie(cookie)
	}

	reqID := r.Header.Get(requestIDHeader)
	if reqID != "" {
		fwdReq.Header.Set(requestIDHeader, reqID)
	}

	return fwdReq.Header
}
