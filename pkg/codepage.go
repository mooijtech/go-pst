// go-pst is a library for reading Personal Storage Table (.pst) files (written in Go/Golang).
//
// Copyright 2023 Marten Mooij
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pst

// CodePageIdentifierToEncoding maps the code page identifier to the IANA encoding for decoding String8 (PropertyReader).
// References https://docs.microsoft.com/en-us/windows/win32/intl/code-page-identifiers
var CodePageIdentifierToEncoding = map[int]string{
	28596: "iso-8859-6",
	1256:  "windows-1256",
	28594: "iso-8859-4",
	1257:  "windows-1257",
	28592: "iso-8859-2",
	1250:  "windows-1250",
	936:   "gb2312",
	52936: "hz-gb-2312",
	54936: "gb18030",
	950:   "big5",
	28595: "iso-8859-5",
	20866: "koi8-r",
	21866: "koi8-u",
	1251:  "windows-1251",
	28597: "iso-8859-7",
	1253:  "windows-1253",
	38598: "iso-8859-8-i",
	1255:  "windows-1255",
	51932: "euc-jp",
	50220: "iso-2022-jp",
	50221: "csISO2022JP",
	932:   "iso-2022-jp",
	949:   "ks_c_5601-1987",
	51949: "euc-kr",
	28593: "iso-8859-3",
	28605: "iso-8859-15",
	874:   "windows-874",
	28599: "iso-8859-9",
	1254:  "windows-1254",
	65000: "utf-7",
	65001: "utf-8",
	20127: "us-ascii",
	1258:  "windows-1258",
	28591: "iso-8859-1",
	1252:  "Windows-1252",
}
