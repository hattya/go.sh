//
// go.sh/pattern :: pattern_windows_test.go
//
//   Copyright (c) 2021 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package pattern

import "testing"

var indexTests = []struct {
	pattern string
	i, w    int
}{
	{`_\\_`, 1, 2},
	{`_/_`, 1, 1},
	{`_\/_`, 1, 2},
	{`_\?\\_`, 3, 2},
	{`_\?/_`, 3, 1},
	{`_\?\/_`, 3, 2},
	{`\\`, 0, 2},
	{`/`, 0, 1},
	{`\/`, 0, 2},
	{`\`, 0, 1},
	{`\?`, -1, 0},
	{"", -1, 0},
}

func TestIndex(t *testing.T) {
	for _, tt := range indexTests {
		switch i, w := indexSep(tt.pattern); {
		case i != tt.i:
			t.Errorf("expected %v, got %v", tt.i, i)
		case w != tt.w:
			t.Errorf("expected %v, got %v", tt.w, w)
		}
	}
}

var splitTests = []struct {
	pattern    string
	base, path string
}{
	// abs
	{`C:\\Windows\\Temp`, `C:\`, `Windows\\Temp`},
	{`C:/Windows/Temp`, `C:/`, `Windows/Temp`},
	{`C:\/Windows\/Temp`, `C:/`, `Windows\/Temp`},
	{`\\Windows\\Temp`, `\`, `Windows\\Temp`},
	{`/Windows/Temp`, `/`, `Windows/Temp`},
	{`\/Windows\/Temp`, `/`, `Windows\/Temp`},
	// rel
	{`Program Files\\Windows Defender`, `.`, `Program Files\\Windows Defender`},
	{`Program Files/Windows Defender`, `.`, `Program Files/Windows Defender`},
	{`Program Files\/Windows Defender`, `.`, `Program Files\/Windows Defender`},
	{`C:Program Files\\Windows Defender`, `C:`, `Program Files\\Windows Defender`},
	{`C:Program Files/Windows Defender`, `C:`, `Program Files/Windows Defender`},
	{`C:Program Files\/Windows Defender`, `C:`, `Program Files\/Windows Defender`},
	{`.\\Program Files\\Windows Defender`, `.`, `.\\Program Files\\Windows Defender`},
	{`./Program Files/Windows Defender`, `.`, `./Program Files/Windows Defender`},
	{`.\/Program Files\/Windows Defender`, `.`, `.\/Program Files\/Windows Defender`},
	// unc
	{`\\\\Server\\Share`, `\\Server\Share`, ""},
	{`//Server/Share`, `//Server/Share`, ""},
	{`\/\/Server\/Share`, `//Server/Share`, ""},
	{`\\\\Server\\Share\\Folder\\File`, `\\Server\Share\`, `Folder\\File`},
	{`//Server/Share/Folder/File`, `//Server/Share/`, `Folder/File`},
	{`\/\/Server\/Share\/Folder\/File`, `//Server/Share/`, `Folder\/File`},
	// dev
	{`\\\\.\\C:\\Windows\\Temp`, `\\.\C:\`, `Windows\\Temp`},
	{`//./C:/Windows/Temp`, `//./C:/`, `Windows/Temp`},
	{`\/\/.\/C:\/Windows\/Temp`, `//./C:/`, `Windows\/Temp`},
	{`\\\\\?\\C:\\Windows\\Temp`, `\\?\C:\`, `Windows\\Temp`},
	{`//\?/C:/Windows/Temp`, `//?/C:/`, `Windows/Temp`},
	{`\/\/\?\/C:\/Windows\/Temp`, `//?/C:/`, `Windows\/Temp`},

	{`\\\\.\\UNC\\Server\\Share`, `\\.\UNC\Server\Share`, ""},
	{`//./UNC/Server/Share`, `//./UNC/Server/Share`, ""},
	{`\/\/.\/UNC\/Server\/Share`, `//./UNC/Server/Share`, ""},
	{`\\\\\?\\UNC\\Server\\Share`, `\\?\UNC\Server\Share`, ""},
	{`//\?/UNC/Server/Share`, `//?/UNC/Server/Share`, ""},
	{`\/\/\?\/UNC\/Server\/Share`, `//?/UNC/Server/Share`, ""},

	{`\\\\.\\UNC\\Server\\Share\\Folder\\File`, `\\.\UNC\Server\Share\`, `Folder\\File`},
	{`//./UNC/Server/Share/Folder/File`, `//./UNC/Server/Share/`, `Folder/File`},
	{`\/\/.\/UNC\/Server\/Share\/Folder\/File`, `//./UNC/Server/Share/`, `Folder\/File`},
	{`\\\\?\\UNC\\Server\\Share\\Folder\\File`, `\\?\UNC\Server\Share\`, `Folder\\File`},
	{`//\?/UNC/Server/Share/Folder/File`, `//?/UNC/Server/Share/`, `Folder/File`},
	{`\/\/\?\/UNC\/Server\/Share\/Folder\/File`, `//?/UNC/Server/Share/`, `Folder\/File`},
	// error
	{`\\\\`, `\`, `\\`},
	{`\\\\Server`, `\`, `\\Server`},
	{`\\\\.\\`, `\\.\`, ""},
	{`\\\\.\\UNC`, `\\.\UNC`, ""},
	{`\\\\.\\UNC\\Server`, `\\.\UNC\Server`, ""},
}

func TestSplit(t *testing.T) {
	for _, tt := range splitTests {
		switch base, path := split(tt.pattern); {
		case base != tt.base:
			t.Errorf("expected %q, got %q", tt.base, base)
		case path != tt.path:
			t.Errorf("expected %q, got %q", tt.path, path)
		}
	}
}
