
| Environment Variable | Description |
|----------------------|-------------|
|    `SERVER_ADDR`     |Address to serve on.|
|    `PUBLIC_ADDR`     |Publicly accessible address for the server, requires schema|
|    `SERVER_OPENID_REALM`     |The OpenID Realm (See: [Section 9.2 of the OpenID Spec](https://openid.net/specs/openid-authentication-2_0-12.html#realms))|
|    `CORS_WHITELIST`     ||
|    `SERVER_COOKIE_DOMAIN`     |Cookie URL domain|
|    `SERVER_REDIRECT_PATH`     |URL to redirect user to after a successful login|
|    `COOKIE_STORE_SECRET`     |base64 encoded key to use for encrypting cookies|
|    `MUMBLE_ADDR`     |Mumble Address|
|    `MUMBLE_PASSWORD`     |Mumble Password|
|    `STEAMID_WHITELIST`     |SteamID Group XML page to use to filter logins|
|    `MOCKUP_AUTH`     |Enable Mockup Authentication|
|    `GEOIP`     |Enable geoip support for getting the location of game servers|
|    `SERVE_STATIC`     |Serve /static/|
|    `RABBITMQ_URL`     |URL for AMQP server|
|    `PAULING_QUEUE`     |Name of queue over which RPC calls to Pauling are sent|
|    `TWITCHBOT_QUEUE`     |Name of queue over which RPC calls to Pauling are sent|
|    `FUMBLE_QUEUE`     |Name of queue over which RPC calls to Fumble are sent|
|    `RABBITMQ_QUEUE`     |Name of queue over which events are sent|
|    `DATABASE_ADDR`     |Database Address|
|    `DATABASE_NAME`     |Database Name|
|    `DATABASE_USERNAME`     |Database username|
|    `DATABASE_PASSWORD`     |Database password|
|    `STEAM_API_KEY`     |Steam API Key|
|    `PROFILER_ADDR`     |Address to serve the web-based profiler over|
|    `SLACK_URL`     |Slack webhook URL|
|    `TWITCH_CLIENT_ID`     |Twitch API Client ID|
|    `TWITCH_CLIENT_SECRET`     |Twitch API Client Secret|
|    `SERVEME_API_KEY`     |serveme.tf API Key|
|    `HEALTH_CHECKS`     |Enable health checks|
