# go-pst properties generation

## Motivation

PST files contain messages, attachments, contacts, appointments and more.
An individual property is for example Subject/To/From/Body/Headers etc.
All properties are defined in [Exchange Server Protocols Master Property List](https://docs.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxprops/f6ab1613-aefe-447d-a49c-18217230b148).

There are **1078 properties**, I want to be able to:
- Not have to type getters for them all manually
- Read all properties
- Read individual properties
- Sort the properties into their functional areas (like Message, Attachment, Appointment, Contact etc)
- Send the properties over a network
- Use the properties in other languages

## Protocol Buffers

[Protocol Buffers](https://developers.google.com/protocol-buffers/docs/proto3) is a language-neutral, platform-neutral, extensible mechanism for serializing structured data. The protocol buffer language is a language for specifying the schema for structured data. This schema is compiled into language specific bindings.

![Explanation](https://protobuf.dev/images/protocol-buffers-concepts.png)

## Implementation

`generate.go` will:
- Download the latest [[MS-OXPROPS].docx](https://docs.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxprops/f6ab1613-aefe-447d-a49c-18217230b148) containing all property definitions
- Unzip the DOCX file
  - .docx is actually just an archive/zip containing XML files
- Convert the XML file to plaintext
- Parse the plaintext to a list of properties
- Create .proto files from the properties, sorted into their functional areas

## Usage

### Protocol Buffers generation

```bash
# Change directory.
$ cd cmd/properties

# Writes the .proto files to cmd/properties/protobufs.
$ go run generate.go
```

### Go generation

Requires that `generate.go` has been run.

Download the [protocol compiler (protoc)](https://github.com/protocolbuffers/protobuf/releases) and [protoc-gen-go](https://github.com/protocolbuffers/protobuf-go/releases).

```bash
# Compile protoc-go-inject
$ go install github.com/favadi/protoc-go-inject-tag@latest

# Move binary
$ mv ~/go/bin/protoc-go-inject-tag ~/path/to/go-pst
```

```bash
# Change directory.
$ cd /path/to/go-pst

# Generates .go files to pkg/properties.
$ ./protoc --proto_path=cmd/properties/protobufs --go_out=paths=source_relative:pkg/properties --plugin=protoc-gen-go=protoc-gen-go $(find cmd/properties/protobufs -iname "*.proto")

# Generate MessagePack tags.
$ ./protoc-go-inject-tag -input="pkg/properties/*.pb.go" -remove_tag_comment

# Generate Message Pack 
$ cd pkg/properties && go generate
```

## Inspiration

Special thanks to the [work](https://github.com/Jmcleodfoss/ms-oxprops-db) of [James McLeod](https://github.com/Jmcleodfoss).
