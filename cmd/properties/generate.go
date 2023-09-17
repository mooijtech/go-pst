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

package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	_ "embed"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// main starts the Protocol Buffers generation.
func main() {
	// Download [MS-OXPROPS].
	downloadedFile, err := download()

	if err != nil {
		panic(fmt.Sprintf("Failed to download [MS-OXPROPS].docx: %+v", err))
	}

	defer func() {
		if err = os.Remove(downloadedFile); err != nil {
			panic(fmt.Sprintf("Failed to remove [MS-OXPROPS].docx: %+v", err))
		}
	}()

	// TODO - Verify hash.

	// Unzip as .DOCX is actually just zipped XML.
	unzippedFile, err := unzip(downloadedFile)

	if err != nil {
		panic(fmt.Sprintf("Failed to unzip [MS-OXPROPS].docx: %+v", err))
	}

	defer func() {
		if err = os.Remove(unzippedFile); err != nil {
			panic(fmt.Sprintf("Failed to remove unzipped file [MS-OXPROPS].xml: %+v", err))
		}
	}()

	// Get all defined properties.
	properties, err := getPropertiesFromXML(unzippedFile)

	if err != nil {
		panic(fmt.Sprintf("Failed to create properties: %+v", errors.WithStack(err)))
	}

	// Create .proto files from the defined properties.
	if err = generateProtocolBuffers(properties); err != nil {
		panic(fmt.Sprintf("Failed to generate Protocol Buffers: %+v", errors.WithStack(err)))
	}

	// Create a property map used by go-pst to map values from the Property Context to their Property Type.
	// Write property map.
	propertyMap, err := os.Create("properties.csv")

	if err != nil {
		panic(fmt.Sprintf("Failed to create output file: %+v", err))
	}

	propertyMapWriter := csv.NewWriter(propertyMap)

	for _, property := range properties {
		propertyID := strconv.Itoa(int(property.PropertyID))
		jsonName := getFormattedPropertyName(property.Name)
		propertySet := property.PropertySet

		if propertyID == "0" {
			fmt.Printf("Unknown property ID for: %s\n", jsonName)
			continue
		}

		if err := propertyMapWriter.Write([]string{propertyID, jsonName, propertySet}); err != nil {
			panic(errors.WithStack(err))
		} else if propertyMapWriter.Error() != nil {
			panic(errors.WithStack(propertyMapWriter.Error()))
		}
	}

	propertyMapWriter.Flush() // Important since writes are buffered.

	if propertyMapWriter.Error() != nil {
		panic(fmt.Sprintf("Failed to write property map: %+v", propertyMapWriter.Error()))
	} else if err := propertyMap.Close(); err != nil {
		panic(fmt.Sprintf("Failed to close property map: %+v", err))
	}

	// TODO - Validate properties exist, parse DOCX in the same way:
	// TODO - [MS-OXCMSG]: Message and Attachment Object Protocol
	// TODO - [MS-OXCFOLD]: Folder Object Protocol and
	// TODO - See referenceToProtocolBufferMessageType for mappings.
	// TODO - Extend properties.Folder to include everything in [MS-OXCFOLD]: Folder Object Protocol.
	//for downloadName, downloadURL := referenceToProtocolBufferMessageType {
	// TODO - Download and apply same parser logic.
	//}
}

// download the latest version of [MS-OXPROPS].docx and returns the downloaded file.
func download() (destinationFilename string, err error) {
	fmt.Println("Connecting to [MS-OXPROPS] (Version 27.0)...")

	httpClientWithTimeout := &http.Client{
		Timeout: 60 * time.Second,
	}

	response, err := httpClientWithTimeout.Get("https://msopenspecs.azureedge.net/files/MS-OXPROPS/%5bMS-OXPROPS%5d-210817.docx")

	if err != nil {
		return "", errors.WithStack(err)
	}

	defer func() {
		if errClosing := response.Body.Close(); errClosing != nil {
			err = errors.WithStack(errClosing)
		}
	}()

	destinationFile, err := os.Create("[MS-OXPROPS].docx")

	if err != nil {
		return "", errors.WithStack(err)
	}

	defer func() {
		if errClosing := destinationFile.Close(); errClosing != nil {
			err = errors.WithStack(errClosing)
		}
	}()

	fmt.Println("Downloading [MS-OXPROPS].docx...")

	if _, err = io.Copy(destinationFile, response.Body); err != nil {
		return "", errors.WithStack(err)
	}

	return destinationFile.Name(), nil
}

