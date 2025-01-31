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
./vxinstagram --port=1234
```
Due to the nature of routers I can't help you with setting the rest up since every routed is different. Google is your best friend here
