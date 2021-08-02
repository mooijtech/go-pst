<h1 align="center">
  <br>
  <a href="https://github.com/mooijtech/go-pst"><img src="https://i.imgur.com/PwKwBRa.png" alt="go-pst" width="320"></a>
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

The PFF (Personal Folder File) and OFF (Offline Folder File) format is used to store Microsoft Outlook e-mails, appointments and contacts. The PST (Personal Storage Table), OST (Offline Storage Table) and PAB (Personal Address Book) file format consist of the PFF format.

## References

### Documentation

- [Personal Folder File (PFF) file format specification](https://github.com/mooijtech/go-pst/blob/main/docs/PFF.pdf)
- [Outlook Personal Folders (.pst) File Format](https://github.com/mooijtech/go-pst/blob/main/docs/MS-PST.pdf)

### Libraries

- [java-libpst](https://github.com/rjohnsondev/java-libpst)
- [libpff](https://github.com/libyal/libpff)
- [XstReader](https://github.com/Dijji/XstReader)
- [pstreader](https://github.com/Jmcleodfoss/pstreader)

## Datasets

This library is tested on the following datasets:

- [enron.pst](https://github.com/mooijtech/go-pst/blob/master/data/enron.pst)
- [32-bit.pst](https://github.com/mooijtech/go-pst/blob/master/data/32-bit.pst)

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

## The node and block b-tree

The following offsets start from the (node/block) b-tree offset.

### 64-bit

| Offset        | Size          | Description   |
| ------------- | ------------- | ------------- |
| 488           |  1            | The number of entries. |
| 490           |  1            | The size of an entry. |
| 491           |  1            | B-tree node level. A zero value represents a leaf node. A value greater than zero represents a branch node, with the highest level representing the root. |

### 64-bit 4k

| Offset        | Size          | Description   |
| ------------- | ------------- | ------------- |
| 4056           |  2            | The number of entries. |
| 4060           |  1            | The size of an entry. |
| 4061           |  1            | B-tree node level. A zero value represents a leaf node. A value greater than zero represents a branch node, with the highest level representing the root. |

### 32-bit

| Offset        | Size          | Description   |
| ------------- | ------------- | ------------- |
| 496           |  1            | The number of entries. |
| 498           |  1            | The size of an entry. |
| 499           |  1            | B-tree node level. A zero value represents a leaf node. A value greater than zero represents a branch node, with the highest level representing the root. |

## Contact

Feel free to contact me if you have any questions.<br/>
**Name**: Marten Mooij<br/>
**Email**: info@mooijtech.com<br/>
**Phone**: +31 6 30 53 47 67

## License

[MIT](https://github.com/mooijtech/go-pst/blob/master/LICENSE.txt)
