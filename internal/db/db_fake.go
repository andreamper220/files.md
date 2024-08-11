package db

import (
	"zakirullin/stuffbot/pkg/tg"
)

// FakeDB is a fake database, used for testing
// We don't have to clear it after each test
type FakeDB struct {
	DirByMessageID      string
	FilenameByMessageID string
}

func NewFakeDB() *FakeDB {
	return &FakeDB{}
}

func (db *FakeDB) LastKeyboardMsgID(userID int64) (int, bool) {
	return 0, false
}

func (db *FakeDB) SetLastKeyboardMsgID(userID int64, ID int) {

}

func (db *FakeDB) DelLastKeyboardMsgID(userID int64) {
}

func (db *FakeDB) InputExpectation(userID int64) *tg.Cmd {
	return nil
}

func (db *FakeDB) SetInputExpectation(userID int64, cmd tg.Cmd) {
}

func (db *FakeDB) DelInputExpectation(userID int64) {
}

func (db *FakeDB) SetFilenameByMsgID(userID int64, msgID int, filename string) {
	db.FilenameByMessageID = filename
}

func (db *FakeDB) FilenameByMsgID(userID int64, msgID int) string {
	return db.FilenameByMessageID
}

func (db *FakeDB) DirByMsgID(userID int64, msgID int) string {
	return db.DirByMessageID
}

func (db *FakeDB) SetDirByMsgID(userID int64, msgID int, dir string) {
	db.DirByMessageID = dir
}

func (db *FakeDB) QuickCommand(userID int64) (string, bool) {
	return "", false
}

func (db *FakeDB) SetQuickCommand(userID int64, cmd string) {
}

func (db *FakeDB) QuickCommandParams(userID int64) ([]string, bool) {
	return nil, false
}

func (db *FakeDB) SetQuickCommandParams(userID int64, params []string) {
}
