<h1 align="center">
  <br>
  <a href="https://github.com/mooijtech/go-pst"><img src="https://i.imgur.com/LIicreP.png" alt="go-pst" width="280"></a>
  <br>
  go-pst
  <br>
</h1>

<h4 align="center">A library for reading PST files (written in Go/Golang)</h4>

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

**Please also ensure that you have a folder with the name - attachments**

```bash
$ go install github.com/mooijtech/go-pst
```
```Go
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	pst "github.com/mooijtech/go-pst/pkg"
	"github.com/mooijtech/go-pst/pkg/properties"

	"github.com/rotisserie/eris"
)

func createEmailOutput(msg *pst.Message) string {
	result := ""
	switch msgProperties := msg.Properties.(type) {
	case *properties.Appointment:
		fmt.Printf("Appointment: %s - %s\n", msgProperties.String(), "denis")
		result += msgProperties.String()
	case *properties.Contact:
		fmt.Printf("Contact: %s - %s\n", msgProperties.String(), "denis")
		result += msgProperties.String()
	case *properties.Task:
		fmt.Printf("Task: %s\n", msgProperties.GetTaskAssigner())
		result += msgProperties.GetTaskAssigner()
	case *properties.RSS:
		fmt.Printf("RSS: %s\n", msgProperties.GetPostRssChannelLink())
		result += msgProperties.GetPostRssChannelLink()
	case *properties.AddressBook:
		fmt.Printf("Address book: %s\n", msgProperties.GetAccount())
		result += msgProperties.GetAccount()
	case *properties.Message:
		// fmt.Printf("msg: %s\n", msgProperties.GetSubject())
		result += "==========================================\n"
		result += "Subject:" + msgProperties.GetSubject() + "\n\n" + msgProperties.GetBody()
		result += "==========================================\n"
	case *properties.Note:
		fmt.Printf("Note: %d\n", msgProperties.GetNoteColor())
		result += string(msgProperties.GetNoteColor())
	default:
		fmt.Printf("Unknown message type\n")
	}
	return result
}

func main() {
	startTime := time.Now()
	fmt.Println("Initializing...")

	// Open the file in append mode and create it if it doesn't exist
	output, err := os.OpenFile("output.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()

	// Create a writer
	writer := bufio.NewWriter(output)

	// Provide the path to your PST container
	file, err := os.Open("./data/denis.pst")
	if err != nil {
		panic(fmt.Sprintf("Failed to open PST file: %+v\n", err))
	}

	defer func() {
		if errClosing := file.Close(); errClosing != nil {
			panic(fmt.Sprintf("Failed to close PST file: %+v\n", errClosing))
		}
	}()
	
	pstFile, err := pst.New(file)
	if err != nil {
		panic(fmt.Sprintf("Failed to open PST file: %+v\n", err))
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
			// Write the string to the file with a new line
			_, err = writer.WriteString(fmt.Sprintf("Message: %+v \n\n\n", createEmailOutput(message)))
			if err != nil {
				log.Fatal(err)
			}

			attachmentIterator, err := message.GetAttachmentIterator()

			if eris.Is(err, pst.ErrAttachmentsNotFound) {
				// This message has no attachments.
				continue
			} else if err != nil {
				return err
			}
			attachID := 1
			// Iterate through attachments.
			for attachmentIterator.Next() {
				attachment := attachmentIterator.Value()

				attachmentOutput, err := os.Create(fmt.Sprintf("attachments/%s", attachment.GetAttachLongFilename()))

				if err != nil {
					log.Fatal("FATAL ERROR\n")
					return err
				} else if _, err := attachment.WriteTo(attachmentOutput); err != nil {
					return err
				}

				if err := attachmentOutput.Close(); err != nil {
					return err
				}
				_, err = writer.WriteString(fmt.Sprintf("Attachement #%d: %+v \n\n", attachID, attachment.GetAttachFilename()))
				if err != nil {
					log.Fatal(err)
				}
				attachID += 1
			}

			if attachmentIterator.Err() != nil {
				return attachmentIterator.Err()
			}
		}

		// Flush the writer to ensure the data is written to the file
		err = writer.Flush()
		if err != nil {
			log.Fatal(err)
		}
		return messageIterator.Err()
	}); err != nil {
		panic(fmt.Sprintf("Failed to walk folders: %+v\n", err))
	}
	fmt.Printf("Time: %s\n", time.Since(startTime).String())
}
```
## Result
As an example of the result you will get a file "output.txt" with the following structure:
---
<div style="font-family: Calibri, Arial, sans-serif;">
  <strong>Message:</strong>
  <hr>
  
  <strong>Subject:</strong> RE: Начало!
  
  Всех с праздником, мой брат вчера позвал нас всех отпраздновать это на футбольном поле! Те, кто не в курсе правил, прилагаю учебник, который он мне прислал.
  
  <br>
  
  <strong>From:</strong> marteng@forensics.ru &lt;marten@forensics.ru&gt;<br>
  <strong>Sent:</strong> Thursday, March 2, 2023 4:54 PM<br>
  <strong>To:</strong> andrey@forensics.ru; denis@forensics.ru; roman@forensics.ru; timur@forensics.ru; marten@forensics.ru
  
  <hr>
  
  <strong>Subject:</strong> RE: Начало!
  
  Спасибо!
  
  Всех с открытием!
  
  <br>
  
  <strong>From:</strong> andrey@forensics.ru &lt;andrey@forensics.ru&gt;<br>
  <strong>Sent:</strong> 2 марта 2023 г. 16:50<br>
  <strong>To:</strong> denis@forensics.ru; oleg@forensics.ru; roman@forensics.ru; timur@forensics.ru; marten@forensics.ru
  
  <hr>
  
  <strong>Subject:</strong> Начало!
  
  Коллеги, добрый день!
  
  Поздравляю вас с началом новой эпохи, открытием "Forensics"!
  
  Нас ждет великий успех!
  
  <hr>
  
  <strong>Attachments:</strong>
  <ul>
    <li>image001.jpg</li>
    <li>Учебник-.pdf</li>
  </ul>
</div>

---

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
