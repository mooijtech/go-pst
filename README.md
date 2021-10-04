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
      <img src="https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square">
  </a>
  <a href="https://github.com/mooijtech/go-pst/issues">
    <img src="https://img.shields.io/github/issues/mooijtech/go-pst.svg?style=flat-square">
  </a>
  <a href="https://github.com/mooijtech/go-pst">
      <img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat-square">
  </a>
</p>

---

## Introduction

**go-pst** is a library for reading PST files (written in Go/Golang).

The PFF (Personal Folder File) and OFF (Offline Folder File) format is used to store Microsoft Outlook e-mails, appointments and contacts. The PST (Personal Storage Table), OST (Offline Storage Table) and PAB (Personal Address Book) file format consist of the PFF format.

## References

Special thanks to [James McLeod](https://github.com/Jmcleodfoss/) for helping me with some debugging.

### Documentation

- [Outlook Personal Folders (.pst) File Format](https://github.com/mooijtech/go-pst/blob/master/docs/MS-PST.pdf)
- [Personal Folder File (PFF) file format specification](https://github.com/mooijtech/go-pst/blob/master/docs/PFF.pdf)

### Libraries

- [java-libpst](https://github.com/rjohnsondev/java-libpst)
- [libpff](https://github.com/libyal/libpff)
- [XstReader](https://github.com/Dijji/XstReader)
- [pstreader](https://github.com/Jmcleodfoss/pstreader)
- [PANhunt](https://github.com/Dionach/PANhunt/blob/master/pst.py)

## Datasets

This library is tested on the following datasets:

- [enron.pst](https://github.com/mooijtech/go-pst/blob/master/data/enron.pst)
  - [Enron Corporation](https://en.wikipedia.org/wiki/Enron)
- [32-bit.pst](https://github.com/mooijtech/go-pst/blob/master/data/32-bit.pst)
  - [DFRWS 2009 Rodeo](http://old.dfrws.org/2009/rodeo.shtml)
- [support.pst](https://github.com/mooijtech/go-pst/blob/master/data/support.pst)
  - [Hacking Team](https://en.wikipedia.org/wiki/Hacking_Team)
  - 50GB worth of PST files from Hacking Team is available via this torrent magnet link (see the folders mail, mail2, mail3): 
    ```
    magnet:?xt=urn:btih:51603bff88e0a1b3bad3962614978929c9d26955&dn=Hacked%20Team&tr=udp%3A%2F%2Fcoppersurfer.tk%3A6969%2Fannounce&tr=udp%3A%2F%2F9.rarbg.me%3A2710%2Fannounce&tr=http%3A%2F%2Fmgtracker.org%3A2710%2Fannounce&tr=http%3A%2F%2Fbt.careland.com.cn%3A6969%2Fannounce&tr=udp%3A%2F%2Fopen.demonii.com%3A1337&tr=udp%3A%2F%2Fexodus.desync.com%3A6969&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.pomf.se&tr=udp%3A%2F%2Ftracker.blackunicorn.xyz%3A6969
    ```

## Usage

```go
package main

import (
  pst "github.com/mooijtech/go-pst/pkg"
  log "github.com/sirupsen/logrus"
)

func main() {
  pstFile := pst.New("data/32-bit.pst")

  log.Infof("Parsing file: %s", pstFile.Filepath)

  isValidSignature, err := pstFile.IsValidSignature()

  if err != nil {
    log.Errorf("Failed to read signature: %s", err)
    return
  }

  if !isValidSignature {
    log.Errorf("Invalid file signature.")
    return
  }

  contentType, err := pstFile.GetContentType()

  if err != nil {
    log.Errorf("Failed to get content type: %s", err)
    return
  }

  log.Infof("Content type: %s", contentType)

  formatType, err := pstFile.GetFormatType()

  if err != nil {
    log.Errorf("Failed to get format type: %s", err)
    return
  }

  log.Infof("Format type: %s", formatType)

  encryptionType, err := pstFile.GetEncryptionType(formatType)

  if err != nil {
    log.Errorf("Failed to get encryption type: %s", err)
    return
  }

  log.Infof("Encryption type: %s", encryptionType)

  rootFolder, err := pstFile.GetRootFolder(formatType, encryptionType)

  if err != nil {
    log.Errorf("Failed to get root folder: %s", err)
    return
  }

  err = GetSubFolders(pstFile, rootFolder, formatType, encryptionType)

  if err != nil {
    log.Errorf("Failed to get sub-folders: %s", err)
    return
  }
}

// GetSubFolders is a recursive function which retrieves all sub-folders for the specified folder.
func GetSubFolders(pstFile pst.File, folder pst.Folder, formatType string, encryptionType string) error {
  subFolders, err := pstFile.GetSubFolders(folder, formatType, encryptionType)

  if err != nil {
    return err
  }

  for _, subFolder := range subFolders {
    log.Infof("Parsing sub-folder: %s", subFolder.DisplayName)

    messages, err := pstFile.GetMessages(subFolder, formatType, encryptionType)

    if err != nil {
      return err
    }

    if len(messages) > 0 {
      log.Infof("Found %d messages.", len(messages))
    }

    err = GetSubFolders(pstFile, subFolder, formatType, encryptionType)

    if err != nil {
      return err
    }
  }

  return nil
}
```

## Implementation

This implementation is based on the [references](#references).<br/>
The source code of go-pst will reference this implementation.

### File Header

| Offset        | Size          | Value                         | Description   |
| ------------- | ------------- | -------------                 | ------------- |
| 0             | 4             | "\x21\x42\x44\x4e" (**!BDN**) | The signature (magic identifier). |
| 8             | 2             |                               | The content type (client signature). See [Content Types](#content-types). |
| 10            | 2             |                               | The data version (NDB version). NDB is short for node database. See [Format Types](#format-types). |

### Content Types

| Value               | Description        |
| -------------       | -------------      |
| "\x53\x4d" (**SM**) | Used for PST files |
| "\x53\x4d" (**SO**) | Used for OST files |
| "\x41\x42" (**AB**) | Used for PAB files |

### Format Types

| Value               | Description        |
| -------------       | -------------      |
| 14                  | 32-bit ANSI format |
| 15                  | 32-bit ANSI format |
| 21                  | 64-bit Unicode format (by Visual Recovery) |
| 23                  | 64-bit Unicode format |
| 36                  | 64-bit Unicode format with 4k |

### The 64-bit header data

| Offset        | Size          | Description   |
| ------------- | ------------- | ------------- |
| 224           | 8             | The node b-tree file offset. |
| 240           | 8             | The block b-tree file offset. |
| 513           | 1             | Encryption type. See [Encryption Types](#encryption-types). |

### The 32-bit header data

| Offset        | Size          | Description   |
| ------------- | ------------- | ------------- |
| 188           | 4             | The node b-tree file offset. |
| 196           | 4             | The block b-tree file offset. |
| 461           | 1             | Encryption type. See [Encryption Types](#encryption-types). |

### Encryption Types

| Value        | Identifier          | Description   |
| -------------| -------------       | ------------- |
| 0x00         | NDB_CRYPT_NONE      | No encryption. |
| 0x01         | NDB_CRYPT_PERMUTE   | Compressible encryption. |
| 0x02         | NDB_CRYPT_CYCLIC    | High encryption. |

### The node and block b-tree

The following offsets start from the (node/block) b-tree offset.

#### 64-bit

| Offset        | Size          | Description   |
| ------------- | ------------- | ------------- |
| 0             |  488            | B-tree node entries (number of entries x entry size). |
| 488           |  1            | The number of entries. |
| 490           |  1            | The size of an entry. |
| 491           |  1            | B-tree node level. A zero value represents a leaf node. A value greater than zero represents a branch node, with the highest level representing the root. |

#### 64-bit 4k

| Offset        | Size          | Description   |
| ------------- | ------------- | ------------- |
| 0             |  4056            | B-tree node entries (number of entries x entry size). |
| 4056           |  2            | The number of entries. |
| 4060           |  1            | The size of an entry. |
| 4061           |  1            | B-tree node level. A zero value represents a leaf node. A value greater than zero represents a branch node, with the highest level representing the root. |

#### 32-bit

| Offset        | Size          | Description   |
| ------------- | ------------- | ------------- |
| 0             |  496            | B-tree node entries (number of entries x entry size). |
| 496           |  1            | The number of entries. |
| 498           |  1            | The size of an entry. |
| 499           |  1            | B-tree node level. A zero value represents a leaf node. A value greater than zero represents a branch node, with the highest level representing the root. |

### The b-tree entries

#### The 64-bit block b-tree branch node entry

| Offset        | Size          | Description   |
| ------------- | ------------- | ------------- |
| 0             |  8            | [The identifier](#identifier) of the first child node. 32-bit integer. |
| 16            |  8            | The file offset, points to a block b-tree branch or leaf node entry. |

#### The 64-bit block b-tree leaf node entry

| Offset        | Size          | Description   |
| ------------- | ------------- | ------------- |
| 0             |  8            | [The identifier](#identifier). 32-bit integer. |
| 8             |  8            | The file offset, points to a [Heap-on-Node](#heap-on-node). |
| 16            |  2            | The size of the [Heap-on-Node](#heap-on-node) (or the [local descriptors](#local-descriptors)). |

#### The 64-bit node b-tree leaf node entry

| Offset        | Size          | Description   |
| ------------- | ------------- | ------------- |
| 0             |  8            | [The identifier.](#identifier) 32-bit integer. |
| 8             |  8            | The node identifier of the data. This node identifier is found in the block b-tree. |
| 16            |  8            | The node identifier of the [local descriptors](#local-descriptors). This node identifier is found in the block b-tree. |

#### The 32-bit block b-tree branch node entry

| Offset        | Size          | Description   |
| ------------- | ------------- | ------------- |
| 0             |  4            | [The identifier](#identifier) of the first child node. 32-bit integer. |
| 8             |  4            | The file offset, points to a block b-tree branch or leaf node entry. |

#### The 32-bit block b-tree leaf node entry

| Offset        | Size          | Description   |
| ------------- | ------------- | ------------- |
| 0             |  4            | [The identifier](#identifier). 32-bit integer. |
| 4             |  4            | The file offset, points to a [Heap-on-Node](#heap-on-node). |
| 8             |  2            | The size of the [Heap-on-Node](#heap-on-node) (or the [local descriptors](#local-descriptors)). |

#### The 32-bit node b-tree leaf node entry

| Offset        | Size          | Description   |
| ------------- | ------------- | ------------- |
| 0             |  4            | [The identifier](#identifier). 32-bit integer. |
| 4             |  4            | The node identifier of the data. This node identifier is found in the block b-tree. |
| 8             |  4            | The node identifier of the [local descriptors](#local-descriptors). This node identifier is found in the block b-tree. |

### Identifier

The 32-bit integer (identifier) can be used to search for b-tree nodes.

| Offset        | Size          | Description   |
| ------------- | ------------- | ------------- |
| 0             |  5 bits       | [The identifier type](#identifier-types). |

#### Identifier types

| Value         | Identifier          | Description   |
| ------------- | ------------- | ------------- |
| 0          |  HID       | Table value (or heap node) |
| 1          |  INTERNAL       | Internal node |
| 2          |  NORMAL_FOLDER       | Folder item |
| 3          |  SEARCH_FOLDER       | Search folder item |
| 4          |  NORMAL_MESSAGE       | Message item |
| 5          |  ATTACHMENT       | Attachment item |
| 6          |  SEARCH_UPDATE_QUEUE       | Queue of changed search folder items |
| 7          |  SEARCH_CRITERIA_OBJECT       | Search folder criteria |
| 8          |  ASSOCIATED_MESSAGE       | Associated contents item |
| 10          |  CONTENTS_TABLE_INDEX       | Unknown |
| 11          |  RECEIVE_FOLDER_TABLE       | Inbox item (or received folder table) |
| 12          |  OUTGOING_QUEUE_TABLE       | Outbox item (or outgoing queue table) |
| 13          |  HIERARCHY_TABLE       | Sub folders item (or hierarchy table) |
| 14          |  CONTENTS_TABLE       | Sub messages items (or contents table) |
| 15          |  ASSOCIATED_CONTENTS_TABLE       | Sub associated contents item (or associated contents table) |
| 16          |  SEARCH_CONTENTS_TABLE       | Search contents table |
| 17          |  ATTACHMENT_TABLE       | Attachments item |
| 18          |  RECIPIENT_TABLE       | Recipients items |
| 19          |  SEARCH_TABLE_INDEX       | Unknown |
| 20          |         | Unknown |
| 21          |         | Unknown |
| 22          |         | Unknown |
| 23          |         | Unknown |
| 24          |         | Unknown |
| 31          |  LTP       | Local descriptor value |
| 290          |  ROOT_FOLDER       | The root folder. |
| 33          |  MESSAGE_STORE       | Message store. |

### Heap-on-Node

If the encryption type was set in the file header, **the entire Heap-on-Node is encrypted**.

#### Compressible encryption

Compressible encryption is a simple [byte-substitution cipher](https://github.com/rjohnsondev/java-libpst/blob/develop/src/main/java/com/pff/PSTObject.java#L843) with a fixed [substitution table](https://github.com/rjohnsondev/java-libpst/blob/develop/src/main/java/com/pff/PSTObject.java#L725).

#### Heap-on-Node HID

| Offset        | Size          | Description   | 
| ------------- | ------------- | ------------- |
| 0             |  5 bits       | HID Type; MUST be set to 0 (IdentifierTypeHID) to indicate a valid HID. |
| 0.5           |  11 bits      | HID index. The index value that identifies an item allocated in the [allocation table](#heap-on-node-page-map). |
| 2.0           |  16 bits      | This number indicates the index of the data block in which this heap item resides.  |

#### Heap-on-Node header

The first block contains the Heap-on-Node header.

| Offset        | Size          | Description   | 
| ------------- | ------------- | ------------- |
| 0             |  2            | The offset to the [Heap-on-Node page map](#heap-on-node-page-map). |
| 2             |  1            | Block signature; MUST be set to 0xEC (236) to indicate a Heap-on-Node. |
| 3             |  1            | [The table type](#table-types). |
| 4             |  4            | [HID](#heap-on-node-hid) User Root. |

#### Heap-on-Node bitmap header

Blocks 8, 136, then every 128th contains the Heap-on-Node bitmap header (``i == 8 || i >= 138 && (i - 8) / 128 == 0``).

| Offset        | Size          | Description   | 
| ------------- | ------------- | ------------- |
| 0             |  2            | The offset to the [Heap-on-Node page map](#heap-on-node-page-map) (starting at the Heap-on-Node page header). |

#### Heap-on-Node page header

This is only used when multiple Heap-on-Node blocks are present.

| Offset        | Size          | Description   | 
| ------------- | ------------- | ------------- |
| 0             |  2            | The offset to the [Heap-on-Node page map](#heap-on-node-page-map) (starting at the Heap-on-Node page header). |

#### Heap-on-Node page map

| Offset        | Size          | Description   | 
| ------------- | ------------- | ------------- |
| 0             |  2            | Allocation count. |
| 4             |  variable     | Allocation table. This contains Allocation count + 1 entries. Each entry is an int (16 bit) value that is the byte offset to the beginning of the allocation. The start of this offset can be retrieved by using ``page map offset + (2 * hidIndex) + 2`` (page map offset plus the start of the allocation table, at the HID index offset). An extra entry exists at the Allocation count +1 position to mark the offset of the next available slot. |

#### Table types

| Table type    | Description   | Features   |
| ------------- | ------------- | ------------- |
| 108             |  6c table            | Has a b5 table header. |
| 124             |  7c table            | Table Context. Has a b5 table header. |
| 140             |  8c table            | Has a b5 table header |
| 156             |  9c table            | Has a b5 table header. |
| 165             |  a5 table            |  |
| 172             |  ac table            | Has a b5 table header. |
| 181             |  b5 table header     | B-Tree on Heap |
| 188             |  bc table            | Property Context. Has a b5 table header. |
| 204             |  cc table            | Unknown |

### B-Tree-on-Heap

#### B-Tree-on-Heap header

All tables should have a BTree-on-Heap header at [HID](#heap-on-node-hid) **0x20** (the start offset to the BTree-on-Heap header is in the [allocation table](#heap-on-node-page-map)).
This is the HID User Root from the [Heap-on-Node header](#heap-on-node-header).

| Offset        | Size          | Description   | 
| ------------- | ------------- | ------------- |
| 0             |  1            | Table type. MUST be 188. |
| 1             |  1            | Size of the BTree Key value. MUST be 2, 4, 8 or 16. |
| 2             |  1            | Size of the data value. MUST be greater than zero and less than or equal to 32. |
| 3             |  1            | Index depth. |
| 4             |  4            | [HID](#heap-on-node-hid) root. Points to the B-Tree-on-Heap entries (offset can be found in the [allocation table](#heap-on-node-page-map)). |

### Property Context

The property context starts at the [HID root](#b-tree-on-heap-header) of the B-Tree-on-Heap header.

#### Property Context B-Tree-on-Heap Record

| Offset        | Size          | Description   | 
| ------------- | ------------- | ------------- |
| 0             |  2            | Property ID.  |
| 2             |  2            | [Property type](#property-types).  |
| 4             |  4            | Reference HNID (HID or NID). In the event where the Reference HNID contains an HID or NID, the actual data is stored in the corresponding heap or subnode entry, respectively.  |

#### Property types

| Type        | Value    | 
| ------------- | ------------- |
| Integer16 | 2 |
| Integer32 | 3 |
| Floating32 | 4 |
| Floating64 | 5 |
| Currency | 6 |
| FloatingTime | 7 |
| ErrorCode | 10 |
| Boolean | 11 |
| Integer64 | 20 |
| String | 31 |
| String8 | 30 |
| Time | 64 |
| GUID | 72 |
| ServerID | 251 |
| Restriction | 253 |
| RuleAction | 254 |
| Binary | 258 |
| MultipleInteger16 | 4098 |
| MultipleInteger32 | 4099 |
| MultipleFloating32 | 4100 |
| MultipleFloating64 | 4101 |
| MultipleCurrency | 4102 |
| MultipleFloatingTime | 4103 |
| MultipleInteger64 | 4116 |
| MultipleString | 4127 |
| MultipleString8 | 4126 |
| MultipleTime | 4160 |
| MultipleGUID | 4168 |
| MultipleBinary | 4354 |
| Unspecified | 0 |
| Null | 1 |
| Object | 13 |

### The root folder

The [identifier](#identifier-types) 290 refers to the root folder.

### Sub-Folders

A folder [identifier](#identifier-types) + 11 refers to the related sub folders (for example, for the root folder this is 290 + 11 = 301).

The related sub folders consists of the Table Context (7c table).

Each property ID of 26610 has a reference HNID which points to a message.

### Messages

Messages have a property context where all data is stored.

### Table Context

The table context has a B-Tree-on-Heap.

The table context starts at the [HID User Root](#heap-on-node-header) of the Heap-on-Node.

#### Table Context Info

| Offset        | Size          | Description   | 
| ------------- | ------------- | ------------- |
| 0             |  1            | [Table context signature](#table-types). MUST be 124.  |
| 1             |  1            | Column count (number of columns in the table context).  |
| 6             |  2            | TCI_1b. Used to get the start offset to the [Cell Existence Block](#cell-existence-block).  |
| 8             |  2            | Row size.  |
| 14            |  4            | HNID to the Row Matrix (the actual table data).  |
| 22            |  variable     | [Column Descriptors](#table-context-column-descriptor). |

#### Table Context Column Descriptor

| Offset        | Size          | Description   | 
| ------------- | ------------- | ------------- |
| 0             |  2            | [Property Type](#property-types).   |
| 2             |  2            | Property ID.  |
| 4             |  2            | Data offset (from the beginning of the Row Matrix).  |
| 6             |  1            | Data size.  |
| 7             |  1            | Cell Existence Bitmap Index. See [Cell Existence Block](#cell-existence-block).  |

#### Blocks

Block sizes:
- Unicode: 8192
- Unicode4k: 65536
- ANSI: 8192

Block trailer sizes:
- Unicode: 16
- Unicode4k: 16
- ANSI: 12

#### Number of records

**Number of blocks**: Size of the table context / ([block size](#blocks) - [block trailer size](#blocks)). 
The size of the table context is retrieved by the [allocation table](#heap-on-node-page-map).

**Rows per block**: ([block size](#blocks) - [block trailer size](#blocks)) / [row size](#table-context)

**Row count**: (number of blocks * rows per block) + ((table row matrix size % (block size - block trailer size)) / row size)

**Current row start offset**: (((startAtRow + currentRowIteration) / rowsPerBlock) * (blockSize - blockTrailerSize)) + (((startAtRow + currentRowIteration) % rowsPerBlock) * rowSize)

##### Cell Existence Block

Checks if a column exists.

**Cell existence block size**: math.Ceil([columnCount](#table-context-info) / 8)

**Cell existence block**: ```tableRowMatrix[currentRowStartOffset + tci1b:currentRowStartOffset + tci1b + cellExistenceBlockSize]```

**Cell existence block exists**: ```cellExistenceBlock[column.CellExistenceBitmapIndex / 8] & (1 << (7 - (column.CellExistenceBitmapIndex % 8))) != 0```

##### Table Context Item

If the [column data size](#table-context-column-descriptor) is 1, 2 or 4, the bytes contain a HNID which points to data in the Heap-on-Node, these offsets are in the [allocation table](#heap-on-node-page-map).

### Local Descriptors

The local descriptors identifier and size is in the [b-tree entries](#the-b-tree-entries).

If the local descriptors' identifier is 0, there are no local descriptors.

The Heap-on-Node HNID (in the allocation table) may point to a local descriptor which contains it's data in the block b-tree (using the data identifier).

Local descriptor entry size:
- **64-bit and 64-bit-with-4k**: 24
- **32-bit**: 12

#### The 64-bit local descriptors

The 64-bit-with-4k local descriptors are the same format as the 64-bit local descriptors.

| Offset        | Size          | Value             | Description | 
| ------------- | ------------- | ----------------  | ----------- |
| 0             |  1            | 2                 | The signature. |
| 1             |  1            |                   | Node level (0 for leaf nodes). |
| 2             |  2            |                   | The number of entries. |
| 8             |  (number of entries * entry size) | | The entries. |

#### The 64-bit local descriptors leaf node

| Offset        | Size          | Description | 
| ------------- | ------------- | ----------- |
| 0             |  8            | The [identifier](#identifier) (HNID). |
| 8             |  8            | The data [identifier](#identifier). Searchable in the block b-tree. |
| 16            |  8            | The local descriptor [identifier](#identifier). |

#### The 32-bit local descriptors

| Offset        | Size          | Value             | Description | 
| ------------- | ------------- | ----------------  | ----------- |
| 0             |  1            | 2                 | The signature. |
| 1             |  1            |                   | Node level (0 for leaf nodes). |
| 2             |  2            |                   | The number of entries. |
| 4             |  (number of entries * entry size) | | The entries. |

#### The 32-bit local descriptors leaf node

| Offset        | Size          | Description | 
| ------------- | ------------- | ----------- |
| 0             |  4            | The [identifier](#identifier) (HNID). |
| 4             |  4            | The data [identifier](#identifier). Searchable in the block b-tree. |
| 8             |  4            | The local descriptor [identifier](#identifier). |


## Contact

Feel free to contact me if you have any questions.<br/>
**Name**: Marten Mooij<br/>
**Email**: info@mooijtech.com<br/>
**Phone**: +31 6 30 53 47 67

## License

[MIT](https://github.com/mooijtech/go-pst/blob/master/LICENSE.txt)
