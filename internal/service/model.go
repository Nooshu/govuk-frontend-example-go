package service

// Application answers, step graph, and completion helpers.

import "slices"

// Licence lengths the applicant can choose.
const (
	LicenceOneDay     = "1-day"
	LicenceEightDays  = "8-days"
	LicenceTwelveMths = "12-months"
)

// StepID identifies one question page.
type StepID string

// Question paths, in journey order.
const (
	StepLicenceLength    StepID = "licence-length"
	StepName             StepID = "name"
	StepDateOfBirth      StepID = "date-of-birth"
	StepWhereYouWillFish StepID = "where-you-will-fish"
	StepEmail            StepID = "email"
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
	{ID: StepLicenceLength, Path: "/licence-length", Heading: "How long do you need the licence for?"},
	{ID: StepName, Path: "/name", Heading: "What is your full name?"},
	{ID: StepDateOfBirth, Path: "/date-of-birth", Heading: "What is your date of birth?"},
	{ID: StepWhereYouWillFish, Path: "/where-you-will-fish", Heading: "Where will you fish?"},
	{ID: StepEmail, Path: "/email", Heading: "What is your email address?"},
}

// Steps returns the question pages in journey order.
//
// The result is a copy, so callers cannot reorder the journey by accident.
func Steps() []Step {
	return slices.Clone(steps)
}

// Application holds the answers collected for one fishing rod licence application.
type Application struct {
	LicenceLength string
	FullName      string
	Day           string
	Month         string
	Year          string
	Country       string
	Email         string
	Submitted     bool
	Reference     string
	Completed     []StepID
}

// NewApplication returns an empty application with no completed steps.
func NewApplication() Application {
	return Application{Completed: []StepID{}}
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

// RequiredStepsComplete reports whether every question is complete, which is when the
// applicant may check their answers.
func RequiredStepsComplete(application Application) bool {
	for _, step := range steps {
		if !application.IsCompleted(step.ID) {
			return false
		}
	}
	return true
}

// FirstIncompleteStep returns the first question that is not complete.
//
// It reports false when the applicant can go straight to check-your-answers.
func FirstIncompleteStep(application Application) (Step, bool) {
	for _, step := range steps {
		if !application.IsCompleted(step.ID) {
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
