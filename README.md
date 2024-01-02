# Hamlet
A [Haml](http://haml.info/) template engine for Go.

![hamlet_header.png](docs/hamlet_header.png)

[![Go Report Card](https://goreportcard.com/badge/github.com/stackus/hamlet)](https://goreportcard.com/report/github.com/stackus/hamlet)
[![](https://godoc.org/github.com/stackus/hamlet?status.svg)](https://pkg.go.dev/github.com/stackus/hamlet)

## Table of Contents
- [Features](#features)
- [Quick Start](#quick-start)
- [Supported Haml Syntax & Features](#supported-haml-syntax--features)
  - [Unsupported Haml Features](#unsupported-haml-features)
- [Hamlet CLI](#hamlet-cli)
- [Library Installation](#library-installation)
- [Using Hamlet](#using-hamlet)
  - [Using Hamlet with HTTP handlers](#using-hamlet-with-http-handlers)
  - [A big nod to Templ](#a-big-nod-to-templ)
- [The Hamlet template](#the-hamlet-template)
- [Haml Syntax](#haml-syntax)
  - [Hamlet and Haml differences](#hamlet-and-haml-differences)
    - [Go package and imports](#go-package-and-imports)
    - [Multiple templates per file](#multiple-templates-per-file)
    - [Doctypes](#doctypes)
    - [Inlined code](#inlined-code)
    - [Rendering code](#rendering-code)
    - [Attributes](#attributes)
    - [Classes](#classes)
    - [Object References](#object-references)
    - [Filters](#filters)
    - [Template nesting](#template-nesting)
- [Contributing](#contributing)
- [License](#license)

## Features
- Full [Haml](http://haml.info/) language support 
- Templates are compiled to type-safe Go and not parsed at runtime
- Multiple templates per file
- Mix Go and templates together in the same file
- Easy nesting of templates

## Quick Start
First create a Hamlet file, a file which mixes Go and Haml with a `.hmlt` extension:
```haml
package main

var siteTitle = "Hamlet"

@hmlt SiteLayout(pageTitle string) {
  !!!
  %html{lang:"en"}
    %head
      %title= siteTitle
    %body
      %h1= pageTitle
      %p A type-safe HAML template engine for Go.
      = @children
}

@hmlt HomePage() {
  = @render SiteLayout("Home Page")
    %p This is the home page for Hamlet.
}
```

Your next step will be to process the Hamlet file to parse the Haml and generate the Go code using the Hamlet [CLI](#hamlet-cli) tool:
```sh
hamlet generate
```

Use the generated Go code to render HTML in your application:
```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_ = HomePage().Render(r.Context(), w)
	})

	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
```
Which would serve the following HTML:
```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <title>Hamlet</title>
  </head>
  <body>
    <h1>Home Page</h1>
    <p>A type-safe Haml template engine for Go.</p>
    <p>This is the home page for Hamlet.</p>
  </body>
</html>
```

## Supported Haml Syntax & Features
- [x] Doctypes (`!!!`)
- [x] Tags (`%tag`)
- [x] Attributes (`{name: value}`) [(more info)](#attributes)
- [x] Classes and IDs (`.class` `#id`) [(more info)](#classes)
- [x] Object References (`[obj]`) [(more info)](#object-references)
- [x] Unescaped Text (`!` `!=`)
- [x] Comments (`/` `-#`)
- [x] Inline Interpolation (`#{value}`)
- [x] Inlining Code (`- code`)
- [x] Rendering Code (`= code`)
- [x] Filters (`:plain`, ...) [(more info)](#filters)
- [x] Whitespace Removal (`%tag>` `%tag<`) [(more info)](#whitespace-removal)

### Unsupported Haml Features
- [ ] Probably something I've missed, please raise an issue if you find something missing.

## Hamlet CLI

### Installation
```sh
go install github.com/stackus/hamlet/cmd/hamlet@latest
```

### Usage
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
When you are using Hamlet you will typically be dealing with the generated Go code, and not the Hamlet runtime directly.
However, if you need to install the Hamlet library, you can do so with:
```sh
go get github.com/stackus/hamlet
```

## Using Hamlet
To start using Hamlet, the first step is to create a Hamlet file with one or more Haml templates.
If you need guidance, the section [The Hamlet template](#the-hamlet-template) has all the information you need.

With your Hamlet files written, the next step involves generating Go code from them.
The [CLI](#hamlet-cli) tool handles this generation step.
It's a straightforward process that converts your Hamlet files and templates into ready to run Go files.

Each generated Go file will include a function corresponding to each of your templates.
The names of the functions are not altered at all,
if you want them to be exported in Go then you need to use an uppercase letter for the first character of the template name.

When this function is executed, it yields a `*hamlet.Template`.
This is what you'll use to render your templates in the application.

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
The above would render the `RemoveWhitespace` example from the [examples](/examples) directory in this repository,
and would output the following:
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
Hamlet templates are files with the extension `.hmlt` that when processed will produce a matching Go file with the extension `.hmlt.go`.

In these files you are free to write any Go code that you wish, and then drop into Haml mode using the `@hmlt` directive.

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
The Haml syntax is documented at the [Haml](http://haml.info/) website.
Please see that site or the [Haml Reference](https://haml.info/docs/yardoc/file.REFERENCE.html) for more information.

Hamlet has implemented the Haml syntax very closely.
So, if you are already familiar with Haml then you should be able to jump right in.
There are some minor differences that I will document in the next section.

### Hamlet and Haml differences

Important differences are:
- [Go package and imports](#go-package-and-imports): You can declare a package and imports for your templates.
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