// unzip the .docx file.
func unzip(sourceFile string) (destinationFilename string, err error) {
	fmt.Println("Unzipping...")

	zipReader, err := zip.OpenReader(sourceFile)

	if err != nil {
		return "", errors.WithStack(err)
	}

	defer func() {
		if errClosing := zipReader.Close(); errClosing != nil {
			err = errors.WithStack(errClosing)
		}
	}()

	destinationFile, err := os.Create("[MS-OXPROPS].xml")

	if err != nil {
		return "", errors.WithStack(err)
	}

	defer func() {
		if errClosing := destinationFile.Close(); err != nil {
			err = errors.WithStack(errClosing)
		}
	}()

	for _, zipFile := range zipReader.File {
		if zipFile.Name == "word/document.xml" {
			zipFileReader, err := zipFile.Open()

			if err != nil {
				return "", err
			}

			copyFile := func() (err error) {
				defer func() {
					if errClosing := zipFileReader.Close(); err != nil {
						err = errors.WithStack(errClosing)
					}
				}()

				// Limit to 10MiB to be safe (for security linter).
				if _, err := io.CopyN(destinationFile, zipFileReader, int64(10*(1024*1024))); !errors.Is(err, io.EOF) {
					if err == nil {
						return errors.New("expected EOF, XML file size passed the limit")
					}

					return err
				}

				return nil
			}

			return destinationFile.Name(), copyFile()
		}
	}

	return "", errors.New("failed to find word/document.xml in [MS-OXPROPS].docx")
}

// xmlProperty represents the parsed property from XML.
type xmlProperty struct {
	Name                string   // Canonical name.
	PropertySet         string   // GUID which maps to [MS-OXPROPS]: 1.3.2 Commonly Used Property Sets.
	PropertyID          uint32   // Identifier.
	PropertyTypes       []uint16 // Data type.
	AreaName            string   // Functional area which maps to [MS-OXPROPS]: 1.3.1 Common Message Classes.
	DefiningReference   string   // Specification reference.
	PropertyDescription string   // For documentation.
}

// getPropertiesFromXML parses all properties from the [MS-OXPROPS] XML file.
func getPropertiesFromXML(sourceXMLFile string) (xmlProperties []xmlProperty, err error) {
	fmt.Println("Parsing XML to properties...")

	xmlFile, err := os.Open(sourceXMLFile)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer func() {
		if closeErr := xmlFile.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	xmlPlaintextLines, err := getXMLPlaintextLines(xmlFile)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	xmlLineReader := bufio.NewScanner(bytes.NewReader(xmlPlaintextLines))
	isAtPropertyStructures := false

	properties := new([]xmlProperty)
	property := new(xmlProperty)

	for xmlLineReader.Scan() {
		line := bytes.TrimSpace(xmlLineReader.Bytes()) // Trimming space is needed because the Property ID line can randomly have a trailing space.

		if bytes.Equal(line, []byte("Structures")) {
			isAtPropertyStructures = true
			continue
		} else if bytes.Equal(line, []byte("Structure Examples")) {
			// End of property structures.
			break
		}

		if isAtPropertyStructures {
			// Update the values of the property.
			if err = createProperties(line, properties, property); err != nil {
				return nil, errors.WithStack(err)
			}
		}
	}

	if err := xmlLineReader.Err(); err != nil {
		return nil, errors.WithStack(err)
	}

	// Since we only add the previous property, don't forget the last one.
	*properties = append(*properties, *property)

	if err = validateProperties(*properties); err != nil {
		return nil, errors.WithStack(err)
	}

	return *properties, nil
}

// getXMLPlaintextLines converts XML to plaintext.
func getXMLPlaintextLines(xmlFile io.Reader) ([]byte, error) {
	xmlDecoder := xml.NewDecoder(xmlFile)

	var xmlPlaintextLines []byte

	for {
		xmlToken, err := xmlDecoder.Token()

		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, errors.WithStack(err)
		}

		switch xmlTokenType := xmlToken.(type) {
		case xml.StartElement:
			switch xmlTokenType.Name.Local {
			case "p":
				// We want to keep actual newlines.
				xmlPlaintextLines = append(xmlPlaintextLines, byte('\n'))
			case "instrText":
				// Text invisible to the user.
				if err := xmlDecoder.Skip(); err != nil {
					return nil, errors.WithStack(err)
				}
			}
		case xml.CharData:
			xmlPlaintextLines = append(xmlPlaintextLines, xmlTokenType...)
		}
	}

	return xmlPlaintextLines, nil
}

