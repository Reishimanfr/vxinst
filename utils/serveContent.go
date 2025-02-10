package utils

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type PostData struct {
	Author   string
	HasVideo bool
	VideoURL string
	HasImage bool
	ImageURL string
	Error    error
}

var (
	client = http.Client{
		Transport: &http.Transport{},
		Timeout:   5 * time.Second,
	}
)

// Searches through the GraphQL data in search of useful data
func ServeContent(postId string) *PostData {
	if postId == "" {
		return &PostData{
			Error: errors.New("no post ID provided"),
		}
	}

	gqlParams := url.Values{
		"av":                       {"0"},
		"__d":                      {"www"},
		"__user":                   {"0"},
		"__a":                      {"1"},
		"__req":                    {"k"},
		"__hs":                     {"19888.HYP:instagram_web_pkg.2.1..0.0"},
		"dpr":                      {"2"},
		"__ccg":                    {"UNKNOWN"},
		"__rev":                    {"1014227545"},
		"__s":                      {"trbjos:n8dn55:yev1rm"},
		"__hsi":                    {"7380500578385702299"},
		"__dyn":                    {"7xeUjG1mxu1syUbFp40NonwgU7SbzEdF8aUco2qwJw5ux609vCwjE1xoswaq0yE6ucw5Mx62G5UswoEcE7O2l0Fwqo31w9a9wtUd8-U2zxe2GewGw9a362W2K0zK5o4q3y1Sx-0iS2Sq2-azo7u3C2u2J0bS1LwTwKG1pg2fwxyo6O1FwlEcUed6goK2O4UrAwCAxW6Uf9EObzVU8U"},
		"__csr":                    {"n2Yfg_5hcQAG5mPtfEzil8Wn-DpKGBXhdczlAhrK8uHBAGuKCJeCieLDyExenh68aQAKta8p8ShogKkF5yaUBqCpF9XHmmhoBXyBKbQp0HCwDjqoOepV8Tzk8xeXqAGFTVoCciGaCgvGUtVU-u5Vp801nrEkO0rC58xw41g0VW07ISyie2W1v7F0CwYwwwvEkw8K5cM0VC1dwdi0hCbc094w6MU1xE02lzw"},
		"__comet_req":              {"7"},
		"lsd":                      {"AVoPBTXMX0Y"},
		"jazoest":                  {"2882"},
		"__spin_r":                 {"1014227545"},
		"__spin_b":                 {"trunk"},
		"__spin_t":                 {"1718406700"},
		"fb_api_caller_class":      {"RelayModern"},
		"fb_api_req_friendly_name": {"PolarisPostActionLoadPostQueryQuery"},
		"variables":                {`{"shortcode":"` + postId + `","fetch_comment_count":40,"parent_comment_count":24,"child_comment_count":3,"fetch_like_count":10,"fetch_tagged_user_count":null,"fetch_preview_comment_count":2,"has_threaded_comments":true,"hoisted_comment_id":null,"hoisted_reply_id":null}`},
		"server_timestamps":        {"true"},
		"doc_id":                   {"25531498899829322"},
	}

	req, err := http.NewRequest("POST", "https://instagram.com/graphql/query", strings.NewReader(gqlParams.Encode()))
	if err != nil {
		return &PostData{
			Error: fmt.Errorf("failed to prepare new request to gql query: %v", err),
		}
	}

	defer req.Body.Close()

	req.Header = http.Header{
		"Accept":                      {"*/*"},
		"Accept-Language":             {"en-US,en;q=0.9"},
		"Content-Type":                {"application/x-www-form-urlencoded"},
		"Origin":                      {"https://www.instagram.com"},
		"Priority":                    {"u=1, i"},
		"Sec-Ch-Prefers-Color-Scheme": {"dark"},
		"Sec-Ch-Ua":                   {`"Google Chrome";v="125", "Chromium";v="125", "Not.A/Brand";v="24"`},
		"Sec-Ch-Ua-Full-Version-List": {`"Google Chrome";v="125.0.6422.142", "Chromium";v="125.0.6422.142", "Not.A/Brand";v="24.0.0.0"`},
		"Sec-Ch-Ua-Mobile":            {"?0"},
		"Sec-Ch-Ua-Model":             {`""`},
		"Sec-Ch-Ua-Platform":          {`"macOS"`},
		"Sec-Ch-Ua-Platform-Version":  {`"12.7.4"`},
		"Sec-Fetch-Dest":              {"empty"},
		"Sec-Fetch-Mode":              {"cors"},
		"Sec-Fetch-Site":              {"same-origin"},
		"User-Agent":                  {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"},
		"X-Asbd-Id":                   {"129477"},
		"X-Bloks-Version-Id":          {"e2004666934296f275a5c6b2c9477b63c80977c7cc0fd4b9867cb37e36092b68"},
		"X-Fb-Friendly-Name":          {"PolarisPostActionLoadPostQueryQuery"},
		"X-Ig-App-Id":                 {"936619743392459"},
	}

	res, err := client.Do(req)
	if err != nil {
		return &PostData{
			Error: fmt.Errorf("gql request failed: %v", err),
		}
	}

	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)
	scanner.Buffer(make([]byte, 16*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		results := ExtractData(line, true)
	}

	return nil
}
