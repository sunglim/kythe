/*
 * Copyright 2016 Google Inc. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package ptypes is a thin wrapper around the golang.org/protobuf/ptypes
// package that adds support for Kythe message types, and handles some type
// format conversions.
package ptypes

import (
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	anypb "kythe.io/third_party/proto/any_proto"
)

// MarshalAny converts pb to a google.protobuf.Any message, fixing the URLs of
// Kythe protobuf types as needed.
func MarshalAny(pb proto.Message) (*anypb.Any, error) {
	// The ptypes package vendors generated code for the Any type, so we have
	// to convert the type. The pointers are convertible, but since we need to
	// do surgery on the URL anyway, we just construct the output separately.
	internalAny, err := ptypes.MarshalAny(pb)
	if err != nil {
		return nil, err
	}

	// Fix up messages in the Kythe namespace.
	url := internalAny.TypeUrl
	if name, _ := ptypes.AnyMessageName(internalAny); strings.HasPrefix(name, "kythe.") {
		url = "kythe.io/proto/" + name
	}
	return &anypb.Any{
		TypeUrl: url,
		Value:   internalAny.Value,
	}, nil
}
