package core

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/kotlin"
	"github.com/smacker/go-tree-sitter/python"
	"golang.org/x/exp/slices"
)

type LangType string

const (
	LangJava       LangType = "JAVA"
	LangGo         LangType = "GOLANG"
	LangPython     LangType = "PYTHON"
	LangKotlin     LangType = "KOTLIN"
	LangJavaScript LangType = "JAVASCRIPT"
	LangUnknown    LangType = "UNKNOWN"
)

var SupportedLangs = []LangType{
	LangJava,
	LangGo,
	LangPython,
	LangKotlin,
	LangJavaScript,
}

func (langType LangType) IsSupported() bool {
	return slices.Contains(SupportedLangs, langType)
}

func (langType LangType) GetValue() string {
	return string(langType)
}

func LangTypeValueOf(raw string) LangType {
	switch raw {
	case LangJava.GetValue():
		return LangJava
	case LangGo.GetValue():
		return LangGo
	case LangPython.GetValue():
		return LangPython
	case LangKotlin.GetValue():
		return LangKotlin
	case LangJavaScript.GetValue():
		return LangJavaScript
	default:
		return LangUnknown
	}
}

func (langType LangType) GetLanguage() *sitter.Language {
	// look up parser from tree-sitter
	switch langType {
	case LangJava:
		return java.GetLanguage()
	case LangGo:
		return golang.GetLanguage()
	case LangPython:
		return python.GetLanguage()
	case LangKotlin:
		return kotlin.GetLanguage()
	case LangJavaScript:
		return javascript.GetLanguage()
	}
	return nil
}

func (langType LangType) GetFileSuffix() string {
	switch langType {
	case LangJava:
		return ".java"
	case LangGo:
		return ".go"
	case LangPython:
		return ".py"
	case LangKotlin:
		return ".kt"
	case LangJavaScript:
		return ".js"
	}
	return ""
}

func (langType LangType) MatchName(name string) bool {
	return strings.HasSuffix(name, langType.GetFileSuffix())
}
