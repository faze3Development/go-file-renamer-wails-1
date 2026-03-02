package namer

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Info struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// templateNamer generates names based on a user-provided template string.
type templateNamer struct{ template string }

// dateTimeNamer generates names based on the current date and time.
type dateTimeNamer struct{ format string }

// randomNamer generates a random alphanumeric string of a specified length.
type randomNamer struct {
	length int
}
