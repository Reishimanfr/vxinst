> This project is not associated with Instagram or Meta in any way.
# What is this?
VxInst is an open-source, blazing fast server that scrapes the direct URL to a video from instagram's CDN to fix embedding in apps like discord.

> [!TIP]
> If you're looking for the public API stuff, go to [the wiki](https://github.com/Reishimanfr/vxinst/wiki)
# How can I use it?
Simply add `vx` to instagram URLs like so:
```
https://vxinst.com/...
```
This will attempt to embed the instagram video:<br>
![image](https://github.com/user-attachments/assets/4e129b3a-efe2-4c42-b15b-25e8a9b51e2e)<br>
> [!TIP]
> Clicking the VxInst URL will redirect you to the original post
# Self-hosting
## 🍏 Mac and Linux 🐧
<details>
<summary>Click to show</summary>

You can either compile the binary from source or download a precompiled binary from the [releases tab](https://github.com/Reishimanfr/vxinst/releases).<br>

### Compilation
```ps
# Clone the repository
git clone --depth=1 https://github.com/Reishimanfr/vxinst

# Cd into the directory
cd vxinst

# Compile the code
go build -ldflags "-s -w" -tags=jsoniter -o vxinsta
```

Check out the examples on how to run VxInst

</details>

## 🪟 Windows
<details>
<summary>Click to show</summary>

There are no precompiled binaries for windows meaning you'll need to compile the code from source.

```ps
# Clone the repository
git clone --depth=1 https://github.com/Reishimanfr/vxinst

# Cd into the directory
cd vxinst

# Compile the code
go build -ldflags "-s -w" -tags=jsoniter -o vxinst.exe
```

</details>

## 🐋 Docker 
<details>
<summary>Click to show</summary>

```sh
# Copy the example docker-compose file
cp docker-compose.yml.example docker-compose.yml

# Edit the docker-compose file
vim docker-compose.yml

# Start the container
docker-compose up -d
```

</details>

## ⚙️ Configuration
The server can be configured using either command-line flags or environment variables. Flags take precedence over environment variables.

| Flag                  | Environment Variable  | Default  | Description                                              |
|-----------------------|-----------------------|----------|----------------------------------------------------------|
| --port                | PORT                  | 8080     | Port to run the server on                                |
| --gin-logs            | GIN_LOGS              | false    | Enable gin debug logs                                    |
| --secure              | SECURE                | false    | Use a secure connection                                  |
| --log-level           | LOG_LEVEL             | info     | Logging verbosity level [debug, error, warn, info]       |
| --cert-file           | CERT_FILE             |          | Path to the SSL certificate (needed with secure mode)    |
| --key-file            | KEY_FILE              |          | Path to the SSL key (needed with secure mode)            |
| --sentry-dsn          | SENTRY_DSN            |          | Sentry DSN used for telemetry                            |
| --cache-lifetime      | CACHE_LIFETIME        | 60       | Time to keep cache for (in minutes)                      |
| --memory-lifetime     | MEMORY_LIFETIME       | 7        | Time to keep memory cache for (in days)                  |
| --redis-enable        | REDIS_ENABLE          | false    | Enables redis for caching (memory if set to false)       |
| --redis-address       | REDIS_ADDR            |          | Address to redis database                                |
| --redis-passwd        | REDIS_PASSWD          |          | Password for redis database                              |
| --redis-db            | REDIS_DB              | -1       | Redis database to use                                    |
| --proxies             | PROXIES               |          | Proxies to make request with. Provide multiple to cycle  | 
| --proxy-scrape-html   | PROXY_SCRAPE_HTML     | false    | Sets if proxies should scrape HTML. May use up bandwidth |
| --insta-cookie        | INSTA_COOKIE          |          | User cookie for API calls with for age restricted posts  |                       
| --insta-xigappid      | INSTA_XIGAPPID        |          | X-IG-App-ID for API calls                                |
| --insta-browser-agent | INSTA_BROWSER_AGENT   | *        | <Firefox, Linux, X11>                                    |

\* = Mozilla/5.0 (X11; Linux x86_64; rv:135.0) Gecko/20100101 Firefox/135.0

## 📚 Examples on running VxInst
Run on the default port with no TLS
```ps
./vxinst
```

Run with TLS enabled, a proxy attached, redis for cache and instagram credentials (recommended)
```ps
./vxinst \
        --secure
        --cert-file="/path/to/your/ssl/certificate"
        --key=file="/path/to/your/ssl/key"
        --proxies="http://yourproxyiphere:someport"
        --redis-enable
        --redis-address="127.0.0.1:6379"
        --redis-passwd="my_very_safe_password"
        --redis-db=0
        --insta-cookie="your instagram cookie here"
        --insta-xigappid="x-ig-app-id here"
```

# 📋 Task list
- [x]  ~~Find a way to fix some reels not embedding~~
- [x] Add Open Graph embeds to videos
- [ ] Add additional info (like the amount of likes) to the Open Graph embed
- [ ] Add monitoring dashboard capabilities
- [x] Create deployment scripts (for docker and some services)
- [x] Create an action to automatically compile the binary and release it
- [ ] Fix reels with usernames at the beginning not working (/:username/reel/:postId)
