# Go-Epoll!
A libuv-like epoll wrapper.

Inspired by [Issue #15735](https://github.com/golang/go/issues/15735)

It's a pity that the issue #15735 was denied.

So I make this.

# Example

It's very easy to use it.

```
tcpConn, err := net.Dial("tcp", "go.dev:80")
onread := func (c *goepoll.Conn) {
    buf := make([]byte, 100)
    n, err := c.Read(buf)
    buf = buf[0:n]
    fmt.Println(string(buf))
}
c := goepoll.NewConn(tcpConn, onread, nil, nil)
events, err := goepoll.New() 
c, err := events.Add(tcpConn, goepoll.EVENT_READABLE)
c.Write([]byte("Hello"))
```

# Why?

Because this can avoid the time of awoking a goroutine.

The futex/awoke is slow.