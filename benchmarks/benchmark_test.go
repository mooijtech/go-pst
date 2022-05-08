// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2022 Marten Mooij (https://www.mooijtech.com/)
package benchmarks

// Uncomment and run via (make sure to move enron.pst to this directory):
// go test -run=. -bench=. -benchtime=5s -count 3 -timeout 0 -benchmem -cpuprofile=cpu.out -memprofile=mem.out -trace=trace.out ./ | tee bench.txt
//
// After run export to SVG:
// go tool pprof -svg cpu.out
// go tool pprof -svg mem.out

//import (
//	"fmt"
//	pstv3 "github.com/mooijtech/go-pst/v3/pkg"
//	pstv4 "github.com/mooijtech/go-pst/v4/pkg"
//	"testing"
//)
//
//func BenchmarkBTreesV4(b *testing.B) {
//	pstFile, err := pstv4.NewFromFile("enron.pst")
//
//	if err != nil {
//		fmt.Printf("Failed to create PST file: %s\n", err)
//		return
//	}
//
//	defer func() {
//		err := pstFile.Close()
//
//		if err != nil {
//			fmt.Printf("Failed to close PST file: %s", err)
//		}
//	}()
//
//	formatType, err := pstFile.GetFormatType()
//
//	if err != nil {
//		fmt.Printf("Failed to get format type: %s\n", err)
//		return
//	}
//
//	err = pstFile.InitializeBTrees(formatType)
//
//	if err != nil {
//		fmt.Printf("Failed to initialize node and block b-tree.\n")
//		return
//	}
//
//	encryptionType, err := pstFile.GetEncryptionType(formatType)
//
//	if err != nil {
//		fmt.Printf("Failed to get encryption type: %s\n", err)
//		return
//	}
//
//	rootFolder, err := pstFile.GetRootFolder(formatType, encryptionType)
//
//	if err != nil {
//		fmt.Printf("Failed to get root folder: %s\n", err)
//		return
//	}
//
//	err = GetSubFoldersV4(pstFile, rootFolder, formatType, encryptionType)
//
//	if err != nil {
//		fmt.Printf("Failed to get sub-folders: %s", err)
//		return
//	}
//}
//
//func GetSubFoldersV4(pstFile pstv4.File, folder pstv4.Folder, formatType string, encryptionType string) error {
//	subFolders, err := pstFile.GetSubFolders(folder, formatType, encryptionType)
//
//	if err != nil {
//		return err
//	}
//
//	for _, subFolder := range subFolders {
//		messages, err := pstFile.GetMessages(subFolder, formatType, encryptionType)
//
//		if err != nil {
//			return err
//		}
//
//		if len(messages) > 0 {
//
//		}
//
//		err = GetSubFoldersV4(pstFile, subFolder, formatType, encryptionType)
//
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func BenchmarkBTreesV3(b *testing.B) {
//	pstFile, err := pstv3.NewFromFile("enron.pst")
//
//	if err != nil {
//		fmt.Printf("Failed to create PST file: %s\n", err)
//		return
//	}
//
//	defer func() {
//		err := pstFile.Close()
//
//		if err != nil {
//			fmt.Printf("Failed to close PST file: %s", err)
//		}
//	}()
//
//	formatType, err := pstFile.GetFormatType()
//
//	if err != nil {
//		fmt.Printf("Failed to get format type: %s\n", err)
//		return
//	}
//
//	err = pstFile.InitializeBTrees(formatType)
//
//	if err != nil {
//		fmt.Printf("Failed to initialize node and block b-tree.\n")
//		return
//	}
//
//	encryptionType, err := pstFile.GetEncryptionType(formatType)
//
//	if err != nil {
//		fmt.Printf("Failed to get encryption type: %s\n", err)
//		return
//	}
//
//	rootFolder, err := pstFile.GetRootFolder(formatType, encryptionType)
//
//	if err != nil {
//		fmt.Printf("Failed to get root folder: %s\n", err)
//		return
//	}
//
//	err = GetSubFoldersV3(pstFile, rootFolder, formatType, encryptionType)
//
//	if err != nil {
//		fmt.Printf("Failed to get sub-folders: %s", err)
//		return
//	}
//}
//
//func GetSubFoldersV3(pstFile pstv3.File, folder pstv3.Folder, formatType string, encryptionType string) error {
//	subFolders, err := pstFile.GetSubFolders(folder, formatType, encryptionType)
//
//	if err != nil {
//		return err
//	}
//
//	for _, subFolder := range subFolders {
//		messages, err := pstFile.GetMessages(subFolder, formatType, encryptionType)
//
//		if err != nil {
//			return err
//		}
//
//		if len(messages) > 0 {
//
//		}
//
//		err = GetSubFoldersV3(pstFile, subFolder, formatType, encryptionType)
//
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
