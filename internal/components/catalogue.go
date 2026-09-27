package components

const designSystem = "https://design-system.service.gov.uk/components"

// Info is the catalogue entry for one GOV.UK Frontend component.
type Info struct {
	Name            string
	Title           string
	Description     string
	DesignSystemURL string
}

// details is this example's own copy for each component in the pinned release. It is deliberately
// hand-written: the package ships fixtures and macros, not descriptions.
var details = map[string]Info{
	"accordion":           {Title: "Accordion", Description: "Lets users show and hide sections of related content.", DesignSystemURL: designSystem + "/accordion/"},
	"back-link":           {Title: "Back link", Description: "Link to the previous page in a journey.", DesignSystemURL: designSystem + "/back-link/"},
	"breadcrumbs":         {Title: "Breadcrumbs", Description: "Helps users move between levels of a section.", DesignSystemURL: designSystem + "/breadcrumbs/"},
	"button":              {Title: "Button", Description: "Starts or continues an action.", DesignSystemURL: designSystem + "/button/"},
	"character-count":     {Title: "Character count", Description: "Shows how many characters are left in a textarea.", DesignSystemURL: designSystem + "/character-count/"},
	"checkboxes":          {Title: "Checkboxes", Description: "Lets users select one or more options.", DesignSystemURL: designSystem + "/checkboxes/"},
	"cookie-banner":       {Title: "Cookie banner", Description: "Asks users to accept or reject analytics cookies.", DesignSystemURL: designSystem + "/cookie-banner/"},
	"date-input":          {Title: "Date input", Description: "Asks users for a date they already know.", DesignSystemURL: designSystem + "/date-input/"},
	"details":             {Title: "Details", Description: "Hides content that only some users need.", DesignSystemURL: designSystem + "/details/"},
	"error-message":       {Title: "Error message", Description: "Tells users how to fix a field that failed validation.", DesignSystemURL: designSystem + "/error-message/"},
	"error-summary":       {Title: "Error summary", Description: "Summarises form errors at the top of the page.", DesignSystemURL: designSystem + "/error-summary/"},
	"exit-this-page":      {Title: "Exit this page", Description: "Lets users leave a page quickly. For services where someone may be in danger.", DesignSystemURL: designSystem + "/exit-this-page/"},
	"feedback":            {Title: "Feedback", Description: "Asks users what they think of a page. Trial component in Frontend 6.5.", DesignSystemURL: designSystem + "/feedback/"},
	"fieldset":            {Title: "Fieldset", Description: "Groups related form fields, such as an address.", DesignSystemURL: designSystem + "/fieldset/"},
	"file-upload":         {Title: "File upload", Description: "Lets users select a file to upload.", DesignSystemURL: designSystem + "/file-upload/"},
	"footer":              {Title: "Footer", Description: "Page footer with Open Government Licence and Crown copyright.", DesignSystemURL: designSystem + "/footer/"},
	"generic-header":      {Title: "Generic header", Description: "Header for services that are not branded as GOV.UK. Shown in the catalogue only.", DesignSystemURL: "https://design-system.service.gov.uk/styles/page-template/"},
	"header":              {Title: "Header", Description: "The GOV.UK masthead.", DesignSystemURL: designSystem + "/header/"},
	"hint":                {Title: "Hint", Description: "Extra help for a form field. Form controls include it; the catalogue shows it on its own.", DesignSystemURL: "https://design-system.service.gov.uk/get-started/labels-legends-headings/"},
	"input":               {Title: "Text input", Description: "Lets users enter a single line of text.", DesignSystemURL: designSystem + "/text-input/"},
	"inset-text":          {Title: "Inset text", Description: "Draws attention to important content on the page.", DesignSystemURL: designSystem + "/inset-text/"},
	"label":               {Title: "Label", Description: "Labels a form field. Form controls include it; the catalogue shows it on its own.", DesignSystemURL: "https://design-system.service.gov.uk/get-started/labels-legends-headings/"},
	"language-navigation": {Title: "Language navigation", Description: "Lets users switch between languages. Trial component in Frontend 6.5.", DesignSystemURL: designSystem + "/language-navigation/"},
	"notification-banner": {Title: "Notification banner", Description: "Tells users about something that affects the whole service.", DesignSystemURL: designSystem + "/notification-banner/"},
	"pagination":          {Title: "Pagination", Description: "Splits a long list across pages.", DesignSystemURL: designSystem + "/pagination/"},
	"panel":               {Title: "Panel", Description: "Confirms a transaction is complete.", DesignSystemURL: designSystem + "/panel/"},
	"password-input":      {Title: "Password input", Description: "Lets users enter a password, with a control to show or hide it.", DesignSystemURL: designSystem + "/password-input/"},
	"phase-banner":        {Title: "Phase banner", Description: "Shows users that the service is still being tried out.", DesignSystemURL: designSystem + "/phase-banner/"},
	"radios":              {Title: "Radios", Description: "Lets users select one option from a list.", DesignSystemURL: designSystem + "/radios/"},
	"select":              {Title: "Select", Description: "Lets users choose one option from a long list.", DesignSystemURL: designSystem + "/select/"},
	"service-navigation":  {Title: "Service navigation", Description: "Shows the service name under the GOV.UK masthead.", DesignSystemURL: designSystem + "/service-navigation/"},
	"skip-link":           {Title: "Skip link", Description: "Lets keyboard users skip to the main content.", DesignSystemURL: designSystem + "/skip-link/"},
	"summary-list":        {Title: "Summary list", Description: "Summarises answers so users can check them.", DesignSystemURL: designSystem + "/summary-list/"},
	"table":               {Title: "Table", Description: "Shows information in rows and columns.", DesignSystemURL: designSystem + "/table/"},
	"tabs":                {Title: "Tabs", Description: "Lets users switch between related views. Content stays in the page without JavaScript.", DesignSystemURL: designSystem + "/tabs/"},
	"tag":                 {Title: "Tag", Description: "Shows a short status, such as on a task list.", DesignSystemURL: designSystem + "/tag/"},
	"task-list":           {Title: "Task list", Description: "Shows the tasks in an application and whether they are done.", DesignSystemURL: designSystem + "/task-list/"},
	"textarea":            {Title: "Textarea", Description: "Lets users enter more than one line of text. This service uses character count, which includes a textarea.", DesignSystemURL: designSystem + "/textarea/"},
	"warning-text":        {Title: "Warning text", Description: "Tells users about something important before they continue.", DesignSystemURL: designSystem + "/warning-text/"},
}

// Describe returns the catalogue copy for one component.
//
// A component this example has not described yet still gets an entry, derived from its name, so
// upgrading GOV.UK Frontend never leaves a gap in the catalogue.
func Describe(name string) Info {
	if known, ok := details[name]; ok {
		known.Name = name
		return known
	}
	return Info{
		Name:            name,
		Title:           TitleFromKebab(name),
		Description:     "GOV.UK Frontend component.",
		DesignSystemURL: designSystem + "/" + name + "/",
	}
}

// Catalogue returns an entry for every component in the pinned package, in name order.
func (l *Library) Catalogue() ([]Info, error) {
	names, err := l.Names()
	if err != nil {
		return nil, err
	}
	entries := make([]Info, 0, len(names))
	for _, name := range names {
		entries = append(entries, Describe(name))
	}
	return entries, nil
}

// DescribedNames returns the component names this catalogue has hand-written copy for, so a test
// can compare them with the pinned package.
func DescribedNames() []string {
	names := make([]string, 0, len(details))
	for name := range details {
		names = append(names, name)
	}
	return names
}
