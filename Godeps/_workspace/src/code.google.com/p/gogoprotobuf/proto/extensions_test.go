// Go support for Protocol Buffers - Google's data interchange format
//
// Copyright 2014 The Go Authors.  All rights reserved.
// http://code.google.com/p/goprotobuf/
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//     * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package proto_test

import (
	"testing"

	pb "./testdata"
	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/code.google.com/p/gogoprotobuf/proto"
)

func TestGetExtensionsWithMissingExtensions(t *testing.T) {
	msg := &pb.MyMessage{}
	ext1 := &pb.Ext{}
	if err := proto.SetExtension(msg, pb.E_Ext_More, ext1); err != nil {
		t.Fatalf("Could not set ext1: %s", ext1)
	}
	exts, err := proto.GetExtensions(msg, []*proto.ExtensionDesc{
		pb.E_Ext_More,
		pb.E_Ext_Text,
	})
	if err != nil {
		t.Fatalf("GetExtensions() failed: %s", err)
	}
	if exts[0] != ext1 {
		t.Errorf("ext1 not in returned extensions: %T %v", exts[0], exts[0])
	}
	if exts[1] != nil {
		t.Errorf("ext2 in returned extensions: %T %v", exts[1], exts[1])
	}
}
