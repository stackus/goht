![hamlet_header.png](docs/hamlet_header.png)

# Hamlet
A [HAML](http://haml.info/)-like template engine for Go.

## Features
- HAML-like syntax
- Templates are compiled to type-safe Go
- Multiple templates per file
- Mix Go and templates together in the same file
- Easy nesting of templates

```
@hmlt SiteLayout() {
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
		%p This is the home page.
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
    <p>A HAML-like template engine for Go.</p>
    <p>This is the home page.</p>
  </body>
</html>
```