// createProperties updates the values of the property.
func createProperties(line []byte, properties *[]xmlProperty, property *xmlProperty) error {
	// Parse the property fields.
	switch {
	case bytes.HasPrefix(line, []byte("Canonical name: ")):
		if property.Name != "" {
			// Done with this property.
			*properties = append(*properties, *property)
			*property = xmlProperty{}
		}

		property.Name = string(bytes.TrimPrefix(line, []byte("Canonical name: ")))
	case bytes.HasPrefix(line, []byte("Property set: ")):
		guidRegex := regexp.MustCompile("[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}")
		guid := guidRegex.Find(line)

		if guid == nil {
			// No match.
			return errors.New("failed to find GUID")
		}

		property.PropertySet = string(guid)
	case bytes.HasPrefix(line, []byte("Property ID: ")) || bytes.HasPrefix(line, []byte("Property long ID (LID): ")):
		propertyID, err := getPropertyIDFromLine(line, property.Name)

		if err != nil {
			return err
		}

		property.PropertyID = propertyID
	case bytes.HasPrefix(line, []byte("Data type: ")):
		dataTypeRegex := regexp.MustCompile("0x[0-9a-fA-F]{4}")
		dataTypeRegexMatches := dataTypeRegex.FindAll(line, 10)

		if dataTypeRegexMatches == nil {
			return errors.New("failed to find matches for data types regex")
		}

		for _, dataTypeMatch := range dataTypeRegexMatches {
			dataType, err := strconv.ParseUint(string(bytes.TrimPrefix(dataTypeMatch, []byte("0x"))), 16, 16)

			if err != nil {
				return errors.WithStack(fmt.Errorf("failed to parse property type hex to int: %w", err))
			}

			property.PropertyTypes = append(property.PropertyTypes, uint16(dataType))
		}
	case bytes.HasPrefix(line, []byte("Area: ")):
		property.AreaName = string(bytes.TrimPrefix(line, []byte("Area: ")))
	case bytes.HasPrefix(line, []byte("Description: ")):
		property.PropertyDescription = string(bytes.TrimPrefix(line, []byte("Description: ")))
	}

	return nil
}

// getPropertyIDFromLine returns the ID of the property.
func getPropertyIDFromLine(line []byte, propertyName string) (uint32, error) {
	propertyIDHex := bytes.TrimPrefix(line, []byte("Property ID: "))
	propertyIDHex = bytes.TrimPrefix(propertyIDHex, []byte("Property long ID (LID): "))
	propertyIDHex = bytes.TrimPrefix(propertyIDHex, []byte("0x"))

	var propertyIDBitSize int

	switch {
	case strings.HasPrefix(propertyName, "PidLid"):
		propertyIDBitSize = 32
	case strings.HasPrefix(propertyName, "PidTag"):
		propertyIDBitSize = 16
	default:
		return 0, errors.WithStack(fmt.Errorf("failed to find property ID prefix for: %s", propertyName))
	}

	propertyID, err := strconv.ParseUint(string(propertyIDHex), propertyIDBitSize, propertyIDBitSize)

	if err != nil {
		return 0, errors.WithStack(fmt.Errorf("failed to parse property ID hex to int: %w", err))
	}

	return uint32(propertyID), nil
}

