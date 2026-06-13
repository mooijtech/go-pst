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

package pst_test

import (
	"os"
	"strings"
	"testing"

	pst "github.com/mooijtech/go-pst/v6/pkg"
)

// TestANSIStringProperties guards the ANSI (PT_STRING8) string-extraction fix.
//
// The 32-bit (ANSI) sample stores text properties as PropertyTypeString8 rather
// than the UTF-16 PropertyTypeString used by Unicode PSTs. Two defects used to
// make every ANSI string come back empty:
//
//  1. WalkFolders aborted on the first folder because the folder name
//     (PidTagDisplayName) was read with GetString, which rejects String8.
//  2. WriteMessagePackValue wrote String8 values under a "<ID>30" message-pack
//     key, but the generated property structs key text fields with the Unicode
//     string type 31 (e.g. Subject is msg:"5531"), so the values never decoded
//     into the struct.
//
// This test reads the sample's one item and asserts that a representative
// PT_STRING8 property (PidTagSubject, 0x0037) decodes to the expected text.
func TestANSIStringProperties(t *testing.T) {
	reader, err := os.Open("../data/32-bit.pst")
	if err != nil {
		t.Fatalf("failed to open ANSI sample: %v", err)
	}
	defer reader.Close()

	pstFile, err := pst.New(reader)
	if err != nil {
		t.Fatalf("failed to open PST file: %v", err)
	}
	defer pstFile.Cleanup()

	if pstFile.FormatType != pst.FormatTypeANSI {
		t.Fatalf("sample is not ANSI: FormatType=%d", pstFile.FormatType)
	}

	var subjects []string

	err = pstFile.WalkFolders(func(folder *pst.Folder) error {
		// The folder name itself is a PT_STRING8 property on ANSI; reaching any
		// sub-folder at all proves defect (1) is fixed.
		if folder.MessageCount == 0 {
			return nil
		}
		iterator, err := folder.GetMessageIterator()
		if err != nil {
			return err
		}
		for iterator.Next() {
			message := iterator.Value()
			// PidTagSubject (0x0037) is stored as PT_STRING8 in this sample.
			reader, err := message.PropertyContext.GetPropertyReader(0x0037, message.LocalDescriptors)
			if err != nil {
				continue // not every item carries a subject
			}
			subject, err := reader.GetStringValue()
			if err != nil {
				t.Errorf("GetStringValue on PidTagSubject failed: %v", err)
				continue
			}
			subjects = append(subjects, subject)
		}
		return iterator.Err()
	})
	if err != nil {
		t.Fatalf("WalkFolders failed: %v", err)
	}

	if len(subjects) == 0 {
		t.Fatal("no subjects read from ANSI sample — String8 extraction is broken")
	}

	// The sample's item carries this subject (with a leading PidTagSubject
	// prefix marker that callers strip); assert the decoded text is present.
	const want = "Updated: Olympus training for new hires"
	found := false
	for _, s := range subjects {
		if strings.Contains(s, want) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected a subject containing %q, got %q", want, subjects)
	}
}
