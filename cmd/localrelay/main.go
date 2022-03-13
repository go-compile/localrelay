package main

import (
	"os"

	"github.com/go-compile/localrelay"
)

func main() {
	r := localrelay.New("nextcloud", "127.0.0.1:90", "localhost:8080", os.Stdout)

	// r := localrelay.New("nextcloud", "127.0.0.1:90", "ltvpcj6ckjcwcbolp2wsnbdnxje4mil7xeb4j7gvafnr43gky7h377qd.onion:80", os.Stdout)
	// prox, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, nil)
	// if err != nil {
	// 	panic(err)
	// }

	// r.SetProxy(prox)

	// Close relay
	// go func() {
	// 	time.Sleep(time.Second * 15)
	// 	r.Close()
	// }()

	panic(r.ListenServe())
}
