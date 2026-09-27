package service

// Persist validated answers onto an Application.

// Each Save… function returns a new [Application] rather than mutating the caller's copy, and
// marks the step complete only when validation passed. Saving invalid answers keeps what the
// applicant typed so the question can be shown back to them with their values retained.

// SaveName stores the name question.
func SaveName(application Application, firstName, lastName string, valid bool) Application {
	application.FirstName = Clean(firstName)
	application.LastName = Clean(lastName)
	application.Completed = finish(application.Completed, StepName, valid)
	return application
}

// SaveDate stores the date of birth exactly as typed, so a wrong date can be corrected.
func SaveDate(application Application, day, month, year string, valid bool) Application {
	application.Day = Clean(day)
	application.Month = Clean(month)
	application.Year = Clean(year)
	application.Completed = finish(application.Completed, StepDateOfBirth, valid)
	return application
}

// SaveEmail stores the email address.
func SaveEmail(application Application, email string, valid bool) Application {
	application.Email = Clean(email)
	application.Completed = finish(application.Completed, StepEmail, valid)
	return application
}

// SaveContact stores the contact preference. The telephone number is kept only as text.
func SaveContact(application Application, contactBy, telephone string, valid bool) Application {
	application.ContactBy = AsContactBy(contactBy)
	application.Telephone = Clean(telephone)
	application.Completed = finish(application.Completed, StepContactPreference, valid)
	return application
}

// SaveRegions stores the regions the applicant selected, dropping any value that is not an
// option on the page.
func SaveRegions(application Application, selected []string, valid bool) Application {
	known := map[string]bool{NotSure: true}
	for _, region := range regions {
		known[region.Value] = true
	}
	kept := make([]string, 0, len(selected))
	for _, region := range selected {
		if known[region] {
			kept = append(kept, region)
		}
	}
	application.Regions = kept
	application.Completed = finish(application.Completed, StepWhereYouWillFish, valid)
	return application
}

// SaveLicence stores the licence length.
func SaveLicence(application Application, value string, valid bool) Application {
	application.LicenceLength = AsLicenceLength(value)
	application.Completed = finish(application.Completed, StepLicenceLength, valid)
	return application
}

// SaveMonth stores the month the licence should start.
func SaveMonth(application Application, value string, valid bool) Application {
	application.StartMonth = value
	application.Completed = finish(application.Completed, StepStartMonth, valid)
	return application
}

// Address holds the four address fields as posted.
type Address struct {
	Line1    string
	Line2    string
	Town     string
	Postcode string
}

// SaveAddress stores the address. The postcode is normalised only when the answer is valid, so
// an unrecognisable postcode is shown back as the applicant typed it.
func SaveAddress(application Application, values Address, valid bool) Application {
	application.AddressLine1 = Clean(values.Line1)
	application.AddressLine2 = Clean(values.Line2)
	application.Town = Clean(values.Town)
	if valid {
		application.Postcode = NormalisePostcode(values.Postcode)
	} else {
		application.Postcode = Clean(values.Postcode)
	}
	application.Completed = finish(application.Completed, StepAddress, valid)
	return application
}

// SaveEvidence stores an evidence filename. A missing or rejected upload keeps the previous
// name, so re-submitting the page without choosing a file does not erase the answer.
func SaveEvidence(application Application, filename string, hasFile, valid bool) Application {
	if hasFile && valid {
		application.EvidenceFilename = filename
	}
	application.Completed = finish(application.Completed, StepEvidence, valid)
	return application
}

// SaveDetails stores the optional extra details.
func SaveDetails(application Application, value string, valid bool) Application {
	application.AdditionalDetails = value
	application.Completed = finish(application.Completed, StepAdditionalDetails, valid)
	return application
}

// SavePassword records that a password was accepted. The password itself is not stored.
func SavePassword(application Application, valid bool) Application {
	application.PasswordCreated = valid
	application.Completed = finish(application.Completed, StepCreateAPassword, valid)
	return application
}

func finish(completed []StepID, id StepID, valid bool) []StepID {
	if valid {
		return MarkCompleted(completed, id)
	}
	return UnmarkCompleted(completed, id)
}
