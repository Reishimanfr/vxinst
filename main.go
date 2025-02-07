package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
	"unsafe"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"github.com/lmittmann/tint"
)

var (
	client = &http.Client{
		Timeout: 10 * time.Second,
	}

	transport = &http.Transport{
		MaxIdleConnsPerHost: 10,
		DisableCompression:  true,
		IdleConnTimeout:     90 * time.Second,
	}

	httpClient = &http.Client{
		Transport: transport,
	}

	store = persist.NewMemoryStore(time.Minute * 1)

	port     = flag.String("port", "8080", "Port to run the server on")
	dev      = flag.Bool("dev", false, "Enable debugging")
	secure   = flag.Bool("secure", false, "Use a secure connection")
	logLevel = flag.String("log-level", "info", "Logging verbositily level [error, info, warn, debug]")
	certFile = flag.String("cert-file", "", "Path to the SSL certificate (only needed with secure enabled)")
	keyFile  = flag.String("key-file", "", "Path to the SSL key (only needed with secure enabled)")
)

// Used for finding the video url
const (
	prefix    = `\"video_url\":`
	quote     = `\"`
	prefixLen = len(prefix) + 1
)

func main() {
	flag.Parse()

	if _, err := strconv.Atoi(*port); err != nil {
		panic("port is not a valid integer")
	}

	if *secure && (*certFile == "" || *keyFile == "") {
		panic("No certificate file or key file provided")
	}

	var level slog.Level

	switch *logLevel {
	case "error":
		level = slog.LevelError
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "debug":
		level = slog.LevelDebug

	default:
		panic("Invalid logging level provided. Must be one of [error, info, warn, debug]")
	}

	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      level,
			TimeFormat: time.Kitchen,
		}),
	))

	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(gin.ErrorLogger())
	r.Use(RateLimiterMiddleware(NewRateLimiter(5, 10)))

	r.LoadHTMLGlob("templates/*")

	if *dev {
		r.Use(gin.Logger())
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r.GET("/reel/:id", cache.CacheByRequestURI(store, time.Minute*1), serveReel)
	r.GET("/reels/:id", cache.CacheByRequestURI(store, time.Minute*1), serveReel)
	r.GET("/p/:id", cache.CacheByRequestURI(store, time.Minute*1), serveReel)

	if *secure {
		slog.Info("Server running with TLS enabled", slog.String("listen", *port))
		r.RunTLS(":"+*port, *certFile, *keyFile)
	} else {
		slog.Info("Server running", slog.String("listen", *port))
		r.Run(":" + *port)
	}
}

func serveReel(c *gin.Context) {
	postId := c.Param("id")

	if postId == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if !strings.Contains(strings.ToLower(c.Request.Header.Get("User-Agent")), "discord") {
		c.Redirect(http.StatusPermanentRedirect, "https://instagram.com/reel/"+postId)
		return
	}

	videoUrl, err := GetCdnUrl(postId)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		slog.Error("Failed to get video URL from cdn", slog.Any("err", err))
		return
	}

	if videoUrl == "" {
		slog.Warn("Instagram returned an empty video URL. This most likely means the video is age restricted")
		c.HTML(http.StatusOK, "no_url.html", nil)
		return
	}

	remote, err := url.Parse(videoUrl)
	if err != nil {
		slog.Error("Failed to parse video url", slog.Any("err", err))
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(r *http.Request) {
		r.Header = c.Request.Header
		r.Host = remote.Host
		r.URL = remote
		r.Header = c.Request.Header.Clone()

		hopHeaders := []string{
			"Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding",
		}

		for _, h := range hopHeaders {
			r.Header.Del(h)
		}
	}

	c.Header("Cache-Control", "max-age=43200")
	proxy.ServeHTTP(c.Writer, c.Request)

	slog.Debug("Everything is OK. Request finished without any errors")
}

// Attempts to get the URL to the reel directly from the CDN
func GetCdnUrl(postId string) (string, error) {
	origin := "https://instagram.com/p/" + postId + "/embed/captioned"

	slog.Debug("Preparing request", slog.String("origin", origin))
	req, err := http.NewRequest("GET", origin, nil)
	if err != nil {
		return "", fmt.Errorf("failed to prepare HTTP request: %v", err)
	}

	// Set the user agent to firefox on pc so we get the correct stuff
	req.Header.Set("User-Agent", "Mozilla/5.0 (platform; rv:gecko-version) Gecko/gecko-trail Firefox/firefox-version")

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %v", err)
	}

	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)
	scanner.Buffer(make([]byte, 16*1024), 1024*1024)

	slog.Debug("Scanning response body for video url")
	for scanner.Scan() {
		line := scanner.Text()
		if url, found := ExtractUrl(line); found && url != "" {
			return url, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error scanning response: %v", err)
	}

	return "", nil
}

// Extracts the url from escaped json
func ExtractUrl(s string) (string, bool) {
	// Thanks a lot for this tyler
	// Find the first "video_url:"
	startIdx := strings.Index(s, prefix)
	if startIdx == -1 {
		return "", false
	}

	// Offset start by prefix len
	start := startIdx + prefixLen

	end := strings.Index(s[start:], quote)
	if end == -1 {
		return "", false
	}

	result := s[start : start+end]
	result = UnescapeJSONString(result)
	result = strings.ReplaceAll(result, `\/`, `/`)

	return result[1:], true
}

func S2B(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func B2S(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func UnescapeJSONString(s string) string {
	n := strings.IndexByte(s, '\\')
	if n < 0 {
		// Fast path - nothing to unescape.
		return s
	}

	// Slow path - unescape string.
	b := S2B(s) // It is safe to do, since s points to a byte slice in Parser.b.
	b = b[:n]
	s = s[n+1:]
	for len(s) > 0 {
		ch := s[0]
		s = s[1:]
		switch ch {
		case '"':
			b = append(b, '"')
		case '\\':
			b = append(b, '\\')
		case '/':
			b = append(b, '/')
		case 'b':
			b = append(b, '\b')
		case 'f':
			b = append(b, '\f')
		case 'n':
			b = append(b, '\n')
		case 'r':
			b = append(b, '\r')
		case 't':
			b = append(b, '\t')
		case 'u':
			if len(s) < 4 {
				// Too short escape sequence. Just store it unchanged.
				b = append(b, "\\u"...)
				break
			}
			xs := s[:4]
			x, err := strconv.ParseUint(xs, 16, 16)
			if err != nil {
				// Invalid escape sequence. Just store it unchanged.
				b = append(b, "\\u"...)
				break
			}
			s = s[4:]
			if !utf16.IsSurrogate(rune(x)) {
				b = append(b, string(rune(x))...)
				break
			}

			// Surrogate.
			// See https://en.wikipedia.org/wiki/Universal_Character_Set_characters#Surrogates
			if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
				b = append(b, "\\u"...)
				b = append(b, xs...)
				break
			}
			x1, err := strconv.ParseUint(s[2:6], 16, 16)
			if err != nil {
				b = append(b, "\\u"...)
				b = append(b, xs...)
				break
			}
			r := utf16.DecodeRune(rune(x), rune(x1))
			b = append(b, string(r)...)
			s = s[6:]
		default:
			// Unknown escape sequence. Just store it unchanged.
			b = append(b, '\\', ch)
		}
		n = strings.IndexByte(s, '\\')
		if n < 0 {
			b = append(b, s...)
			break
		}
		b = append(b, s[:n]...)
		s = s[n+1:]
	}
	return B2S(b)
}
