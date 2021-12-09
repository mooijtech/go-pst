// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

// GetContactGivenName returns the given name of this contact.
func (message *Message) GetContactGivenName(pstFile *File, formatType string, encryptionType string) (string, error) {
	givenName, err := message.GetString(14854, pstFile, formatType, encryptionType)

	if err != nil {
		return "", err
	}

	return givenName, nil
}

// GetContactBusinessPhoneNumber returns the business phone number of this contact.
func (message *Message) GetContactBusinessPhoneNumber(pstFile *File, formatType string, encryptionType string) (string, error) {
	businessPhoneNumber, err := message.GetString(14856, pstFile, formatType, encryptionType)

	if err != nil {
		return "", err
	}

	return businessPhoneNumber, nil
}

// GetContactMobilePhoneNumber returns the contact's mobile phone number.
func (message *Message) GetContactMobilePhoneNumber(pstFile *File, formatType string, encryptionType string) (string, error) {
	mobilePhoneNumber, err := message.GetString(14876, pstFile, formatType, encryptionType)

	if err != nil {
		return "", err
	}

	return mobilePhoneNumber, nil
}

// GetContactCompanyName returns the contact's company name.
func (message *Message) GetContactCompanyName(pstFile *File, formatType string, encryptionType string) (string, error) {
	companyName, err := message.GetString(14870, pstFile, formatType, encryptionType)

	if err != nil {
		return "", err
	}

	return companyName, nil
}

// GetContactEmailDisplayName returns the contact's email display name.
func (message *Message) GetContactEmailDisplayName(pstFile *File, formatType string, encryptionType string) (string, error) {
	propertyID, err := pstFile.NameToIDMap.GetPropertyID(0x00008080, PropertySetAddress)

	if err != nil {
		return "", err
	}

	return message.GetString(propertyID, pstFile, formatType, encryptionType)
}