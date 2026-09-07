package scripts

import "testing"

func TestGetLanguageDefaultsToLua(t *testing.T) {
	s := Script{}
	if s.GetLanguage() != ScriptLanguageLua {
		t.Fatalf("got %q want lua", s.GetLanguage())
	}
	s.Language = ScriptLanguageJavaScript
	if s.GetLanguage() != ScriptLanguageJavaScript {
		t.Fatalf("got %q want javascript", s.GetLanguage())
	}
}
