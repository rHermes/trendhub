// Copyright 2019 Teodor Spæren
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

var (
	LangAny        = Language{StoreName: "any", QueryName: ""}
	LangJava       = Language{StoreName: "java", QueryName: "java"}
	LangKotlin     = Language{StoreName: "kotlin", QueryName: "kotlin"}
	LangGo         = Language{StoreName: "go", QueryName: "go"}
	LangC          = Language{StoreName: "c", QueryName: "c"}
	LangCPP        = Language{StoreName: "cpp", QueryName: "c++"}
	LangRust       = Language{StoreName: "rust", QueryName: "rust"}
	LangHaskell    = Language{StoreName: "haskell", QueryName: "haskell"}
	LangTypescript = Language{StoreName: "typescript", QueryName: "typescript"}
	LangPHP        = Language{StoreName: "php", QueryName: "php"}
	LangJavascript = Language{StoreName: "javascript", QueryName: "javascript"}
	LangAssembly   = Language{StoreName: "assembly", QueryName: "assembly"}
	LangRuby       = Language{StoreName: "ruby", QueryName: "ruby"}
	LangHTML       = Language{StoreName: "html", QueryName: "html"}
	LangUnknown    = Language{StoreName: "unknown", QueryName: "unknown"}

	StoreToLang = map[string]Language{
		LangAny.StoreName:        LangAny,
		LangJava.StoreName:       LangJava,
		LangKotlin.StoreName:     LangKotlin,
		LangGo.StoreName:         LangGo,
		LangC.StoreName:          LangC,
		LangCPP.StoreName:        LangCPP,
		LangRust.StoreName:       LangRust,
		LangHaskell.StoreName:    LangHaskell,
		LangTypescript.StoreName: LangTypescript,
		LangPHP.StoreName:        LangPHP,
		LangJavascript.StoreName: LangJavascript,
		LangAssembly.StoreName:   LangAssembly,
		LangRuby.StoreName:       LangRuby,
		LangHTML.StoreName:       LangHTML,
		LangUnknown.StoreName:    LangUnknown,
	}
)
