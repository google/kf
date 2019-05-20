package clientbuilder

import (
	"bytes"
	"fmt"
	"strings"
)

type Field struct {
	Name    string
	Type    string
	Varadic bool
}

func (f *Field) Definition() string {
	if f.Varadic {
		return fmt.Sprintf("%s ...%s", f.Name, f.Type)
	}

	return fmt.Sprintf("%s %s", f.Name, f.Type)
}

func (f *Field) Invocation() string {
	if f.Varadic {
		return fmt.Sprintf("%s...", f.Name)
	}

	return fmt.Sprintf("%s", f.Name)
}

type FieldList []Field

func (f FieldList) Definition() string {
	var out []string

	for _, field := range f {
		out = append(out, field.Definition())
	}

	return strings.Join(out, ", ")
}

func (f FieldList) Invocation() string {
	var out []string

	for _, field := range f {
		out = append(out, field.Invocation())
	}

	return strings.Join(out, ", ")
}

type Function struct {
	Reciever *Field
	Name     string
	Args     FieldList
	Response FieldList
}

func (f *Function) Definition() string {
	out := &bytes.Buffer{}
	fmt.Fprintf(out, "func ")
	if f.Reciever != nil {
		fmt.Fprintf(out, "(%s)", f.Reciever.Definition())
	}

	fmt.Fprintf(out, "%s(%s) (%s)", f.Name, f.Args.Definition(), f.Response.Definition())

	return out.String()
}

func (f *Function) Invocation() string {
	out := &bytes.Buffer{}

	if f.Reciever != nil {
		fmt.Fprintf(out, "%s", f.Reciever.Invocation())
	}

	fmt.Fprintf(out, "%s(%s)", f.Name, f.Args.Invocation())

	return out.String()

}
