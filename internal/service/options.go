package service

// Static option lists (countries, licence lengths, fees).

// Option is a selectable value and the label shown for it.
type Option struct {
	Value string
	Text  string
}

// FeeOption is a licence length and the example fee shown on the fees page.
type FeeOption struct {
	Text string
	Fee  string
}

var countries = []Option{
	{Value: "England", Text: "England"},
	{Value: "Wales", Text: "Wales"},
	{Value: "Scotland", Text: "Scotland"},
}

var licenceLengths = []Option{
	{Value: LicenceOneDay, Text: "1 day"},
	{Value: LicenceEightDays, Text: "8 days"},
	{Value: LicenceTwelveMths, Text: "12 months"},
}

var licenceFees = []FeeOption{
	{Text: "1 day", Fee: "£7.10"},
	{Text: "8 days", Fee: "£14.20"},
	{Text: "12 months", Fee: "£36.80"},
}

// Countries returns the countries the applicant can fish in.
func Countries() []Option { return append([]Option(nil), countries...) }

// LicenceLengths returns the licence lengths shown on the question (no fees).
func LicenceLengths() []Option { return append([]Option(nil), licenceLengths...) }

// LicenceFees returns the fees shown on the separate fees demo page.
func LicenceFees() []FeeOption { return append([]FeeOption(nil), licenceFees...) }

// LabelFor returns the label for a selected value.
//
// An unknown value is returned unchanged, so a stored answer is never silently blanked.
func LabelFor(options []Option, value string) string {
	for _, option := range options {
		if option.Value == value {
			return option.Text
		}
	}
	return value
}
