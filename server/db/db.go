// Package db provides an in-memory database for storing user-specific temporary data.
package db

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/zakirullin/files.md/server/pkg/tg"
)

// In-memory database
var (
	hashOrPathByMsgID     sync.Map
	inputExpectations     sync.Map
	recentCommands        sync.Map
	recentCommandsTargets sync.Map
	sentPhotoMsgIDs       sync.Map
	pendingDrafts         sync.Map
	editNoteTargets       sync.Map
	editTaskTargets       sync.Map
)

// DB Do we need a type at all?
type DB struct {
	UserID int64
}

func NewDB(userID int64) *DB {
	return &DB{UserID: userID}
}

// TODO add locks
func (db *DB) LastKeyboardMsgID() (int, bool) {
	msgIDStr, err := os.ReadFile(tmpFilePath(db.UserID, "msgid"))
	if err != nil {
		return 0, false
	}

	msgID, err := strconv.Atoi(string(msgIDStr))
	if err != nil {
		return 0, false
	}

	return msgID, true
}

func (db *DB) SetLastKeyboardMsgID(ID int) {
	_ = os.WriteFile(tmpFilePath(db.UserID, "msgid"), []byte(strconv.Itoa(ID)), 0o644)
}

func (db *DB) DelLastKeyboardMsgID() {
	_ = os.Remove(tmpFilePath(db.UserID, "msgid"))
}

func (db *DB) InputExpectation() *tg.Cmd {
	val, ok := inputExpectations.Load(inputExpectationKey(db.UserID))
	if !ok {
		return nil
	}

	cmd := val.(tg.Cmd)
	return &cmd
}

func (db *DB) SetInputExpectation(cmd tg.Cmd) {
	inputExpectations.Store(inputExpectationKey(db.UserID), cmd)
}

func (db *DB) DelInputExpectation() {
	inputExpectations.Delete(inputExpectationKey(db.UserID))
}

// SetEditNoteTarget marks a note being edited (text and/or file attachments).
// mode is "r" (replace text) or "a" (append).
func (db *DB) SetEditNoteTarget(dirHash, filenameHash, mode string) {
	editNoteTargets.Store(editNoteTargetKey(db.UserID), dirHash+"/"+filenameHash+"/"+mode)
}

// EditNoteTarget returns the note currently being edited.
func (db *DB) EditNoteTarget() (dirHash, filenameHash, mode string, ok bool) {
	v, ok := editNoteTargets.Load(editNoteTargetKey(db.UserID))
	if !ok {
		return "", "", "", false
	}
	parts := strings.SplitN(v.(string), "/", 3)
	if len(parts) < 2 {
		return "", "", "", false
	}
	if len(parts) == 2 {
		return parts[0], parts[1], "r", true
	}
	return parts[0], parts[1], parts[2], true
}

func (db *DB) DelEditNoteTarget() {
	editNoteTargets.Delete(editNoteTargetKey(db.UserID))
}

// SetEditTaskTarget marks a task being edited (text and/or file attachments).
// params identify the task: [c,msgHash], [k,checklistHash,itemHash], or [a,projectHash,itemHash].
// mode is "r" (replace) or "a" (append).
func (db *DB) SetEditTaskTarget(params []string, mode string) {
	parts := append(append([]string(nil), params...), mode)
	editTaskTargets.Store(editTaskTargetKey(db.UserID), strings.Join(parts, "/"))
}

// EditTaskTarget returns the task currently being edited.
func (db *DB) EditTaskTarget() (params []string, mode string, ok bool) {
	v, ok := editTaskTargets.Load(editTaskTargetKey(db.UserID))
	if !ok {
		return nil, "", false
	}
	parts := strings.Split(v.(string), "/")
	if len(parts) < 3 {
		return nil, "", false
	}
	mode = parts[len(parts)-1]
	params = parts[:len(parts)-1]
	return params, mode, true
}

func (db *DB) DelEditTaskTarget() {
	editTaskTargets.Delete(editTaskTargetKey(db.UserID))
}

// HashOrPathByMsgID returns the target the bot rendered for this msgID -
// either a chat-block hash or an absolute file path. Callers distinguish
// the two by the leading "/" (paths) vs no slash (hash).
func (db *DB) HashOrPathByMsgID(msgID int) (string, bool) {
	v, ok := hashOrPathByMsgID.Load(hashOrPathByMsgIDKey(db.UserID, msgID))
	if !ok {
		return "", false
	}
	return v.(string), true
}

func (db *DB) SetHashOrPathByMsgID(msgID int, value string) {
	hashOrPathByMsgID.Store(hashOrPathByMsgIDKey(db.UserID, msgID), value)
}

func (db *DB) RecentCommand() (string, bool) {
	cmd, ok := recentCommands.Load(db.UserID)
	if !ok {
		return "", false
	}

	return cmd.(string), true
}

func (db *DB) SetRecentCommand(cmd string) {
	recentCommands.Store(db.UserID, cmd)
}

func (db *DB) RecentCommandParams() ([]string, bool) {
	params, ok := recentCommandsTargets.Load(db.UserID)
	if !ok {
		return nil, false
	}

	return params.([]string), true
}

func (db *DB) SetRecentCommandParams(params []string) {
	recentCommandsTargets.Store(db.UserID, params)
}

func (db *DB) AddImgMsgID(msgID int) {
	key := photoMsgIDKey(db.UserID)
	if val, ok := sentPhotoMsgIDs.Load(key); ok {
		ids := val.([]int)
		sentPhotoMsgIDs.Store(key, append(ids, msgID))
	} else {
		sentPhotoMsgIDs.Store(key, []int{msgID})
	}
}

func (db *DB) ImgMsgID() ([]int, bool) {
	key := photoMsgIDKey(db.UserID)
	val, ok := sentPhotoMsgIDs.Load(key)
	if !ok {
		return nil, false
	}

	ids, _ := val.([]int)
	return ids, true
}

func (db *DB) DelImgMsgID() {
	key := photoMsgIDKey(db.UserID)
	sentPhotoMsgIDs.Delete(key)
}

func photoMsgIDKey(userID int64) string {
	return fmt.Sprintf("%d:sentPhotoMsgIDs", userID)
}

func inputExpectationKey(userID int64) string {
	return fmt.Sprintf("%d:inputExpectations", userID)
}

func editNoteTargetKey(userID int64) string {
	return fmt.Sprintf("%d:editNoteTarget", userID)
}

func editTaskTargetKey(userID int64) string {
	return fmt.Sprintf("%d:editTaskTarget", userID)
}

func hashOrPathByMsgIDKey(userID int64, msgID int) string {
	return fmt.Sprintf("%d:hashOrPathByMsgID:%d", userID, msgID)
}

func tmpFilePath(userID int64, name string) string {
	return fmt.Sprintf("%s/%d.%s", os.TempDir(), userID, name)
}

func (db *DB) SetPendingDraft(hash, content string) {
	pendingDrafts.Store(pendingDraftKey(db.UserID, hash), content)
}

func (db *DB) PendingDraft(hash string) (string, bool) {
	v, ok := pendingDrafts.Load(pendingDraftKey(db.UserID, hash))
	if !ok {
		return "", false
	}
	return v.(string), true
}

func (db *DB) DelPendingDraft(hash string) {
	pendingDrafts.Delete(pendingDraftKey(db.UserID, hash))
}

func pendingDraftKey(userID int64, hash string) string {
	return fmt.Sprintf("%d:pending:%s", userID, hash)
}
