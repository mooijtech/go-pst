// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package main

import (
	log "github.com/sirupsen/logrus"
	pst "pst/pkg"
)

func main() {
	pstFile := pst.New("data/enron.pst")

	log.Infof("Parsing file: %s", pstFile.Filepath)
}
