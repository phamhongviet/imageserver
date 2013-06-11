package source

import (
	"github.com/pierrre/imageserver"
	"net/http"
)

type SourceParser struct {
}

func (parser *SourceParser) Parse(request *http.Request) (parameters imageserver.Parameters, err error) {
	parameters = make(imageserver.Parameters)

	query := request.URL.Query()

	source := query.Get("source")
	if len(source) > 0 {
		parameters.Set("source", source)
	}

	return
}