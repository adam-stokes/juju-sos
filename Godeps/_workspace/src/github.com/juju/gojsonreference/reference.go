// Copyright 2013 sigu-399 ( https://github.com/sigu-399 )
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// author  			sigu-399
// author-github 	https://github.com/sigu-399
// author-mail		sigu.399@gmail.com
//
// repository-name	gojsonreference
// repository-desc	An implementation of JSON Reference - Go language
//
// description		Main and unique file.
//
// created      	26-02-2013

package gojsonreference

import (
	"errors"
	"net/url"
	"strings"

	"github.com/juju/gojsonpointer"
)

const (
	const_fragment_char = `#`
)

func NewJsonReference(jsonReferenceString string) (JsonReference, error) {
	var r JsonReference

	var err error

	r.referenceUrl, err = url.Parse(jsonReferenceString)
	if err != nil {
		return JsonReference{}, err
	}

	r.referencePointer, err = gojsonpointer.NewJsonPointer(r.referenceUrl.Fragment)
	if err != nil {
		return JsonReference{}, err
	}

	return r, err
}

type JsonReference struct {
	referenceUrl     *url.URL
	referencePointer gojsonpointer.JsonPointer
}

func (r *JsonReference) GetUrl() *url.URL {
	return r.referenceUrl
}

func (r *JsonReference) GetPointer() *gojsonpointer.JsonPointer {
	return &r.referencePointer
}

func (r *JsonReference) String() string {

	if r.referenceUrl != nil {
		return r.referenceUrl.String()
	}

	return r.referencePointer.String()
}

// Creates a new reference from a parent and a child
// If the child cannot inherit from the parent, an error is returned
func (r *JsonReference) Inherits(child JsonReference) (*JsonReference, error) {

	if r.referenceUrl == nil {
		return nil, errors.New("parent reference nil")
	}

	if !r.referenceUrl.IsAbs() {
		return nil, errors.New("parent reference must be absolute URL.")
	}

	if r.referenceUrl.Scheme != "http" && r.referenceUrl.Scheme != "file" {
		return nil, errors.New("scheme type " + r.referenceUrl.Scheme + " not handled")
	}

	if child.referenceUrl.IsAbs() {
		if child.referenceUrl.Scheme != r.referenceUrl.Scheme {
			return nil, errors.New("scheme of child " + child.String() +
				" incompatible with scheme of parent " + r.String())
		}

		if r.referenceUrl.Host != child.referenceUrl.Host {
			return nil, errors.New("references have different hosts")
		}
	}

	inheritedReference, err := NewJsonReference(r.String())
	if err != nil {
		return nil, err
	}

	// Child reference is not a fragment, and has a different path than parent
	//if child.referenceUrl != nil && child.referenceUrl.Path != "" {
	if child.referenceUrl.Path != "" {
		if !strings.HasPrefix(child.referenceUrl.Path, r.referenceUrl.Path) {
			return nil, errors.New("child reference " + child.String() +
				" has divergent path " + child.referenceUrl.Path +
				" from parent " + r.String() +
				", which has path " + r.referenceUrl.Path)
		}

		inheritedReference.referenceUrl.Path = child.referenceUrl.Path
	}

	if child.referenceUrl != nil && child.referenceUrl.Fragment != "" {
		if !strings.HasPrefix(child.referenceUrl.Fragment, r.referenceUrl.Fragment) {
			return nil, errors.New("child reference " + child.String() +
				" has divergent pointer " + child.referenceUrl.Fragment +
				" from parent " + r.String() +
				", which has pointer " + r.referenceUrl.Fragment)
		}

		inheritedReference.referenceUrl.Fragment = child.referenceUrl.Fragment
		inheritedReference.referencePointer = child.referencePointer
	}

	return &inheritedReference, nil
}
