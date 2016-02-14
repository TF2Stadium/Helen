-- name: create-player-slots-view
CREATE OR REPLACE VIEW player_slots AS
SELECT players.steam_id,
       players.name,
       lobby_slots.slot,
       lobby_slots.lobby_id,
       lobby_slots.needs_sub,
       lobbies.state
FROM players
INNER JOIN lobby_slots ON lobby_slots.player_id = players.id
INNER JOIN lobbies ON lobby_slots.lobby_id = lobbies.id;