// validateProperties validates the parsed properties by checking if all values are set.
func validateProperties(properties []xmlProperty) error {
	for _, property := range properties {
		if property.Name == "" {
			return errors.New("unset property name found")
		}
		if property.PropertyID == 0 {
			if strings.HasPrefix(property.Name, "PidName") {
				// Named properties are in the Name-to-ID Map but don't have a property ID.
			} else {
				return errors.New("unset property ID found for: " + property.Name)
			}
		}
		if property.AreaName == "" {
			return errors.New("unset property area found for: " + property.Name)
		}
		if property.PropertyTypes == nil {
			return errors.New("unset property types found for: " + property.Name)
		}

		for _, propertyType := range property.PropertyTypes {
			if propertyType == 0 {
				return errors.New("unset property type found for: " + property.Name)
			}
		}
	}

	return nil
}

// Property types are defined in [MS-OXCDATA].
// Mapped to https://developers.google.com/protocol-buffers/docs/proto3#scalar
var propertyTypeToProtocolBufferFieldType = map[uint16]string{
	2:   "int32",
	3:   "int32",
	4:   "float",
	5:   "double",
	6:   "double",
	7:   "int64", // Time
	10:  "uint32",
	11:  "bool",
	20:  "double",
	31:  "string",
	30:  "string",
	64:  "int64", // Time
	72:  "uint64",
	251: "uint32",
	// The rest are variable size.
}

// areaNameToProtocolBufferMessageType defines the mapping from area names to Protocol Buffer Message Types.
// TODO - Layer on top of "Common" area name, check the defining reference then sort accordingly.
var areaNameToProtocolBufferMessageType = map[string]string{
	"Sharing":                             "Sharing",
	"Extracted Entities":                  "Extracted Entity",
	"Unified Messaging":                   "Voicemail",
	"Email":                               "Message",
	"Journal":                             "Journal",
	"Tasks":                               "Task",
	"RSS":                                 "RSS",
	"Meetings":                            "Appointment",
	"Calendar":                            "Appointment",
	"Conferencing":                        "Appointment",
	"Spam":                                "Spam",
	"Reminders":                           "Appointment",
	"General Message Properties":          "Message",
	"Message Properties":                  "Message",
	"Message Time Properties":             "Message",
	"Message Attachment Properties":       "Attachment",
	"Sticky Notes":                        "Note",
	"Secure Messaging Properties":         "Message", // [MS-OXORMMS]: Rights-Managed Email Object Protocol.
	"SMS":                                 "SMS",     // [MS-OXOSMMS]: Short Message Service (SMS) and Multimedia Messaging Service (MMS) Object Protocol.
	"Contact Properties":                  "Contact",
	"MapiMailUser":                        "Contact",
	"Address Book":                        "Address Book", // [MS-OXOABK]: Address Book Object Protocol.
	"MapiAddressBook":                     "Address Book",
	"Free/Busy Properties":                "Appointment",
	"Archive":                             "Message",
	"MapiRecipient":                       "Message",
	"MIME Properties":                     "Message",
	"Address Properties":                  "Message",
	"Common":                              "Message",
	"Access Control Properties":           "Message",
	"Outlook Application":                 "Message",
	"Address book":                        "Address Book",
	"History Properties":                  "Message",
	"Sync":                                "Message",
	"ICS":                                 "Message",
	"Folder Properties":                   "Folder",
	"Conversations":                       "Message",
	"Configuration":                       "Message",
	"Miscellaneous Properties":            "Message",
	"MapiNonTransmittable":                "Message",
	"Logon Properties":                    "Message",
	"Mail":                                "Message",
	"MessageClassDefinedTransmittable":    "Message",
	"MapiMessageStore":                    "Message", // TODO - Message Store?
	"ExchangeMessageStore":                "Message", // TODO - Message Store?
	"Message Store Properties":            "Message",
	"Appointment":                         "Appointment",
	"MapiEnvelope":                        "Message",
	"ExchangeNonTransmittableReserved":    "Message",
	"Exchange Administrative":             "Message",
	"TransportEnvelope":                   "Message",
	"MapiNonTransmittable Property set":   "Message",
	"MapiContainer":                       "Message",
	"MapiAttachment":                      "Attachment",
	"Rules":                               "Message",
	"MapiStatus":                          "Message",
	"Server-Side Rules Properties":        "Message",
	"ID Properties":                       "Message",
	"AB Container":                        "Message",
	"Search":                              "Message",
	"MapiEnvelope Property set":           "Message",
	"Transport Envelope":                  "Message",
	"Common Property set":                 "Message",
	"Conflict Note":                       "Message",
	"Table Properties":                    "Message",
	"MAPI Display Tables":                 "Message",
	"Server":                              "Message",
	"General Report Properties":           "Message",
	"Run-time configuration":              "Message",
	"Meeting Response":                    "Meeting", // TODO - Meeting?
	"Calendar Property set":               "Message", // TODO - Calendar?
	"Site Mailbox":                        "Message",
	"Flagging":                            "Message",
	"Offline Address Book Properties":     "Address Book",
	"MIME properties":                     "Message",
	"BestBody":                            "Message",
	"TransportRecipient":                  "Message",
	"ExchangeAdministrative":              "Message",
	"ProviderDefinedNonTransmittable":     "Message",
	"MapiMessage":                         "Message",
	"MapiCommon":                          "Message",
	"ExchangeFolder":                      "Folder",
	"ExchangeMessageReadOnly":             "Message",
	"Message Class Defined Transmittable": "Message",
}

