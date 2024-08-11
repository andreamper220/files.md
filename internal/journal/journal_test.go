package journal

import (
	"testing"
	"time"

	"zakirullin/stuffbot/internal/fs"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestAddRecord(t *testing.T) {
	r := require.New(t)
	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(2023, 0o5, 30, 10, 0o4, 36, 0, time.UTC)
	}

	type testcase struct {
		name                string
		md                  string
		record              string
		want                string
		journalHeaderFormat string
	}

	tests := []testcase{
		{
			name:   "Empty MD",
			record: "note 1",
			want:   "#### 30 May, Tuesday\n`10:04` note 1\n",
		},
		{
			name:   "No Headers",
			md:     "some text",
			record: "note 1",
			want:   "some text\n#### 30 May, Tuesday\n`10:04` note 1\n",
		},
		{
			name:   "Bare header",
			md:     "#### 30 May, Tuesday\n",
			record: "note 1",
			want:   "#### 30 May, Tuesday\n`10:04` note 1\n",
		},
		{
			name:   "New daily note",
			md:     "#### 29 May, Tuesday\nnote 1",
			record: "note 2",
			want:   "#### 29 May, Tuesday\nnote 1\n#### 30 May, Tuesday\n`10:04` note 2\n",
		},
		{
			name:   "Append daily note",
			md:     "#### 29 May, Tuesday\nnote 1\n#### 30 May, Tuesday\nnote 2",
			record: "note 3",
			want:   "#### 29 May, Tuesday\nnote 1\n#### 30 May, Tuesday\nnote 2\n`10:04` note 3\n",
		},
		{
			name:   "Image without caption",
			md:     "some text",
			record: "![[../img/tg_HASH.jpg|center|400]]",
			want:   "some text\n#### 30 May, Tuesday\n`10:04` ![[../img/tg_HASH.jpg|center|400]]\n",
		},
		{
			name:   "Image with caption",
			md:     "some text",
			record: "![[../img/tg_HASH.jpg|center|400]]\nCaption",
			want:   "some text\n#### 30 May, Tuesday\n![[../img/tg_HASH.jpg|center|400]]\n`10:04` Caption\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userFS, err := fs.NewFS("/", afero.NewMemMapFs())
			r.NoError(err)
			userFS.Write(fs.DirJournal, "2023.05 May.md", test.md)
			userFS.Write(fs.DirToday, "record.md", test.record)

			err = AddRecord(userFS, "record.md")
			r.NoError(err)

			md, err := userFS.Read(fs.DirJournal, "2023.05 May.md")
			r.NoError(err)
			r.Equal(test.want, md)
		})
	}
}

func TestAddEmojiNewFile(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())

	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(2024, time.January, 1, 1, 0, 0, 0, time.UTC)
	}

	err = AddEmoji(userFS, "🙂")
	r.NoError(err)

	content, err := userFS.Read("journal", "2024.01 January.md")
	r.NoError(err)

	r.Equal("#### 1 January, Monday 🙂", content)
}

func TestAddEmojiExistingFile(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	md := "#### 0, Sunday\nSome Note\n#### 1 January, Monday\nSome Note"
	userFS.Write("journal", "2024.01 January.md", md)

	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(2024, time.January, 1, 1, 0, 0, 0, time.UTC)
	}

	err = AddEmoji(userFS, "🙂")
	r.NoError(err)

	content, err := userFS.Read("journal", "2024.01 January.md")
	r.NoError(err)

	r.Equal("#### 0, Sunday\nSome Note\n#### 1 January, Monday 🙂\nSome Note", content)
}

func TestAddEmojiExistingFileMissingDay(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	md := "#### 0, Sunday\nSome Note\n"
	userFS.Write("journal", "2024.01 January.md", md)

	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(2024, time.January, 1, 1, 0, 0, 0, time.UTC)
	}

	err = AddEmoji(userFS, "🙂")
	r.NoError(err)

	content, err := userFS.Read("journal", "2024.01 January.md")
	r.NoError(err)

	r.Equal("#### 0, Sunday\nSome Note\n#### 1 January, Monday 🙂", content)
}

func TestAddMoodEmojiExistingFileExistingEmojis(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	md := "#### 0, Sunday\nSome Note\n#### 1 January, Monday 🌱📵\nSome Note"
	userFS.Write("journal", "2024.01 January.md", md)

	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(2024, time.January, 1, 1, 0, 0, 0, time.UTC)
	}

	err = AddEmoji(userFS, "🙂")
	r.NoError(err)

	content, err := userFS.Read("journal", "2024.01 January.md")
	r.NoError(err)

	r.Equal("#### 0, Sunday\nSome Note\n#### 1 January, Monday 🙂🌱📵\nSome Note", content)
}

func TestAddRegularEmojiExistingFileExistingEmojis(t *testing.T) {
	r := require.New(t)

	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	md := "#### 0, Sunday\nSome Note\n#### 1 January, Monday 🌱📵\nSome Note"
	userFS.Write("journal", "2024.01 January.md", md)

	savedNow := now
	defer func() {
		now = savedNow
	}()
	now = func() time.Time {
		return time.Date(2024, time.January, 1, 1, 0, 0, 0, time.UTC)
	}

	err = AddEmoji(userFS, "🎃")
	r.NoError(err)

	content, err := userFS.Read("journal", "2024.01 January.md")
	r.NoError(err)

	r.Equal("#### 0, Sunday\nSome Note\n#### 1 January, Monday 🌱📵🎃\nSome Note", content)
}
