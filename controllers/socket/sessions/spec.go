package sessions

//SetSpectator indicates that the socket with the given socketID is now
//spectating lobbyID
func SetSpectator(socketID string, lobbyID uint) {
	socketsMu.Lock()
	defer socketsMu.Unlock()
	socketSpectating[socketID] = lobbyID
}

//GetSpectating returns the lobbyID of the lobby the socketID is currently spectating.
//ok is false if the socket is not spectating any lobby
func GetSpectating(socketID string) (lobbyID uint, ok bool) {
	socketsMu.RLock()
	defer socketsMu.RUnlock()
	lobbyID, ok = socketSpectating[socketID]
	return
}

//RemoveSpectator indicates that socketID is no longer spectating the lobby it was earlier.
func RemoveSpectator(socketID string) {
	socketsMu.Lock()
	defer socketsMu.Unlock()
	delete(socketSpectating, socketID)
}

//IsSpectating returns whether socketID is spectating lobbyID
func IsSpectating(socketID string, lobbyID uint) bool {
	socketsMu.RLock()
	defer socketsMu.RUnlock()
	id, _ := socketSpectating[socketID]
	return id == lobbyID
}
