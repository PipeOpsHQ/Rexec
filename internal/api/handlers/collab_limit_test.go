package handlers

import "testing"

func TestSessionHasRoom(t *testing.T) {
	full := &CollabSession{
		MaxUsers: 1,
		Participants: map[string]*CollabParticipant{
			"owner": {UserID: "owner", Role: "owner"},
		},
	}
	if sessionHasRoom(full, "stranger") {
		t.Fatal("full session accepted a new user")
	}
	if !sessionHasRoom(full, "owner") {
		t.Fatal("existing participant was blocked")
	}

	closed := &CollabSession{MaxUsers: 0, Participants: map[string]*CollabParticipant{}}
	if sessionHasRoom(closed, "anyone") {
		t.Fatal("MaxUsers 0 accepted a new user")
	}
	closed.Participants["back"] = &CollabParticipant{UserID: "back", Role: "viewer"}
	if !sessionHasRoom(closed, "back") {
		t.Fatal("MaxUsers 0 blocked an existing participant")
	}
	if sessionHasRoom(nil, "owner") || sessionHasRoom(full, "") {
		t.Fatal("nil session or empty user was allowed")
	}
}
