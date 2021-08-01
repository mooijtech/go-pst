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

The file header common to both the 64-bit and 32-bit PFF format consists of 24 bytes and consists of:

| Offset        | Size          | Value                     | Description   |
| ------------- | ------------- | -------------             | ------------- |
| 0             | 4             | "\x21\x42\x44\x4e" (**!BDN**) | The signature (magic identifier)

## License

[MIT](https://github.com/mooijtech/go-pst/blob/master/LICENSE.txt)
