// Package greet is the seed package autodev agents are asked to extend.
package greet

// Goodbye returns a farewell message for the given name.
func Goodbye(name string) string {
	if name == "" {
		return "Goodbye, friend!"
	}
	return "Goodbye, " + name + "!"
}
