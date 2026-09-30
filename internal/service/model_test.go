package service_test

import (
	"testing"

	"github.com/Nooshu/govuk-frontend-example-go/internal/service"
)

func TestTheJourneyIsAChainOfQuestions(t *testing.T) {
	t.Parallel()
	steps := service.Steps()
	if len(steps) == 0 {
		t.Fatal("the journey has no questions")
	}
	wantIDs := []service.StepID{
		service.StepLicenceLength,
		service.StepName,
		service.StepDateOfBirth,
		service.StepWhereYouWillFish,
		service.StepEmail,
	}
	if len(steps) != len(wantIDs) {
		t.Fatalf("got %d steps, want %d", len(steps), len(wantIDs))
	}
	for index, step := range steps {
		if step.ID != wantIDs[index] {
			t.Errorf("step %d is %s, want %s", index, step.ID, wantIDs[index])
		}
		if step.Path != "/"+string(step.ID) {
			t.Errorf("%s is served from %q, want /%s", step.ID, step.Path, step.ID)
		}
		if step.Heading == "" {
			t.Errorf("%s has no heading", step.ID)
		}
		next, hasNext := service.NextStep(step.ID)
		if index == len(steps)-1 {
			if hasNext {
				t.Errorf("the last question has a next step: %s", next.ID)
			}
		} else if !hasNext || next.ID != steps[index+1].ID {
			t.Errorf("after %s comes %s, want %s", step.ID, next.ID, steps[index+1].ID)
		}
		previous, hasPrevious := service.PreviousStep(step.ID)
		if index == 0 {
			if hasPrevious {
				t.Errorf("the first question has a previous step: %s", previous.ID)
			}
		} else if !hasPrevious || previous.ID != steps[index-1].ID {
			t.Errorf("before %s comes %s, want %s", step.ID, previous.ID, steps[index-1].ID)
		}
	}
}

func TestStepsAreFoundByIDAndPath(t *testing.T) {
	t.Parallel()
	step, ok := service.StepByID("licence-length")
	if !ok || step.Path != "/licence-length" {
		t.Errorf("StepByID(\"licence-length\") = %+v, %t", step, ok)
	}
	if _, ok := service.StepByID("not-a-step"); ok {
		t.Error("an unknown id was treated as a question")
	}
	step, ok = service.StepByPath("/name")
	if !ok || step.ID != service.StepName {
		t.Errorf("StepByPath(\"/name\") = %+v, %t", step, ok)
	}
	if _, ok := service.StepByPath("/fees"); ok {
		t.Error("a content page was treated as a question")
	}
	if _, ok := service.NextStep("not-a-step"); ok {
		t.Error("an unknown id has a next step")
	}
	if _, ok := service.PreviousStep("not-a-step"); ok {
		t.Error("an unknown id has a previous step")
	}
}

func TestMarkingStepsCompleteDoesNotDuplicateThem(t *testing.T) {
	t.Parallel()
	completed := service.MarkCompleted(nil, service.StepName)
	completed = service.MarkCompleted(completed, service.StepName)
	if len(completed) != 1 {
		t.Errorf("marking twice gave %v", completed)
	}
	completed = service.MarkCompleted(completed, service.StepEmail)
	completed = service.UnmarkCompleted(completed, service.StepName)
	if len(completed) != 1 || completed[0] != service.StepEmail {
		t.Errorf("after unmarking, completed = %v", completed)
	}
	if got := service.UnmarkCompleted(completed, service.StepLicenceLength); len(got) != 1 {
		t.Errorf("unmarking a step that was never complete changed the list: %v", got)
	}
}

func TestEveryQuestionMustBeComplete(t *testing.T) {
	t.Parallel()
	application := service.NewApplication()
	if service.RequiredStepsComplete(application) {
		t.Fatal("an empty application is ready to submit")
	}
	first, ok := service.FirstIncompleteStep(application)
	if !ok || first.ID != service.StepLicenceLength {
		t.Errorf("the first incomplete question is %s, want licence-length", first.ID)
	}

	for _, step := range service.Steps() {
		application.Completed = service.MarkCompleted(application.Completed, step.ID)
	}
	if !service.RequiredStepsComplete(application) {
		t.Error("an application with every question answered is not ready")
	}
	if _, ok := service.FirstIncompleteStep(application); ok {
		t.Error("a ready application still reports an incomplete question")
	}
	if !application.IsCompleted(service.StepLicenceLength) {
		t.Error("IsCompleted does not see a completed step")
	}
}

func TestANewApplicationIsEmpty(t *testing.T) {
	t.Parallel()
	application := service.NewApplication()
	if len(application.Completed) != 0 || application.FullName != "" || application.Country != "" {
		t.Errorf("a new application is not empty: %+v", application)
	}
	if application.Submitted {
		t.Error("a new application is already submitted")
	}
}
