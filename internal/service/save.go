package service

// Persist validated answers onto an Application.

// Each Save… function returns a new [Application] rather than mutating the caller's copy, and
// marks the step complete only when validation passed. Saving invalid answers keeps what the
// applicant typed so the question can be shown back to them with their values retained.

// SaveName stores the full name question.
func SaveName(application Application, fullName string, valid bool) Application {
	application.FullName = Clean(fullName)
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

// SaveCountry stores where the applicant will fish.
func SaveCountry(application Application, country string, valid bool) Application {
	application.Country = Clean(country)
	application.Completed = finish(application.Completed, StepWhereYouWillFish, valid)
	return application
}

// SaveLicence stores the licence length.
func SaveLicence(application Application, value string, valid bool) Application {
	application.LicenceLength = AsLicenceLength(value)
	application.Completed = finish(application.Completed, StepLicenceLength, valid)
	return application
}

func finish(completed []StepID, id StepID, valid bool) []StepID {
	if valid {
		return MarkCompleted(completed, id)
	}
	return UnmarkCompleted(completed, id)
}
