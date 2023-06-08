<h1 align="center">
  <br>
  <a href="https://github.com/mooijtech/go-pst"><img src="https://i.imgur.com/LIicreP.png" alt="go-pst" width="280"></a>
  <br>
  go-pst
  <br>
</h1>

<h4 align="center">A library for reading PST files (written in Go/Golang).</h4>

<p align="center">
  <a href="https://github.com/mooijtech/go-pst/blob/master/LICENSE.txt">
      <img src="https://img.shields.io/badge/license-Apache%202-blue.svg?style=flat-square">
  </a>
  <a href="https://github.com/mooijtech/go-pst/issues">
    <img src="https://img.shields.io/github/issues/mooijtech/go-pst.svg?style=flat-square">
  </a>
  <a href="https://github.com/mooijtech/go-pst">
      <img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat-square">
  </a>
</p>

---

[![pkg.go.dev reference](https://img.shields.io/badge/pkg.go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/mooijtech/go-pst/v5)

## Introduction

**go-pst** is a library for reading PST files (written in Go/Golang).

The PFF (Personal Folder File) and OFF (Offline Folder File) format is used to store Microsoft Outlook e-mails, appointments and contacts. The PST (Personal Storage Table), OST (Offline Storage Table) and PAB (Personal Address Book) file format consist of the PFF format.

## Usage

**Requires Go 1.20** for the new `WithCancelCause` added to `context`.

```bash
$ go install github.com/mooijtech/go-pst
```

```go
package main

import (
  "fmt"
  "github.com/mooijtech/go-pst/pkg"
  "github.com/mooijtech/go-pst/pkg/properties"
  "github.com/rotisserie/eris"
  "golang.org/x/text/encoding"
  "os"
  "testing"
  "time"

  charsets "github.com/emersion/go-message/charset"
)

func main() {
  pst.ExtendCharsets(func(name string, enc encoding.Encoding) {
    charsets.RegisterEncoding(name, enc)
  })

  startTime := time.Now()

  fmt.Println("Initializing...")

  reader, err := os.Open("../data/enron.pst")

  if err != nil {
    panic(fmt.Sprintf("Failed to open PST file: %+v\n", err))
  }

  pstFile, err := pst.New(reader)

  if err != nil {
    panic(fmt.Sprintf("Failed to open PST file: %+v\n", err))
  }

  defer func() {
    pstFile.Cleanup()

    if errClosing := reader.Close(); errClosing != nil {
      panic(fmt.Sprintf("Failed to close PST file: %+v\n", err))
    }
  }()

  // Create attachments directory
  if _, err := os.Stat("attachments"); err != nil {
    if err := os.Mkdir("attachments", 0755); err != nil {
      panic(fmt.Sprintf("Failed to create attachments directory: %+v", err))
    }
  }

  // Walk through folders.
  if err := pstFile.WalkFolders(func(folder *pst.Folder) error {
    fmt.Printf("Walking folder: %s\n", folder.Name)

    messageIterator, err := folder.GetMessageIterator()

    if eris.Is(err, pst.ErrMessagesNotFound) {
      // Folder has no messages.
      return nil
    } else if err != nil {
      return err
    }

    // Iterate through messages.
    for messageIterator.Next() {
      message := messageIterator.Value()

      switch messageProperties := message.Properties.(type) {
      case *properties.Appointment:
        //fmt.Printf("Appointment: %s\n", messageProperties.String())
      case *properties.Contact:
        //fmt.Printf("Contact: %s\n", messageProperties.String())
      case *properties.Task:
        //fmt.Printf("Task: %s\n", messageProperties.String())
      case *properties.RSS:
        //fmt.Printf("RSS: %s\n", messageProperties.String())
      case *properties.AddressBook:
        //fmt.Printf("Address book: %s\n", messageProperties.String())
      case *properties.Message:
        fmt.Printf("Subject: %s\n", messageProperties.GetSubject())
      case *properties.Note:
        //fmt.Printf("Note: %s\n", messageProperties.String())
      default:
        fmt.Printf("Unknown message type\n")
      }

      attachmentIterator, err := message.GetAttachmentIterator()

      if eris.Is(err, pst.ErrAttachmentsNotFound) {
        // This message has no attachments.
        continue
      } else if err != nil {
        return err
      }

      // Iterate through attachments.
      for attachmentIterator.Next() {
        attachment := attachmentIterator.Value()

        var attachmentOutputPath string

        if attachment.GetAttachFilename() != "" {
          attachmentOutputPath = fmt.Sprintf("attachments/%d-%s", attachment.Identifier, attachment.GetAttachFilename())
        } else {
          attachmentOutputPath = fmt.Sprintf("attachments/UNKNOWN_ATTACHMENT_FILE_NAME_%d", attachment.Identifier)
        }

        attachmentOutput, err := os.Create(attachmentOutputPath)

        if err != nil {
          return err
        }

        if _, err := attachment.WriteTo(attachmentOutput); err != nil {
          return err
        }

        if err := attachmentOutput.Close(); err != nil {
          return err
        }
      }

      if attachmentIterator.Err() != nil {
        return attachmentIterator.Err()
      }
    }

    return messageIterator.Err()
  }); err != nil {
    panic(fmt.Sprintf("Failed to walk folders: %+v\n", err))
  }

  fmt.Printf("Time: %s\n", time.Since(startTime).String())
}
```

## License 

This project is licensed under the [Apache License 2.0]().

## Documentation

- [Outlook Personal Folders (.pst) File Format](https://github.com/mooijtech/go-pst/blob/master/docs/README.md)
- [Exchange Server Protocols Master Property List](https://learn.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxprops/f6ab1613-aefe-447d-a49c-18217230b148)

## Implementations

- [java-libpst](https://github.com/rjohnsondev/java-libpst)
- [pstreader](https://github.com/Jmcleodfoss/pstreader)
  - Special thanks to [James McLeod](https://github.com/Jmcleodfoss)
- [libpff](https://github.com/libyal/libpff)
- [XstReader](https://github.com/Dijji/XstReader)
- [PANhunt](https://github.com/Dionach/PANhunt/blob/master/pst.py)

## Contact

Feel free to contact me if you have any questions.<br/>
**Name**: Marten Mooij<br/>
**Email**: info@mooijtech.com<br/>