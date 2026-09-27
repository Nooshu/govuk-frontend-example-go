// Package service holds the rod fishing licence journey: the answers an applicant gives, the
// rules those answers must satisfy, and the GOV.UK Frontend component parameters each question
// page needs.
//
// Nothing in this package writes HTML. Component parameters are plain maps handed to a
// [github.com/Nooshu/govuk-frontend-example-go/internal/render.Renderer], so the HTML is always
// produced by GOV.UK Frontend rather than by hand-written markup.
package service

import "slices"

// Contact methods the applicant can choose on the contact-preference question.
const (
	ContactByEmail     = "email"
	ContactByTelephone = "telephone"
)

// Licence lengths the applicant can choose.
const (
	LicenceOneDay    = "1-day"
	LicenceEightDay  = "8-day"
	LicenceTwelveMth = "12-month"
)

// StepID identifies one question page.
type StepID string

// Question paths, in the order the task list walks them.
const (
	StepName              StepID = "name"
	StepDateOfBirth       StepID = "date-of-birth"
	StepEmail             StepID = "email"
	StepContactPreference StepID = "contact-preference"
	StepWhereYouWillFish  StepID = "where-you-will-fish"
	StepLicenceLength     StepID = "licence-length"
	StepStartMonth        StepID = "start-month"
	StepAddress           StepID = "address"
	StepEvidence          StepID = "evidence"
	StepAdditionalDetails StepID = "additional-details"
	StepCreateAPassword   StepID = "create-a-password"
)

// Step is one question page in the journey.
type Step struct {
	// ID is the step identifier, which is also its path segment.
	ID StepID
	// Path is the URL the question is served from, such as "/name".
	Path string
	// Heading is the page h1, or the legend or label that acts as the h1.
	Heading string
}

// steps are the question pages in journey order.
var steps = []Step{
	{ID: StepName, Path: "/name", Heading: "What is your name?"},
	{ID: StepDateOfBirth, Path: "/date-of-birth", Heading: "What is your date of birth?"},
	{ID: StepEmail, Path: "/email", Heading: "What is your email address?"},
	{ID: StepContactPreference, Path: "/contact-preference", Heading: "How should we contact you?"},
	{ID: StepWhereYouWillFish, Path: "/where-you-will-fish", Heading: "Where will you fish?"},
	{ID: StepLicenceLength, Path: "/licence-length", Heading: "How long do you need a licence for?"},
	{ID: StepStartMonth, Path: "/start-month", Heading: "When should the licence start?"},
	{ID: StepAddress, Path: "/address", Heading: "What is your address?"},
	{ID: StepEvidence, Path: "/evidence", Heading: "Upload evidence of a concession"},
	{ID: StepAdditionalDetails, Path: "/additional-details", Heading: "Is there anything else we should know?"},
	{ID: StepCreateAPassword, Path: "/create-a-password", Heading: "Create a password"},
}

// optionalSteps are the questions an applicant can skip and still submit.
var optionalSteps = map[StepID]bool{
	StepEvidence:          true,
	StepAdditionalDetails: true,
}

// Steps returns the question pages in journey order.
//
// The result is a copy, so callers cannot reorder the journey by accident.
func Steps() []Step {
	return slices.Clone(steps)
}

// Optional reports whether a step can be left incomplete.
func Optional(id StepID) bool {
	return optionalSteps[id]
}

// Application holds the answers collected for one rod licence application.
//
// The password itself is never stored; only [Application.PasswordCreated] records that one was
// accepted.
type Application struct {
	FirstName         string
	LastName          string
	Day               string
	Month             string
	Year              string
	Email             string
	ContactBy         string
	Telephone         string
	Regions           []string
	LicenceLength     string
	StartMonth        string
	AddressLine1      string
	AddressLine2      string
	Town              string
	Postcode          string
	EvidenceFilename  string
	AdditionalDetails string
	PasswordCreated   bool
	Submitted         bool
	Reference         string
	Completed         []StepID
}

// NewApplication returns an empty application with no completed steps.
func NewApplication() Application {
	return Application{Regions: []string{}, Completed: []StepID{}}
}

// IsCompleted reports whether the applicant has finished a step.
func (a Application) IsCompleted(id StepID) bool {
	return slices.Contains(a.Completed, id)
}

// StepByID finds a question by its identifier.
func StepByID(id string) (Step, bool) {
	for _, step := range steps {
		if string(step.ID) == id {
			return step, true
		}
	}
	return Step{}, false
}

// StepByPath finds a question by its path, such as "/name".
func StepByPath(path string) (Step, bool) {
	for _, step := range steps {
		if step.Path == path {
			return step, true
		}
	}
	return Step{}, false
}

// NextStep returns the question after id, or false after the last question.
func NextStep(id StepID) (Step, bool) {
	index := indexOf(id)
	if index == -1 || index+1 >= len(steps) {
		return Step{}, false
	}
	return steps[index+1], true
}

// PreviousStep returns the question before id, or false on the first question.
func PreviousStep(id StepID) (Step, bool) {
	index := indexOf(id)
	if index <= 0 {
		return Step{}, false
	}
	return steps[index-1], true
}

// MarkCompleted returns completed with id added, without duplicating it.
func MarkCompleted(completed []StepID, id StepID) []StepID {
	if slices.Contains(completed, id) {
		return slices.Clone(completed)
	}
	return append(slices.Clone(completed), id)
}

// UnmarkCompleted returns completed without id.
func UnmarkCompleted(completed []StepID, id StepID) []StepID {
	result := make([]StepID, 0, len(completed))
	for _, item := range completed {
		if item != id {
			result = append(result, item)
		}
	}
	return result
}

// RequiredStepsComplete reports whether every required question is complete, which is when the
// applicant may check their answers.
func RequiredStepsComplete(application Application) bool {
	for _, step := range steps {
		if !Optional(step.ID) && !application.IsCompleted(step.ID) {
			return false
		}
	}
	return true
}

// FirstIncompleteStep returns the first required question that is not complete.
//
// It reports false when the applicant can go straight to check-your-answers.
func FirstIncompleteStep(application Application) (Step, bool) {
	for _, step := range steps {
		if !Optional(step.ID) && !application.IsCompleted(step.ID) {
			return step, true
		}
	}
	return Step{}, false
}

func indexOf(id StepID) int {
	for index, step := range steps {
		if step.ID == id {
			return index
		}
	}
	return -1
}