// referenceToProtocolBufferMessageType maps the defining reference prefix to the message type.
// Preferred over area name mapping.
// TODO - Use these documents directly :)
var referenceToProtocolBufferMessageType = map[string]string{
	// References https://learn.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxomsg/daa9120f-f325-4afb-a738-28f91049ab3c
	"[MS-OXOMSG]": "Message",
	// References https://learn.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxocntc/9b636532-9150-4836-9635-9c9b756c9ccf
	"[MS-OXOCNTC]": "Contact",
	// References https://learn.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxocal/09861fde-c8e4-4028-9346-e7c214cfdba1
	"[MS-OXOCAL]": "Appointment",
	// References https://learn.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxcfold/c0f31b95-c07f-486c-98d9-535ed9705fbf
	"[MS-OXCFOLD]": "Folder",
	// References
	"[MS-OXOABK]": "Address Book Object Protocol",
}

//go:embed header.txt
var protoHeader string

// generateProtocolBuffers creates .proto files from the parsed properties.
func generateProtocolBuffers(properties []xmlProperty) error {
	fmt.Printf("Found %d properties...\n", len(properties))

	protocolBufferMessageTypeToBuilder := make(map[string]*strings.Builder)
	protocolBufferMessageTypeToUniqueFieldID := make(map[string]int)

	// TODO - Validate all area name/reference mappings
	// missingMappings := map[string]struct{}

	// Create the Protocol Buffers.
	for _, property := range properties {
		protocolBufferMessageType := areaNameToProtocolBufferMessageType[property.AreaName]

		if protocolBufferMessageType == "" {
			// We need to map this.
			fmt.Printf("Unmapped area name \"%s\", defining reference: %s\n", property.AreaName, property.DefiningReference)
			fmt.Println("Writing to unknown...")
		} else {
			protocolBufferBuilder := protocolBufferMessageTypeToBuilder[protocolBufferMessageType]

			if protocolBufferBuilder == nil {
				// Initialize the Protocol Buffer builder.
				protocolBufferBuilder = &strings.Builder{}
				protocolBufferMessageTypeToBuilder[protocolBufferMessageType] = protocolBufferBuilder

				// Add .proto file header.
				protocolBufferBuilder.WriteString(protoHeader)
				protocolBufferBuilder.WriteString("\n")

				// Define the Message Type.
				protocolBufferBuilder.WriteString(fmt.Sprintf("message %s {\n", strings.ReplaceAll(protocolBufferMessageType, " ", "")))
			}

			// Create the (property) fields.
			for _, propertyType := range property.PropertyTypes {
				protocolBufferFieldType := propertyTypeToProtocolBufferFieldType[propertyType]
				uniqueFieldID := protocolBufferMessageTypeToUniqueFieldID[protocolBufferMessageType] + 1

				// Increment unique field ID.
				protocolBufferMessageTypeToUniqueFieldID[protocolBufferMessageType] = uniqueFieldID

				// Write the property field.
				protocolBufferBuilder.WriteString(createProtoFileField(property, protocolBufferFieldType, uniqueFieldID))

				break // TODO - Resolve naming conflicts for properties with the same name but multiple value types.
			}
		}
	}

	return createProtoFilesFromBuilders(protocolBufferMessageTypeToBuilder)
}

