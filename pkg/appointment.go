// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"time"
)

// GetAppointmentStartTime returns the appointment's start time.
func (message *Message) GetAppointmentStartTime(pstFile *File) (time.Time, error) {
	propertyID, err := pstFile.NameToIDMap.GetPropertyID(0x0000820d, PropertySetAppointment)

	if err != nil {
		return time.Time{}, err
	}

	return message.GetDate(propertyID)
}

// GetAppointmentEndTime returns the appointment's end time.
func (message *Message) GetAppointmentEndTime(pstFile *File) (time.Time, error) {
	propertyID, err := pstFile.NameToIDMap.GetPropertyID(0x0000820e, PropertySetAppointment)

	if err != nil {
		return time.Time{}, err
	}

	return message.GetDate(propertyID)
}

// GetAppointmentLocation returns the location of the appointment.
func (message *Message) GetAppointmentLocation(pstFile *File, formatType string, encryptionType string) (string, error) {
	propertyID, err := pstFile.NameToIDMap.GetPropertyID(0x00008208, PropertySetAppointment)

	if err != nil {
		return "", err
	}

	return message.GetString(propertyID, pstFile, formatType, encryptionType)
}

// GetAppointmentAllAttendees returns all attendees to this appointment.
func (message *Message) GetAppointmentAllAttendees(pstFile *File, formatType string, encryptionType string) (string, error) {
	propertyID, err := pstFile.NameToIDMap.GetPropertyID(0x00008238, PropertySetAppointment)

	if err != nil {
		return "", err
	}

	return message.GetString(propertyID, pstFile, formatType, encryptionType)
}