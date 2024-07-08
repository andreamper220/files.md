package journal

import (
	"testing"
	"time"

	"zakirullin/stuffbot/internal/fs"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func Test_AddRecord(t *testing.T) {
	r := require.New(t)
	now = func() time.Time {
		return time.Date(2023, 05, 30, 10, 04, 36, 0, time.UTC)
	}

	type testcase struct {
		name                string
		md                  string
		record                string
		want                string
		journalHeaderFormat string
	}

	tests := []testcase{
		{
			name: "Empty MD",
			record: "note 1",
			want: "#### 30, Tuesday\n`10:04` note 1\n",
		},
		{
			name: "No Headers",
			md:   "some text",
			record: "note 1",
			want: "some text\n#### 30, Tuesday\n`10:04` note 1\n",
		},
		{
			name: "Bare header",
			md:   "#### 30, Tuesday\n",
			record: "note 1",
			want: "#### 30, Tuesday\n`10:04` note 1\n",
		},
		{
			name: "New daily note",
			md:   "#### 29, Tuesday\nnote 1",
			record: "note 2",
			want: "#### 29, Tuesday\nnote 1\n#### 30, Tuesday\n`10:04` note 2\n",
		},
		{
			name: "Append daily note",
			md:   "#### 29, Tuesday\nnote 1\n#### 30, Tuesday\nnote 2",
			record: "note 3",
			want: "#### 29, Tuesday\nnote 1\n#### 30, Tuesday\nnote 2\n`10:04` note 3\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			botFS, err := fs.NewFS("/", afero.NewMemMapFs())
			r.NoError(err)
			botFS.Write(fs.DirJournal, "2023 May.md", test.md)
			botFS.Write(fs.DirToday, "record.md", test.record)

			err = AddRecord(botFS, "record.md")
			r.NoError(err)
			
			md, err := botFS.Read(fs.DirJournal, "2023 May.md")
			r.NoError(err)
			r.Equal(test.want, md)
		})
	}
}
