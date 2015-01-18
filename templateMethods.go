package glaze

import "html/template"

// SafeHTML will return the given string as a template.HTML object,
// which will cause the template package to skip escaping.
func SafeHTML(text string) template.HTML {
	return template.HTML(text)
}
