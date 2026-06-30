package db

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEditNoteTarget_PersistsAcrossDBInstances(t *testing.T) {
	r := require.New(t)
	const userID int64 = 4242

	db1 := NewDB(userID)
	db1.SetEditNoteTarget("abc12", "def34", "r")

	// Simulate process restart: in-memory state is gone, tmp file remains.
	editNoteTargets.Delete(editNoteTargetKey(userID))

	db2 := NewDB(userID)
	dir, file, mode, ok := db2.EditNoteTarget()
	r.True(ok)
	r.Equal("abc12", dir)
	r.Equal("def34", file)
	r.Equal("r", mode)

	db2.DelEditNoteTarget()
	_, _, _, ok = NewDB(userID).EditNoteTarget()
	r.False(ok)

	_, err := os.Stat(tmpFilePath(userID, "editNote"))
	r.True(os.IsNotExist(err))
}
