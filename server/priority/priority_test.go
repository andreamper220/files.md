package priority

import "testing"

func TestApplyPreservesImageMarkdown(t *testing.T) {
	text := "![](media/tg_PHOTO_ID)\nCaption"
	got := Apply(text, "🔴", DefaultEmojis)
	want := "🔴 ![](media/tg_PHOTO_ID)\nCaption"
	if got != want {
		t.Fatalf("Apply() = %q, want %q", got, want)
	}
}
