# How to use
Simply add `vx` to instagram URLs like so:
```
https://vxinstagram.com/....
```
This will attempt to embed the instagram video like so:<br>
![image](https://github.com/user-attachments/assets/b124bb26-0815-4b34-b8f5-70da24dcec20)

Age restricted reels may sometimes result in VxInstagram failing:<br>
![image](https://github.com/user-attachments/assets/bb090f29-166b-4c9e-96e3-2a1b3f9ac216)<br>
Unfortunately I haven't found a fix for this as of yet, but I'm working on bypassing this

Clicking the VxInstagram URL will redirect you to the original post

# How to self-host
> [!WARNING]
> This assumes you have git and golang working correctly
```sh
git clone https://github.com/Reishimanfr/vxinstagram
cd vxinstagram
go build -ldflags "-s -w" -tags=jsoniter .
```
## With SSL
> [!TIP]
> You can get a free SSL certificate from [Let's Encrypt](https://letsencrypt.org/) using [certbot](https://certbot.eff.org/)
```sh
./vxinstagram --port=8080 --cert-file=path/to/your/certificate --key-file=path/to/your/key
```
## Without SSL
```sh
./vxinstagram --port=8080
```

## Docker
```sh
# Copy the example docker-compose file
cp docker-compose.yml.example docker-compose.yml
# Edit the docker-compose file
vim docker-compose.yml
# Start the container
docker-compose up -d
```

## Configuration

The server can be configured using either command-line flags or environment variables. Flags take precedence over environment variables.

| Flag        | Environment Variable  | Default | Description                                           |
|-------------|-----------------------|---------|-------------------------------------------------------|
| --port      | PORT                  | 8080    | Port to run the server on                             |
| --gin-logs  | GIN_LOGS              | false   | Enable gin debug logs                                 |
| --secure    | SECURE                | false   | Use a secure connection                               |
| --log-level | LOG_LEVEL             | info    | Logging verbosity level [debug, error, warn, info]    |
| --cert-file | CERT_FILE             |         | Path to the SSL certificate (needed with secure mode) |
| --key-file  | KEY_FILE              |         | Path to the SSL key (needed with secure mode)         |
| --sentry-dsn| SENTRY_DSN            |         | Sentry DSN used for telemetry                         |


# Task list
- [ ] Find a way to fix some reels not embedding
- [ ] Add Open Graph embeds to videos
- [ ] Add additional info (like the amount of likes) to the Open Graph embed
- [ ] Add monitoring dashboard capabilities
- [ ] Create deployment scripts (for docker and some services)
- [ ] Create an action to automatically compile the binary and release it
- [ ] Fix reels with usernames at the beginning not working (/:username/reel/:postId)
