package i18n

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed ru.json
var ruJSON string

var ruTranslations map[string]string

func init() {
	ruTranslations = make(map[string]string)
	if err := json.Unmarshal([]byte(ruJSON), &ruTranslations); err != nil {
		panic(fmt.Errorf("i18n: can't load ru.json: %w", err))
	}
}
