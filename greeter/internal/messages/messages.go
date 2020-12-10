package messages

// Greeter creates a greeting message on a certain language
type Greeter func(string) string

var (
	// Greeters is the set of all available Greeter instances
	Greeters map[string]Greeter

	// DefaultLanguage language to use if none is 
	DefaultLanguage = "en"
)

func init() {
	Greeters = make(map[string]Greeter)
	Greeters["en"] = func(who string) string { return "Hello " + who }
	Greeters["fr"] = func(who string) string { return "Bonjour " + who }
	Greeters["bg"] = func(who string) string { return "Здравей " + who }
}