// createProtoFileField returns the Protocol Buffer field for this property.
func createProtoFileField(property xmlProperty, protocolBufferFieldType string, uniqueFieldID int) string {
	var fieldBuilder strings.Builder

	if protocolBufferFieldType != "" {
		// Format property name.
		formattedPropertyName := getFormattedPropertyName(property.Name)

		// Create Go struct tags so go-pst knows how to map values to fields.
		goTags := strings.Builder{}

		if property.PropertyID != 0 {
			// TODO - There may be multiple property types.
			goTags.WriteString(fmt.Sprintf("// @gotags: msg:\"%d-%d,omitempty\"", property.PropertyID, property.PropertyTypes[0]))
		} // PropertySets are mapped via the PropertyMap in property_context.go

		// Write.
		fieldBuilder.WriteString(fmt.Sprintf("  // %s\n", property.PropertyDescription))
		fieldBuilder.WriteString(
			fmt.Sprintf("  optional %s %s = %d; %s\n",
				protocolBufferFieldType,
				formattedPropertyName,
				uniqueFieldID,
				goTags.String(),
			))
	} else {
		// Variable size/multiple values.
		fmt.Printf("Unmapped variable size/multiple values property type: %s\n", protocolBufferFieldType)
	}

	return fieldBuilder.String()
}

// getFormattedPropertyName returns the formatted property named per https://developers.google.com/protocol-buffers/docs/style#message_and_field_names
func getFormattedPropertyName(propertyName string) string {
	propertyName = strings.TrimPrefix(propertyName, "PidLid")
	propertyName = strings.TrimPrefix(propertyName, "PidTag")
	propertyName = strings.TrimPrefix(propertyName, "PidName")

	// Add an underscore "_" after each capital letter except the first one.
	var protocolBufferPropertyName strings.Builder
	isFirstChar := true

	for i, r := range propertyName {
		switch {
		case isFirstChar:
			isFirstChar = false
			protocolBufferPropertyName.WriteRune(r)
		case unicode.IsUpper(r):
			if len(propertyName) == i+1 || !unicode.IsUpper([]rune(propertyName)[i+1]) {
				protocolBufferPropertyName.WriteString("_")
				protocolBufferPropertyName.WriteRune(r)
			} else {
				protocolBufferPropertyName.WriteRune(r)
			}
		default:
			protocolBufferPropertyName.WriteRune(r)
		}
	}

	return strings.ToLower(protocolBufferPropertyName.String())
}

// createProtoFilesFromBuilders writes the .proto files.
func createProtoFilesFromBuilders(protocolBufferMessageTypeToBuilder map[string]*strings.Builder) error {
	if err := os.MkdirAll("protobufs", 0755); err != nil {
		return errors.WithStack(err)
	}

	// Create the .proto files.
	for key := range protocolBufferMessageTypeToBuilder {
		builder := protocolBufferMessageTypeToBuilder[key]

		if builder.Len() == 0 {
			fmt.Printf("No protocol buffer builder for key: %s\n", key)
			continue
		}

		// Close the ending code of each Protocol Buffer builder.
		builder.WriteString("}\n")

		protoFileName := fmt.Sprintf("protobufs/%s.proto", strings.ToLower(strings.ReplaceAll(key, " ", "_")))

		fmt.Printf("Creating %s...\n", protoFileName)

		protoFile, err := os.Create(protoFileName)

		if err != nil {
			return errors.WithStack(err)
		}

		copyBuilderToFile := func() (err error) {
			defer func() {
				if errClosing := protoFile.Close(); errClosing != nil {
					err = errors.WithStack(err)
				}
			}()

			if _, err := io.Copy(protoFile, strings.NewReader(builder.String())); err != nil {
				return errors.WithStack(err)
			}

			return nil
		}

		if err = copyBuilderToFile(); err != nil {
			return err
		}
	}

	return nil
}
