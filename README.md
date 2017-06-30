# Keker
Easy to use pub-sub message broker via websocket.

#### About:

In recent time become very popular - SPA(single page applications), when all 'view' it`s - 'javascript bundle', which loads on client side and communicates with 'backend' API.

Sometimes SPA wants to provide some interactive events for it's clients, like - friendship request, new message request or something else. So, there are many ways to provide it, Let's look at most popular:
 * `polling` (Bad way, because this generates unnecessary load on the server(API) without the guarantee of getting the desired(new) data)
 * `websocket` (Right way, SPA makes only one TCP (full-duplex) connection and waits from backend for new data)
 
Keker provides websocket's for SPA, for subscriptions, and event listener URL, where the backend sends all events.

### How to use:

#### Build `keker`

```sh
git clone https://github.com/tears-of-noobs/keker.git
cd keker
make
```

#### Run `keker`

```sh
./out/kekerd
```

After start kekerd listen on all interfaces, port - `3498`. You may change some default values, see `--help` for additional information.

#### Connecting SPA to kekerd
WARNING: You must send `HELLO` message when your connection is open. Format is simple - `HELLO <AUTH_INFO>`, where `<AUTH_INFO>` it's information about client (name, auth token, etc...). If you don't send `HELLO` message, connection will be fail after one minute.

```js
var socket = new WebSocket("ws://YOU_KEKER_IP:3498/v1/subscribe");

socket.onopen = () => {
    socket.send(`HELLO ${authToken}`);
};

socket.onmessage = function(event) {
  console.log(event.data);
};
```
#### Publish events from backend
For publishing events you can send HTTP request with data in arbitrary format(JSON, plain text, etc...) to URL - `http://YOU_KEKER_IP:3498/v1/publish`, with Header - `Token: <AUTH_INFO>`, where `<AUTH_INFO>` - client identificator which used by SPA for subscribing to events.

Thats all :)

## License
----
MIT




