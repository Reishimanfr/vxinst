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
./vxinstagram --port="8080" --cert-file="path/to/your/certificate" --key-file="path/to/your/key"
```
## Without SSL
```sh
./vxinstagram --port="8080"
```

# Available flags
| Flag | Description | Default |
| ---- | ----------- | ------- |
| `-port` | Sets the port for the server to listen on | 8080 |
| `-log-level` | Sets the logging level `[debug, info, warn, error]` | info |
| `-dev` | Enables gin debugging data | false |
| `-secure` | Enable https instead of http | false |
| `-cert-file` | Path to your SSL certificate file (if `-secure` is `true`) | - |
| `-key-file` | Path to your SSL key file (if `-secure` is `true`) | - |

# Task list
- [ ] Find a way to fix some reels not embedding
- [ ] Add Open Graph embeds to videos
- [ ] Add additional info (like the amount of likes) to the Open Graph embed
- [ ] Add monitoring dashboard capabilities
- [ ] Create deployment scripts (for docker and some services)
- [ ] Create an action to automatically compile the binary and release it
