package views

import "html/template"

var List = template.Must(template.New("list").Parse(list))
