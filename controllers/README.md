controllers
===========

Helen uses WebSockets to communicate with the frontend, with the help of WSEvent.
Handlers exposed through websocket can be found at socket/handlers/

Each user on entering the site, is added to the room `0_public`, where all global
chat messages and lobby updates are sent to. Each lobby created has two rooms,
`0_public` and `0_private`. Lobby chat messages, and updates (player joined, left, 
lobby started, lobby closed, etc) are broadcasted on `0_public`, while ready up messages and
lobby server info are sent to `0_private`.
