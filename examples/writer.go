package main

import (
	"flag"
	"fmt"
	"github.com/dustin/go-humanize"
	pst "github.com/mooijtech/go-pst/v6/pkg"
	"github.com/mooijtech/go-pst/v6/pkg/properties"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"os"
	"time"
)

func main() {
	outputName := *flag.String("output", "1337.pst", "Specifies the output path of the PST file.")

	flag.Parse()

	slog.Info("Starting go-pst...")

	startTime := time.Now()

	// Output file.
	outputFile, err := os.Create(outputName)

	if err != nil {
		t.Fatalf("Failed to create output file: %+v", err)
	}

	// TODO - Unsupported Unicode4k (test OST).
	formatType := pst.FormatTypeUnicode
	encryptionType := pst.EncryptionTypePermute

	// Setup Goroutines for the writers.
	writeCancelContext, writeCancelFunc := context.WithCancel(context.Background())
	writeGroup, _ := errgroup.WithContext(writeCancelContext)

	defer writeCancelFunc()

	// Root writer which starts everything.
	writeOptions := NewWriteOptions(formatType, encryptionType)
	writer := NewWriter(outputFile, writeGroup, writeOptions)

	// Root folder which can contain sub-folders with messages and attachments.
	rootFolder := NewFolderWriter(outputFile, writeGroup, formatType)

	if err != nil {
		t.Fatalf("Failed to create folder writer: %+v", err)
	}

	rootFolder.SetIdentifier(pst.IdentifierRootFolder)
	rootFolder.AddProperties(&properties.Folder{
		Name: "IdentifierRootFolder",
	})

	// Add sub-folders.
	for i := 0; i < 6; i++ {
		subFolder := NewFolderWriter(outputFile, writeGroup, formatType)

		subFolder.AddProperties(&properties.Folder{
			Name: fmt.Sprintf("Sub-folder #%d", i),
		})

		// Add messages to the sub-folder.
		message := NewMessageWriter(formatType, writeGroup)

		message.AddProperties(&properties.Message{
			Subject: proto.String("Goodbye, world!"),
		})

		// Add attachments to message.
		for x := 0; x < 9; x++ {
			attachment := NewAttachmentWriter()

			attachment.AddProperties(&properties.Attachment{
				AttachFilename:     proto.String(fmt.Sprintf("nudes%d.png", x)),
				AttachLongFilename: proto.String(fmt.Sprintf("nudes%d.png", x)),
			})

			message.AddAttachments(attachment)
		}

		// Add sub-folders with messages containing attachments to root folder.
		rootFolder.AddFolder(subFolder)
	}

	// Writer which starts everything (has the PST file header).
	writer.AddFolders(rootFolder)

	// WriteTo writes the PST file.
	// WriteTo follows the path to root folder (fixed pst.Identifier, pst.IdentifierRootFolder) then to the pst.TableContext of the root folder.
	// Once there, we can get the child folders ([]pst.Identifier, see FolderWriter), each folder can contain messages (see MessageWriter).
	// Each message uses the pst.BTreeOnHeapHeader to construct a pst.HeapOnNode (this is where the data is).
	//
	// Extending the pst.HeapOnNode (where the data is) we can also use Local Descriptors (extend where this data is):
	// pst.LocalDescriptor (see LocalDescriptorsWriter) are B-Tree nodes pointing to other B-Tree nodes.
	// These local descriptors also have the pst.HeapOnNode structure which can be built upon (explained below).
	// Local Descriptors are used to store more data in the pst.HeapOnNode structure (B-Tree with the nodes containing the data).
	// XBlocks and XXBlocks include an array of []pst.Identifier pointing to B-Tree nodes, it is the format used to store data (see BlockWriter).
	// This structure is used by the Local Descriptors.
	//
	// Each pst.HeapOnNode can contain either a pst.TableContext or pst.PropertyContext:
	// pst.TableContext (see TableContextWriter):
	// The pst.TableContext contains a Row Matrix structure to store data, used by folders (to find data such as the folder identifiers ([]pst.Identifier)).
	// The pst.TableContext is column structured with data exceeding 8 bytes moving to different B-Tree nodes:
	// pst.HeapOnNode which is <= 3580 bytes.
	// pst.LocalDescriptor which is > 3580 bytes.
	// pst.PropertyContext (see PropertyContextWriter):
	// The pst.PropertyContext contains a list of properties ([]pst.Property) of the message, we can write this with PropertyWriter.
	//
	// Combining these structures we make up a PST file to write.
	bytesWritten, err := writer.WriteTo(outputFile)

	if err != nil {
		t.Fatalf("Failed to write PST file: %+v", err)
	}

	// Wait for writers to finish.
	if err := writeGroup.Wait(); err != nil {
		t.Fatalf("Failed to write PST file: %+v", err)
	}

	fmt.Printf("Wrote %s in %s", humanize.Bytes(uint64(bytesWritten)), humanize.Time(time.Now().Add(-time.Since(startTime))))
}
