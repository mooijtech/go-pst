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
      <img src="https://img.shields.io/badge/license-AGPLv3-blue.svg?style=flat-square">
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

```bash
$ go install github.com/mooijtech/go-pst/v5
```

```go
package main

import (
  "fmt"
  "github.com/pkg/errors"
  "os"
  "time"

  pst "github.com/mooijtech/go-pst/v5/pkg"
)

func main() {
  startTime := time.Now()

  fmt.Println("Initializing...")

  pstFile, err := pst.NewFromFile("../data/enron.pst")

  if err != nil {
    panic(fmt.Sprintf("Failed to open PST file: %+v\n", err))
  }

  defer func() {
    if errClosing := pstFile.Close(); errClosing != nil {
      panic(fmt.Sprintf("Failed to close PST file: %+v\n", err))
    }
  }()

  // Walk through folders.
  if err := pstFile.WalkFolders(func(folder pst.Folder) error {
    fmt.Printf("Walking folder: %s\n", folder.Name)

    messageIterator, err := folder.GetMessageIterator()

    if errors.Is(err, pst.ErrMessagesNotFound) {
      // Folder has no messages.
      return nil
    } else if err != nil {
      return err
    }

    // Iterate through messages.
    for messageIterator.Next() {
      message := messageIterator.Value()

      fmt.Printf("Message subject: %s\n", message.GetSubject())

      attachmentIterator, err := message.GetAttachmentIterator()

      if errors.Is(err, pst.ErrAttachmentsNotFound) {
        // This message has no attachments.
        continue
      } else if err != nil {
        return err
      }

      // Iterate through attachments.
      for attachmentIterator.Next() {
        attachment := attachmentIterator.Value()

        fmt.Printf("Attachment: %s\n", attachment.GetAttachFilename())

        attachmentOutput, err := os.Create(fmt.Sprintf("attachments/%s", attachment.GetAttachFilename()))

        if err != nil {
          return err
        } else if _, err := attachment.WriteTo(attachmentOutput); err != nil {
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

go-pst is open-source under the GNU Affero General Public License Version 3 - AGPLv3. Fundamentally, this means that you are free to use go-pst for your project, as long as you don't modify go-pst. If you do, you have to make the modifications public.

## Contact

Feel free to contact me if you have any questions.<br/>
**Name**: Marten Mooij<br/>
**Email**: info@mooijtech.com<br/>

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

## Datasets

This library is tested on the following datasets:

- [enron.pst](https://github.com/mooijtech/go-pst/blob/master/data/enron.pst)
  - [Enron Corporation](https://en.wikipedia.org/wiki/Enron)
- [32-bit.pst](https://github.com/mooijtech/go-pst/blob/master/data/32-bit.pst)
  - [DFRWS 2009 Rodeo](http://old.dfrws.org/2009/rodeo.shtml)
- [support.pst](https://github.com/mooijtech/go-pst/blob/master/data/support.pst)
  - [Hacking Team](https://en.wikipedia.org/wiki/Hacking_Team)
  - ~200GB worth of PST files from Hacking Team is available via this torrent magnet link (see the folders mail, mail2, mail3):
    ```
    magnet:?xt=urn:btih:51603bff88e0a1b3bad3962614978929c9d26955&dn=Hacked%20Team&tr=udp%3A%2F%2Fcoppersurfer.tk%3A6969%2Fannounce&tr=udp%3A%2F%2F9.rarbg.me%3A2710%2Fannounce&tr=http%3A%2F%2Fmgtracker.org%3A2710%2Fannounce&tr=http%3A%2F%2Fbt.careland.com.cn%3A6969%2Fannounce&tr=udp%3A%2F%2Fopen.demonii.com%3A1337&tr=udp%3A%2F%2Fexodus.desync.com%3A6969&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.pomf.se&tr=udp%3A%2F%2Ftracker.blackunicorn.xyz%3A6969
    ```
