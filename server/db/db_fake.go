package db

import (
	"github.com/zakirullin/files.md/server/pkg/tg"
)

// FakeDB is a fake database, used for testing
// We don't have to clear it after each test
type FakeDB struct {
	HashOrPathByMID     string
	InputExpectationCMD *tg.Cmd
	EditNoteDirHash     string
	EditNoteFileHash    string
	EditNoteMode        string
	LastKeyboardMID     int
	RecentCMD           string
	RecentCMDParams     []string
	PendingDrafts       map[string]string
}

func NewFakeDB() *FakeDB {
	return &FakeDB{LastKeyboardMID: -1}
}

func (db *FakeDB) LastKeyboardMsgID() (int, bool) {
	if db.LastKeyboardMID == -1 {
		return 0, false
	}

	return db.LastKeyboardMID, true
}

func (db *FakeDB) SetLastKeyboardMsgID(msgID int) {
	db.LastKeyboardMID = msgID
}

func (db *FakeDB) DelLastKeyboardMsgID() {
	db.LastKeyboardMID = -1
}

func (db *FakeDB) InputExpectation() *tg.Cmd {
	return db.InputExpectationCMD
}

func (db *FakeDB) SetInputExpectation(cmd tg.Cmd) {
	db.InputExpectationCMD = &cmd
}

func (db *FakeDB) DelInputExpectation() {
	db.InputExpectationCMD = nil
}

func (db *FakeDB) SetEditNoteTarget(dirHash, filenameHash, mode string) {
	db.EditNoteDirHash = dirHash
	db.EditNoteFileHash = filenameHash
	db.EditNoteMode = mode
}

func (db *FakeDB) EditNoteTarget() (dirHash, filenameHash, mode string, ok bool) {
	if db.EditNoteDirHash == "" || db.EditNoteFileHash == "" {
		return "", "", "", false
	}
	if db.EditNoteMode == "" {
		return db.EditNoteDirHash, db.EditNoteFileHash, "r", true
	}
	return db.EditNoteDirHash, db.EditNoteFileHash, db.EditNoteMode, true
}

func (db *FakeDB) DelEditNoteTarget() {
	db.EditNoteDirHash = ""
	db.EditNoteFileHash = ""
	db.EditNoteMode = ""
}

func (db *FakeDB) HashOrPathByMsgID(msgID int) (string, bool) {
	return db.HashOrPathByMID, db.HashOrPathByMID != ""
}

func (db *FakeDB) SetHashOrPathByMsgID(msgID int, value string) {
	db.HashOrPathByMID = value
}

func (db *FakeDB) RecentCommand() (string, bool) {
	return db.RecentCMD, db.RecentCMD != ""
}

func (db *FakeDB) SetRecentCommand(cmd string) {
	db.RecentCMD = cmd
}

func (db *FakeDB) RecentCommandParams() ([]string, bool) {
	return db.RecentCMDParams, len(db.RecentCMDParams) > 0
}

func (db *FakeDB) SetRecentCommandParams(params []string) {
	db.RecentCMDParams = params
}

func (db *FakeDB) AddImgMsgID(msgID int) {
}

func (db *FakeDB) ImgMsgID() ([]int, bool) {
	return nil, false
}

func (db *FakeDB) DelImgMsgID() {
}

func (db *FakeDB) SetPendingDraft(hash, content string) {
	if db.PendingDrafts == nil {
		db.PendingDrafts = make(map[string]string)
	}
	db.PendingDrafts[hash] = content
}

func (db *FakeDB) PendingDraft(hash string) (string, bool) {
	if db.PendingDrafts == nil {
		return "", false
	}
	c, ok := db.PendingDrafts[hash]
	return c, ok
}

func (db *FakeDB) DelPendingDraft(hash string) {
	if db.PendingDrafts != nil {
		delete(db.PendingDrafts, hash)
	}
}
