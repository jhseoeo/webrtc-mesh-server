<div id="top"></div>

<br />
<div align="center">
  <h2 align="center">WebRTC Mesh Server</h2>

  <p align="center">
    A signaling server for many-to-many WebRTC connection witten in Golang
  </p>
</div>

## About The Project

Well....

If you want more information(in Korean), you can get here: <https://junhyuk0801.github.io/posts/post/Networking/WebRTC/many%20to%20many%20signaling%20server>

<br>

### Built With

-   [Golang](https://go.dev/)
-   [GoFiber](https://gofiber.io/)
-   [GoFiber/Websocket](https://github.com/gofiber/websocket)

<br>

## Prerequisites

-   Go 1.18 or above

<br>

## Getting Started

```bash
git clone https://github.com/junhyuk0801/webrtc-mesh-server
cd webrtc-mesh-server
go mod download # OR go mod vendor
go run ./...
```

Then connect <http://127.0.0.1:3005> with your browser and check how it works!
