## How to use
Simply prefix instagram urls with `vx` like so: 
```
https://instagram.com/
```
Turns into
```
https://vxinstagram.com/
```

## How to self-host
> [!WARNING]
> This assumes you have git and golang working correctly
```sh
git clone https://github.com/Reishimanfr/vxinstagram
cd vxinstagram
go build -ldflags "-s -w" -tags=jsoniter .
```
### With SSL
> [!TIP]
> You can get a free SSL certificate from [Let's Encrypt](https://letsencrypt.org/) using [certbot](https://certbot.eff.org/)
```sh
./vxinstagram --port="8080" --cert-file="path/to/your/certificate" --key-file="path/to/your/key"
```

### Without SSL
```sh
./vxinstagram --port="8080"
```
