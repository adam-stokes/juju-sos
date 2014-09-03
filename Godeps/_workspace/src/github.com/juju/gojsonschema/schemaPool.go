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

// author           sigu-399
// author-github    https://github.com/sigu-399
// author-mail      sigu.399@gmail.com
//
// repository-name  gojsonschema
// repository-desc  An implementation of JSON Schema, based on IETF's draft v4 - Go language.
//
// description		Defines resources pooling.
//                  Eases referencing and avoids downloading the same resource twice.
//
// created          26-02-2013

package gojsonschema

import (
	"errors"
	"fmt"
	"strings"

	"github.com/juju/gojsonreference"
)

type schemaPoolDocument struct {
	Document interface{}
}

type schemaPool struct {
	schemaPoolDocuments map[string]*schemaPoolDocument
	standaloneDocument  interface{}
}

func newSchemaPool() *schemaPool {

	p := &schemaPool{}
	p.schemaPoolDocuments = make(map[string]*schemaPoolDocument)
	p.standaloneDocument = nil

	return p
}

func (p *schemaPool) SetStandaloneDocument(document interface{}) {
	p.standaloneDocument = document
}

func (p *schemaPool) GetStandaloneDocument() (document interface{}) {
	return p.standaloneDocument
}

func (p *schemaPool) GetDocument(reference gojsonreference.JsonReference) (*schemaPoolDocument, error) {

	internalLog(fmt.Sprintf("Get document from pool (%s) :", reference.String()))

	var err error

	// It is not possible to load anything that is not canonical...
	if !reference.GetUrl().IsAbs() {
		return nil, errors.New(fmt.Sprintf("Reference must be canonical %s", reference))
	}

	refToUrl := reference
	refToUrl.GetUrl().Fragment = ""

	var spd *schemaPoolDocument

	// Try to find the requested document in the pool
	for k := range p.schemaPoolDocuments {
		if k == refToUrl.String() {
			spd = p.schemaPoolDocuments[k]
		}
	}

	if spd != nil {
		internalLog(" Found in pool")
		return spd, nil
	}

	// Load the document

	var document interface{}

	if reference.GetUrl().Scheme == "file" {

		internalLog(" Loading new document from file")

		// Load from file
		filename := strings.Replace(refToUrl.String(), "file://", "", -1)
		document, err = GetFileJson(filename)
		if err != nil {
			return nil, err
		}

	} else if reference.GetUrl().Scheme == "http" {

		internalLog(" Loading new document from http")

		// Load from HTTP
		document, err = GetHttpJson(refToUrl.String())
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("unhandled scheme " + reference.GetUrl().Scheme)
	}

	spd = &schemaPoolDocument{Document: document}
	// add the document to the pool for potential later use
	p.schemaPoolDocuments[refToUrl.String()] = spd

	return spd, nil
}
