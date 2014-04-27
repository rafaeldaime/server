package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/go-martini/martini"
)

// An Encoder implements an encoding format of values to be sent as response to
// requests on the API endpoints.
type Encoder interface {
	Encode(v ...interface{}) (string, error)
}

// The regex to check for the requested format (allows an optional trailing
// slash).
var rxExt = regexp.MustCompile(`(\.(?:xml|text|json|html))\/?$`)

// EncoderMiddleware intercepts the request's URL, detects the requested format,
// and injects the correct encoder dependency for this request. It rewrites
// the URL to remove the format extension, so that routes can be defined
// without it.
func EncoderMiddleware(c martini.Context, w http.ResponseWriter, r *http.Request) {
	// Here we will consider .json de default,
	// except when the main path (/) is accessed
	ft := ".json"
	if r.URL.Path == "/" {
		ft = ".html"
	}

	// Get the format extension
	matches := rxExt.FindStringSubmatch(r.URL.Path)
	if len(matches) > 1 {
		// Rewrite the URL without the format extension
		l := len(r.URL.Path) - len(matches[1])
		if strings.HasSuffix(r.URL.Path, "/") {
			l--
		}
		ft = matches[1]

		// *** Remove the extension if it is not an .html extension
		if ft != ".html" {
			r.URL.Path = r.URL.Path[:l]
		}
	}
	// Inject the requested encoder
	switch ft {
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".xml":
		c.MapTo(xmlEncoder{}, (*Encoder)(nil))
		w.Header().Set("Content-Type", "application/xml")
	case ".text":
		c.MapTo(textEncoder{}, (*Encoder)(nil))
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	default:
		c.MapTo(jsonEncoder{}, (*Encoder)(nil))
		w.Header().Set("Content-Type", "application/json")
	}
}

// Because `panic`s are caught by martini's Recovery handler, it can be used
// to return server-side errors (500). Some helpful text message should probably
// be sent, although not the technical error (which is printed in the log).
func Must(data string, err error) string {
	if err != nil {
		panic(err)
	}
	return data
}

type jsonEncoder struct{}

// jsonEncoder is an Encoder that produces JSON-formatted responses.
func (_ jsonEncoder) Encode(v ...interface{}) (string, error) {
	/*
		var data interface{} = v
		var result interface{}

		if v == nil {
			// So that empty results produces `[]` and not `null`
			data = []interface{}{}
		} else if len(v) == 1 {
			data = v[0]
		}

		t := reflect.TypeOf(data)

		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		switch t.Kind() {
		case reflect.Slice:
			result = iterateSlice(reflect.ValueOf(data)).Interface()

		case reflect.Struct:
			result = copyStruct(reflect.ValueOf(data)).Interface()

		default:
			result = data
		}




		return string(buffer), err
	*/
	var buffer, err = json.MarshalIndent(v, "", "    ")

	if err != nil {
		return "", err
	}
	return string(buffer), err
}

type xmlEncoder struct{}

// xmlEncoder is an Encoder that produces XML-formatted responses.
func (_ xmlEncoder) Encode(v ...interface{}) (string, error) {
	var data interface{} = v
	var buffer bytes.Buffer

	if v == nil {
		data = []interface{}{}
	} else if len(v) == 1 {
		data = v[0]
	}

	if _, err := buffer.Write([]byte(xml.Header)); err != nil {
		return "", err
	}

	b, err := xml.MarshalIndent(data, "", "    ")
	if err != nil {
		return "", err
	}

	buffer.Write(b)

	return buffer.String(), nil
}

type textEncoder struct{}

// textEncoder is an Encoder that produces plain text-formatted responses.
func (_ textEncoder) Encode(v ...interface{}) (string, error) {
	var buf bytes.Buffer
	for _, v := range v {
		if _, err := fmt.Fprintf(&buf, "%s\n", v); err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

func copyStruct(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	result := reflect.New(v.Type()).Elem()

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		vfield := v.Field(i)

		if tag := t.Field(i).Tag.Get("out"); tag == "false" {
			continue
		}

		if vfield.Type() == reflect.TypeOf(time.Time{}) {
			result.Field(i).Set(vfield)
			continue
		}

		if vfield.Kind() == reflect.Interface && vfield.Interface() != nil {
			vfield = vfield.Elem()

			for vfield.Kind() == reflect.Ptr {
				vfield = vfield.Elem()
			}

			result.Field(i).Set(copyStruct(vfield))
			continue
		}

		if vfield.Kind() == reflect.Struct || vfield.Kind() == reflect.Ptr {
			r := copyStruct(vfield)

			if result.Field(i).Kind() == reflect.Ptr {
				result.Field(i).Set(reflect.New(r.Type()))
			} else {
				result.Field(i).Set(r)
			}

			continue
		}

		if vfield.Kind() == reflect.Array || vfield.Kind() == reflect.Slice {
			result.Field(i).Set(iterateSlice(vfield))
			continue
		}

		if result.Field(i).CanSet() {
			result.Field(i).Set(vfield)
		}
	}

	return result
}

func iterateSlice(v reflect.Value) reflect.Value {
	result := reflect.MakeSlice(v.Type(), 0, v.Len())

	for i := 0; i < v.Len(); i++ {
		value := v.Index(i)
		result = reflect.Append(result, copyStruct(value))
	}

	return result
}
