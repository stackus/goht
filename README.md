# Hamlet
A [Haml](http://haml.info/) template engine for Go.

![hamlet_header.png](docs/hamlet_header.png)

[![Go Report Card](https://goreportcard.com/badge/github.com/stackus/hamlet)](https://goreportcard.com/report/github.com/stackus/hamlet)
[![](https://godoc.org/github.com/stackus/hamlet?status.svg)](https://pkg.go.dev/github.com/stackus/hamlet)


## Features
- Full [Haml](http://haml.info/) language support 
- Templates are compiled to type-safe Go
- Multiple templates per file
- Mix Go and templates together in the same file
- Easy nesting of templates

```haml
@hmlt SiteLayout() {
	!!!
	%html{lang:"en"}
		%head
			%title Hamlet
		%body
			%h1 Hamlet
			%p A HAML-like template engine for Go.
			= @children
}

@hmlt HomePage() {
	= @render SiteLayout()
		%p This is the home page for Hamlet.
}
```

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <title>Hamlet</title>
  </head>
  <body>
    <h1>Hamlet</h1>
    <p>A Haml template engine for Go.</p>
    <p>This is the home page for Hamlet.</p>
  </body>
</html>
```

## Supported Haml Syntax & Features
- [x] Doctypes
- [x] Tags
- [x] Attributes [(more info)](#attributes)
- [x] Classes and IDs [(more info)](#classes)
- [x] Object References [(more info)](#object-references)
- [x] Unescaped Text
- [x] Comments
- [x] Inline Interpolation
- [x] Inlining Code
- [x] Rendering Code
- [x] Filters [(more info)](#filters)
- [x] Whitespace Removal

## Unsupported HAML Features
- [ ] Probably something I've missed, please raise an issue if you find something missing.

## Hamlet CLI

### CLI Installation
```sh
go install github.com/stackus/hamlet/cmd/hamlet@latest
```

### CLI Usage
Use `generate` to generate Go code from Hamlet template files,
that are new or newer than the generated Go files, in the current directory and subdirectories:
```sh
hamlet generate
```
Use the `--path` flag to specify a path to generate code for:
```sh
hamlet generate --path=./templates
```
In both examples, the generated code will be placed in the same directory as the template files.

Use the `--force` to generate code for all Hamlet template files, even if they are older than the generated Go files:
```sh
hamlet generate --force
```
See more options with `hamlet help generate` or `hamlet generate -h`.

## Library Installation
```sh
go get github.com/stackus/hamlet
```

### Using Hamlet
To use Hamlet you will need to create a Hamlet template file. See [The Hamlet template](#the-hamlet-template) for more information.

After you have your templates written, you will need to generate the Go code for them. This is done using the before mentioned [CLI](#hamlet-cli) tool.

The resulting Go files will have a Hamlet template function for each of the templates in the file.

When called, the template function will return a `*hamlet.Template` which can be used to render the template.

```go
package main

import (
	"context"
	"os"

	"github.com/stackus/hamlet/examples/tags"
)

func main() {
	tmpl := tags.RemoveWhitespace()

	err := tmpl.Render(context.Background(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
```
The above would render one of the examples from this library, and would output the following:
```html
<p>This text has no whitespace between it and the parent tag.</p>
<p>
There is whitespace between this text and the parent tag.<p>This text has no whitespace between it and the parent tag.
There is also no whitespace between this tag and the sibling text above it.
Finally, the tag has no whitespace between it and the outer tag.</p></p>
```
The second parameter passed into the `Render` method can be anything that implements the `io.Writer` interface,
such as a file or a buffer, or the `http.ResponseWriter` that you get from an HTTP handler.

### Using Hamlet with HTTP handlers
Using the Hamlet templates is made very easy.
```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/stackus/hamlet/examples/hello"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_ = hello.World().Render(r.Context(), w)
	})

	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
```
There are a number of examples showing various Haml and Hamlet features in the [examples](examples) directory.

### A big nod to Templ
The way that you use Hamlet is very similar to how you would use [Templ](https://templ.guide). This is no accident as I am a big fan of the work being done with that engine.

After getting the Haml properly lexed, and parsed, I did not want to reinvent the wheel and come up with a whole new rendering API.
The API that Templ presents is nice and easy to use, so I decided to replicate it in Hamlet.

## The Hamlet template
Hamlet templates are files with the extension `.hmlt` that when process will produce a matching Go file with the extension `.hmlt.go`.

In these files you are free to write any Go code that you wish, and then drop into HAML mode using the `@hmlt` directive.

The following starts the creation of a SiteLayout template:
```haml
@hmlt SiteLayout() {
```

Hamlet templates are closed like Go functions, with a closing brace `}`. So a complete but empty example is this:
```haml
@hmlt SiteLayout() {
}
```
Inside the template you can use any Haml features, such as tags, attributes, classes,
IDs, text, comments, interpolation, code inlining, code rendering, and filters.

## Haml Syntax
The Haml syntax is documented at [haml.info](http://haml.info/).
Please see that site for more information on the Haml syntax.
There are some differences that I will document here.

Important differences are:
- [Multiple templates per file](#multiple-templates-per-file): You can declare as many templates in a file as you wish.
- [Doctypes](#doctypes): Limited doctype support.
- [Inlined code](#inlined-code): You won't be using Ruby here, you'll be using Go.
- [Rendering code](#rendering-code): The catch is what is being outputted will need to be a string in all cases.
- [Attributes](#attributes): Only the Ruby 1.9 style of attributes is supported.
- [Classes](#classes): Multiple sources of classes are supported.
- [Object References](#object-references): Limited support for object references.
- [Filters](#filters): Partial list of supported filters.
- [Template nesting](#template-nesting): Templates can be nested, and content can be passed into them.

### Go package and imports
You can provide a package name at the top of your Hamlet template file. If you do not provide one then `main` will be used.

You may also import any packages that you need to use in your template. The imports you use and the ones brought in by Hamlet will be combined and deduplicated.

### Multiple templates per file
You can declare as many templates in a file as you wish. Each template must have a unique name in the module they will be output into.
```haml
@hmlt SiteLayout() {
}

@hmlt HomePage() {
}
```

The templates are converted into Go functions, so they must be valid Go function names.
This also means that you can declare them with parameters and can use those parameters in the template.
```haml
@hmlt SiteLayout(title string) {
	!!!
	%html{lang:"en"}
		%head
			%title= title
		%body
			-# ... the rest of the template
}
```

### Doctypes
Only the HTML 5 doctype is supported, and is written using `!!!`.
```haml
@hmlt SiteLayout() {
	!!!
}
```

> Note about indenting. It does not matter if you indent or do not indent. Haml normally complains if a doctype is not at the start of the line, but Hamlet does not. I just recommend choosing to indent or not indent, and sticking to it. 

### Inlined code
You won't be using Ruby here, you'll be using Go.
And in Go you need to use curly braces to denote the start and end of a block of code.
So instead of this:
```ruby
- if user
	%strong The user exists
```
You would write this:
```haml
- if user != nil {
	%strong The user exists
- }
```
There is no processing performed on the Go code you put into the templates, so it needs to be valid Go code.

### Rendering code
Like in Haml, you can output variables and the results of expressions. The `=` script syntax and text interpolation `#{}` are supported.
```haml
%strong= user.Name
%strong The user's name is #{user.Name}
```

The catch is what is being outputted will need to be a string in all cases.
So instead of writing this to output an integer value:
```haml
%strong= user.Age
```
You would need to write this:
```haml
%strong= fmt.Sprintf("%d", user.Age)
```
Which to be honest can be a bit long to write, so a shortcut is provided:
```haml
%strong=%d user.Age
```
The interpolation also supports the shortcut:
```haml
%strong #{user.Name} is #{%d user.Age} years old.
```
When formatting a value into a string `fmt.Sprintf` is used under the hood, so you can use any of the formatting options that it supports.

### Attributes
Only the Ruby 1.9 style of attributes is supported.
This syntax is closest to the Go syntax, and is the most readable.
Between the attribute name, operator, and value you can include or leave out as much whitespace as you like.
```haml
%a{href: "https://github.com/stackus/hamlet", target: "_blank"} Hamlet
```
You can supply a value to an attribute using the text interpolation syntax.
```haml
%a{href:#{url}} Hamlet
```
Attributes can be written over multiple lines, and the closing brace can be on a new line.
```haml
%a{
	href: "...",
	target: "_blank",
} Hamlet
```
Attributes which you want to render conditionally use the `?` operator instead of the `:` operator.
For conditional attributes the attribute value is required to be an interpolated value which will be used as the condition in a Go `if` statement.
```haml
%button{
	disabled ? #{disabled},
} Click me
```
> Note: The final comma is not required on the last attribute when they are spread across multiple lines like it would be in Go. Including it is fine and will not cause any issues.

Certain characters in the attribute name will require that the name be escaped.
```haml
%button{
	"@click": "onClick",
	":disabled": "disabled",
} Click me
```
Keep in mind that attribute names cannot be replaced with an interpolated string; only the value can.

To support dynamic lists of attributes, you can use the `@attributes` directive.
This directive takes a list of arguments which comes in two forms:
- `map[string]string`
	- The key is the attribute name, the value is the attribute value.
	- The attribute will be rendered if the value is not empty.
- `map[string]bool`
  - The key is the attribute name, the value is the condition to render the attribute.
```haml
%button{
	"@click": "onClick",
	":disabled": "disabled",
	@attributes: #{myAttrs},
} Click me
```
### Classes
Hamlet supports the `.` operator for classes and also will accept the `class` attribute such as `class:"foo bar"`.
However, if the class attribute is given an interpolated value, it will need to be a comma separated list of values.
These values can be the following types:
- `string`
	- `myClass` variable or `"foo bar"` string literal
- `[]string`
	- Each item will be added to the class list if it is not blank.
- `map[string]bool`
	- The key is the class name, the value is the condition to include the class.

Examples:
```haml
%button.foo.bar.baz Click me
%button.fizz{class:"foo bar baz"} Click me
%button.foo{class:#{myStrClasses, myBoolClasses}} Click me
```
All sources of classes will be combined and deduplicated into a single class attribute.

### Object References
Haml supports using a Ruby object to supply the id and class for a tag using the `[]` object reference syntax.
This is supported but is rather limited in Hamlet.
The type that you use within the brackets will be expected to implement at least one or both of the following interfaces:
```go
type ObjectIDer interface {
	ObjectID() string
}

type ObjectClasser interface {
	ObjectClass() string
}
```
The result of these methods will be used
to populate the id and class attributes in a similar way to how Haml would apply the Ruby object references.

### Filters
Only the following Haml filters are supported:
- `:plain`
- `:escaped`
- `:preserve`
- `:javascript`
- `:css`

### Template nesting
The biggest departure from Haml is how templates can be combined. When working Haml you could use `= render :partial_name` or `= haml :partial_name` to render a partial. The `render` and `haml` functions are not available in Hamlet, instead you can use the `@render` directive.
```haml
@hmlt HomePage() {
	= @render SiteLayout()
}
```
The above would render the `SiteLayout` template, and you would call it with any parameters that it needs. You can also call it and provide it with a block of content to render where it chooses.
```haml
@hmlt HomePage() {
	= @render SiteLayout()
		%p This is the home page for Hamlet.
}
```
Any content nested under the `@render` directive will be passed into the template that it can render where it wants using the `@children` directive.
```haml
@hmlt SiteLayout() {
	!!!
	%html{lang:"en"}
		%head
			%title Hamlet
		%body
			%h1 Hamlet
			%p A HAML-like template engine for Go.
			= @children
}
```

## Contributing
Contributions are welcome. Please see the [contributing guide](CONTRIBUTING.md) for more information.

## License
[MIT](LICENSE)
