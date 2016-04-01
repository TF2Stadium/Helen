package player_test

import (
	"testing"

	"github.com/TF2Stadium/Helen/internal/testhelpers"
	. "github.com/TF2Stadium/Helen/models/player"
	"github.com/stretchr/testify/assert"
	"time"
)

func init() {
	testhelpers.CleanupDB()
}

func TestReportSubs(t *testing.T) {
	t.Parallel()

	p := testhelpers.CreatePlayer()
	l1 := testhelpers.CreateLobby()
	defer l1.Close(false, false)
	l2 := testhelpers.CreateLobby()
	defer l2.Close(false, false)
	// l3 := testhelpers.CreateLobby()
	// defer l3.Close(false, false)

	p.NewReport(Substitute, l1.ID)
	p.NewReport(Substitute, l2.ID)

	banned, until := p.IsBannedWithTime(BanJoin)
	assert.True(t, banned, "Player should be banned from joining lobbies")
	assert.WithinDuration(t, until, time.Now(), 30*time.Minute)
}

func TestReportVoted(t *testing.T) {
	t.Parallel()
	p := testhelpers.CreatePlayer()
	l1 := testhelpers.CreateLobby()
	defer l1.Close(false, false)
	l2 := testhelpers.CreateLobby()
	defer l2.Close(false, false)

	// RageQuit = Vote + 1, so we don't need to test that
	p.NewReport(Vote, l1.ID)
	p.NewReport(Vote, l2.ID)

	banned, until := p.IsBannedWithTime(BanJoin)
	assert.True(t, banned, "Player should be banned from joining lobbies")
	assert.WithinDuration(t, until, time.Now(), 30*time.Minute)
}
