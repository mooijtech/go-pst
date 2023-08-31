package writer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	pst "github.com/mooijtech/go-pst/v6/pkg"
	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/proto"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// PropertiesWriter represents a writer for properties.
type PropertiesWriter struct {
	// Properties represents the properties to write.
	Properties proto.Message
}

// NewPropertiesWriter creates a new PropertiesWriter.
func NewPropertiesWriter(properties proto.Message) *PropertiesWriter {
	return &PropertiesWriter{
		Properties: properties,
	}
}

// Property represents a property that can be written.
type Property struct {
	ID    pst.Identifier
	Type  pst.PropertyType
	Value bytes.Buffer
}

// GetProperties returns a list of properties to write.
func (propertiesWriter *PropertiesWriter) GetProperties() ([]Property, error) {
	var properties []Property

	propertyTypes := reflect.TypeOf(propertiesWriter.Properties).Elem()
	propertyValues := reflect.ValueOf(propertiesWriter.Properties).Elem()

	for i := 0; i < propertyTypes.NumField(); i++ {
		if !propertyTypes.Field(i).IsExported() {
			continue
		}
		if propertyValues.Field(i).IsNil() {
			continue
		}

		tag := strings.ReplaceAll(propertyTypes.Field(i).Tag.Get("msg"), ",omitempty", "")

		if tag == "" {
			fmt.Printf("Skipping property without tag: %s\n", propertyTypes.Field(i).Name)
			continue
		}

		propertyID, err := strconv.Atoi(strings.Split(tag, "-")[0])

		if err != nil {
			return nil, eris.Wrap(err, "failed to convert propertyID to int")
		}

		propertyType, err := strconv.Atoi(strings.Split(tag, "-")[1])

		if err != nil {
			return nil, eris.Wrap(err, "failed to convert propertyType to int")
		}

		var propertyBuffer bytes.Buffer

		switch propertyValue := propertyValues.Field(i).Elem().Interface().(type) {
		case string:
			// Binary is intended for fixed-size structures with obvious encodings.
			// Strings are not fixed size and do not have an obvious encoding.
			if _, err := io.WriteString(&propertyBuffer, propertyValue); err != nil {
				return nil, eris.Wrap(err, "failed to write string")
			}
		default:
			if err := binary.Write(&propertyBuffer, binary.LittleEndian, propertyValue); err != nil {
				return nil, eris.Wrap(err, "failed to write property")
			}
		}

		properties = append(properties, Property{
			ID:    pst.Identifier(propertyID),
			Type:  pst.PropertyType(propertyType),
			Value: propertyBuffer,
		})
	}

	return properties, nil
}
