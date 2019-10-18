package tagging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"oracle.com/oci/tagging/protobuf"
)

var (
	invalidSlug = TagSlug("thisisnotavalidprotobufslug")
	val1        = "val1"
	val2        = "val2"
	val3        = "val3"
)

func TestNewTagSlug(t *testing.T) {
	type args struct {
		freeformTags FreeformTagSet
		definedTags  DefinedTagSet
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		wantFreeform FreeformTagSet
		wantDefined  DefinedTagSet
	}{
		{
			name: "creates a slug with freeform tags and nil defined tags",
			args: args{
				freeformTags: FreeformTagSet{"key1": "val1"},
			},
			wantFreeform: FreeformTagSet{"key1": "val1"},
			wantDefined:  DefinedTagSet{},
		},
		{
			name: "creates a slug with freeform tags and empty defined tags",
			args: args{
				freeformTags: FreeformTagSet{"key1": "val1"},
				definedTags:  DefinedTagSet{},
			},
			wantFreeform: FreeformTagSet{"key1": "val1"},
			wantDefined:  DefinedTagSet{},
		},
		{
			name: "creates a slug with defined tags and nil freeform tags",
			args: args{
				definedTags: DefinedTagSet{
					"namespace1": map[string]string{
						"key1": "val1",
						"key2": "val2",
					},
				},
			},
			wantFreeform: FreeformTagSet{},
			wantDefined: DefinedTagSet{
				"namespace1": map[string]string{
					"key1": "val1",
					"key2": "val2",
				},
			},
		},
		{
			name: "creates a slug with defined tags and empty freeform tags",
			args: args{
				freeformTags: FreeformTagSet{},
				definedTags: DefinedTagSet{
					"namespace1": map[string]string{
						"key1": "val1",
						"key2": "val2",
					},
				},
			},
			wantFreeform: FreeformTagSet{},
			wantDefined: DefinedTagSet{
				"namespace1": map[string]string{
					"key1": "val1",
					"key2": "val2",
				},
			},
		},
		{
			name: "creates a slug with freeform and defined tags",
			args: args{
				freeformTags: FreeformTagSet{"key1": "val1"},
				definedTags: DefinedTagSet{
					"namespace1": map[string]string{
						"key2": "val2",
						"key3": "val3",
					},
				},
			},
			wantFreeform: FreeformTagSet{"key1": "val1"},
			wantDefined: DefinedTagSet{
				"namespace1": map[string]string{
					"key2": "val2",
					"key3": "val3",
				},
			},
		},
		{
			name: "creates a slug with freeform and multi namespace defined tags",
			args: args{
				freeformTags: FreeformTagSet{"key1": "val1"},
				definedTags: DefinedTagSet{
					"namespace1": map[string]string{
						"key2": "val2",
						"key3": "val3",
					},
					"namespace2": map[string]string{
						"key1": "val1",
						"key3": "val3",
					},
				},
			},
			wantFreeform: FreeformTagSet{"key1": "val1"},
			wantDefined: DefinedTagSet{
				"namespace1": map[string]string{
					"key2": "val2",
					"key3": "val3",
				},
				"namespace2": map[string]string{
					"key1": "val1",
					"key3": "val3",
				},
			},
		},
		{
			name: "returns an empty tag slug for empty tag sets",
			args: args{
				freeformTags: FreeformTagSet{},
				definedTags:  DefinedTagSet{},
			},
			wantFreeform: FreeformTagSet{},
			wantDefined:  DefinedTagSet{},
		},
		{
			name: "returns an empty tag slug for nil tag sets",
			args: args{
				freeformTags: nil,
				definedTags:  nil,
			},
			wantFreeform: FreeformTagSet{},
			wantDefined:  DefinedTagSet{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTagSlug(tt.args.freeformTags, tt.args.definedTags)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, got)

			wantSlug, err := unmarshal(*got)
			assert.NoError(t, err)
			assert.NotNil(t, wantSlug)

			freeSet, err := got.FreeformTagSet()
			assert.NoError(t, err)
			assert.Equal(t, tt.wantFreeform, *freeSet)

			definedSet, err := got.DefinedTagSet()
			assert.NoError(t, err)
			assert.Equal(t, tt.wantDefined, *definedSet)
		})
	}
}

func TestSlug_FreeformTagSet(t *testing.T) {
	tests := []struct {
		name    string
		slug    *protobuf.Slug
		want    *FreeformTagSet
		wantErr bool
	}{
		{
			name: "returns an empty set for an empty slug",
			slug: &protobuf.Slug{FreeformTags: &protobuf.StringMap{}},
			want: &FreeformTagSet{},
		},
		{
			name: "returns for defined tags with a single key",
			slug: &protobuf.Slug{
				FreeformTags: &protobuf.StringMap{
					Entries: FreeformTagSet{"key1": "val1"},
				},
			},
			want: &FreeformTagSet{"key1": "val1"},
		},
		{
			name: "returns for defined tags with multiple keys",
			slug: &protobuf.Slug{
				FreeformTags: &protobuf.StringMap{
					Entries: FreeformTagSet{
						"key1": "val1",
						"key2": "val2",
						"key3": "val3",
					},
				},
			},
			want: &FreeformTagSet{
				"key1": "val1",
				"key2": "val2",
				"key3": "val3",
			},
		},
		{
			name:    "returns an error for an invalid slug",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// An invalid slug should return an error and nil tag set
			if tt.wantErr {
				ts := &invalidSlug
				got, err := ts.FreeformTagSet()
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				ts, err := Marshal(tt.slug)
				assert.NoError(t, err)
				assert.NotNil(t, ts)
				got, err := ts.FreeformTagSet()
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSlug_SwaggerFreeformTagSet(t *testing.T) {
	tests := []struct {
		name    string
		slug    *protobuf.Slug
		want    map[string]*string
		wantErr bool
	}{
		{
			name: "returns an empty set for an empty slug",
			slug: &protobuf.Slug{FreeformTags: &protobuf.StringMap{}},
			want: map[string]*string{},
		},
		{
			name: "returns for defined tags with a single key",
			slug: &protobuf.Slug{
				FreeformTags: &protobuf.StringMap{
					Entries: FreeformTagSet{"key1": "val1"},
				},
			},
			want: map[string]*string{"key1": &val1},
		},
		{
			name: "returns for defined tags with multiple keys",
			slug: &protobuf.Slug{
				FreeformTags: &protobuf.StringMap{
					Entries: FreeformTagSet{
						"key1": "val1",
						"key2": "val2",
						"key3": "val3",
					},
				},
			},
			want: map[string]*string{
				"key1": &val1,
				"key2": &val2,
				"key3": &val3,
			},
		},
		{
			name:    "returns an error for an invalid slug",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// An invalid slug should return an error and nil tag set
			if tt.wantErr {
				ts := &invalidSlug
				got, err := ts.SwaggerFreeformTagSet()
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				ts, err := Marshal(tt.slug)
				assert.NoError(t, err)
				assert.NotNil(t, ts)
				got, err := ts.SwaggerFreeformTagSet()
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSlug_DefinedTagSet(t *testing.T) {
	tests := []struct {
		name    string
		slug    *protobuf.Slug
		want    *DefinedTagSet
		wantErr bool
	}{
		{
			name: "returns an empty set for an empty slug",
			slug: &protobuf.Slug{DefinedTags: &protobuf.NamespaceMap{Namespaces: map[string]*protobuf.NamespaceDefinedTags{}}},
			want: &DefinedTagSet{},
		},
		{
			name: "returns for defined tags with a single key",
			slug: &protobuf.Slug{
				DefinedTags: &protobuf.NamespaceMap{
					Namespaces: map[string]*protobuf.NamespaceDefinedTags{
						"ns1": {
							Tags: map[string]*protobuf.DefinedTagValue{
								"key1": {Value: []string{"val1"}},
							},
						},
					},
				},
			},
			want: &DefinedTagSet{"ns1": map[string]string{"key1": "val1"}},
		},
		{
			name: "returns for defined tags with multiple keys",
			slug: &protobuf.Slug{
				DefinedTags: &protobuf.NamespaceMap{
					Namespaces: map[string]*protobuf.NamespaceDefinedTags{
						"ns1": {
							Tags: map[string]*protobuf.DefinedTagValue{
								"key1": {Value: []string{"val1"}},
								"key2": {Value: []string{"val2"}},
							},
						},
						"ns2": {
							Tags: map[string]*protobuf.DefinedTagValue{
								"key2": {Value: []string{"val2"}},
								"key3": {Value: []string{"val3"}},
							},
						},
					},
				},
			},
			want: &DefinedTagSet{
				"ns1": map[string]string{
					"key1": "val1",
					"key2": "val2",
				},
				"ns2": map[string]string{
					"key2": "val2",
					"key3": "val3",
				},
			},
		},
		{
			name:    "returns an error for an invalid slug",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// An invalid slug should return an error and nil tag set
			if tt.wantErr {
				ts := &invalidSlug
				got, err := ts.DefinedTagSet()
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				ts, err := Marshal(tt.slug)
				assert.NoError(t, err)
				assert.NotNil(t, ts)
				got, err := ts.DefinedTagSet()
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSlug_SwaggerDefinedTagSet(t *testing.T) {
	tests := []struct {
		name    string
		slug    *protobuf.Slug
		want    map[string]map[string]interface{}
		wantErr bool
	}{
		{
			name: "returns an empty set for an empty slug",
			slug: &protobuf.Slug{DefinedTags: &protobuf.NamespaceMap{Namespaces: map[string]*protobuf.NamespaceDefinedTags{}}},
			want: map[string]map[string]interface{}{},
		},
		{
			name: "returns for defined tags with a single key",
			slug: &protobuf.Slug{
				DefinedTags: &protobuf.NamespaceMap{
					Namespaces: map[string]*protobuf.NamespaceDefinedTags{
						"ns1": {
							Tags: map[string]*protobuf.DefinedTagValue{
								"key1": {Value: []string{"val1"}},
							},
						},
					},
				},
			},
			want: map[string]map[string]interface{}{
				"ns1": {"key1": "val1"},
			},
		},
		{
			name: "returns for defined tags with multiple keys",
			slug: &protobuf.Slug{
				DefinedTags: &protobuf.NamespaceMap{
					Namespaces: map[string]*protobuf.NamespaceDefinedTags{
						"ns1": {
							Tags: map[string]*protobuf.DefinedTagValue{
								"key1": {Value: []string{"val1"}},
								"key2": {Value: []string{"val2"}},
							},
						},
						"ns2": {
							Tags: map[string]*protobuf.DefinedTagValue{
								"key2": {Value: []string{"val2"}},
								"key3": {Value: []string{"val3"}},
							},
						},
					},
				},
			},
			want: map[string]map[string]interface{}{
				"ns1": {
					"key1": "val1",
					"key2": "val2",
				},
				"ns2": {
					"key2": "val2",
					"key3": "val3",
				},
			},
		},
		{
			name:    "returns an error for an invalid slug",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// An invalid slug should return an error and nil tag set
			if tt.wantErr {
				ts := &invalidSlug
				got, err := ts.SwaggerDefinedTagSet()
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				ts, err := Marshal(tt.slug)
				assert.NoError(t, err)
				assert.NotNil(t, ts)
				got, err := ts.SwaggerDefinedTagSet()
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCopyTagSlug(t *testing.T) {
	testSlug, _ := NewTagSlug(
		FreeformTagSet{"key1": "val1"},
		DefinedTagSet{
			"namespace1": map[string]string{
				"key1": "val1",
				"key2": "val2",
			},
		},
	)
	testIO := []struct {
		name      string
		input     *TagSlug
		expectNil bool
	}{
		{
			name:      "returns nil for a nil tag slug",
			input:     nil,
			expectNil: true,
		},
		{
			name:  "copies a non-nil slug",
			input: testSlug,
		},
	}

	for _, test := range testIO {
		t.Run(test.name, func(t *testing.T) {
			newSlug := test.input.Copy()
			if test.expectNil {
				assert.Nil(t, newSlug)
			} else {
				assert.Equal(t, test.input, newSlug)
				assert.False(t, test.input == newSlug, "pointers were not different")
			}

		})
	}
}

func TestMarshal(t *testing.T) {
	type args struct {
		protobufSlug *protobuf.Slug
	}
	tests := []struct {
		name    string
		args    args
		want    *TagSlug
		wantErr bool
	}{
		{
			name: "returns an error for a nil protobuf slug",
			args: args{
				protobufSlug: nil,
			},
			wantErr: true,
		},
		{
			name: "prepends the header to an empty protobuf slug",
			args: args{
				protobufSlug: &protobuf.Slug{},
			},
			want: &TagSlug{0x1, 0x3, 0x2, 0x2},
		},
		{
			name: "prepends the header to a protobuf slug w/ freeform tags",
			args: args{
				protobufSlug: &protobuf.Slug{
					FreeformTags: &protobuf.StringMap{
						Entries: FreeformTagSet{"key1": "val1"},
					},
				},
			},
			want: &TagSlug{0x1, 0x3, 0x2, 0x2, 0xa, 0xe, 0xa, 0xc, 0xa, 0x4, 0x6b, 0x65, 0x79, 0x31, 0x12, 0x4, 0x76, 0x61, 0x6c, 0x31},
		},
		{
			name: "prepends the header to a protobuf slug w/ defined tags",
			args: args{
				protobufSlug: &protobuf.Slug{
					DefinedTags: &protobuf.NamespaceMap{
						Namespaces: map[string]*protobuf.NamespaceDefinedTags{
							"ns1": {
								Tags: map[string]*protobuf.DefinedTagValue{
									"key1": {Value: []string{"val1"}},
								},
							},
						},
					},
				},
			},
			want: &TagSlug{0x1, 0x3, 0x2, 0x2, 0x12, 0x19, 0xa, 0x17, 0xa, 0x3, 0x6e, 0x73, 0x31, 0x12, 0x10, 0xa, 0xe, 0xa, 0x4, 0x6b, 0x65, 0x79, 0x31, 0x12, 0x6, 0xa, 0x4, 0x76, 0x61, 0x6c, 0x31},
		},
		{
			name: "prepends the header to a protobuf slug w/ freeform and defined tags",
			args: args{
				protobufSlug: &protobuf.Slug{
					FreeformTags: &protobuf.StringMap{
						Entries: FreeformTagSet{"key1": "val1"},
					},
					DefinedTags: &protobuf.NamespaceMap{
						Namespaces: map[string]*protobuf.NamespaceDefinedTags{
							"ns1": {
								Tags: map[string]*protobuf.DefinedTagValue{
									"key1": {Value: []string{"val1"}},
								},
							},
						},
					},
				},
			},
			want: &TagSlug{0x1, 0x3, 0x2, 0x2, 0xa, 0xe, 0xa, 0xc, 0xa, 0x4, 0x6b, 0x65, 0x79, 0x31, 0x12, 0x4, 0x76, 0x61, 0x6c, 0x31, 0x12, 0x19, 0xa, 0x17, 0xa, 0x3, 0x6e, 0x73, 0x31, 0x12, 0x10, 0xa, 0xe, 0xa, 0x4, 0x6b, 0x65, 0x79, 0x31, 0x12, 0x6, 0xa, 0x4, 0x76, 0x61, 0x6c, 0x31},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.args.protobufSlug)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	type args struct {
		slug TagSlug
	}
	tests := []struct {
		name    string
		args    args
		want    *protobuf.Slug
		wantErr bool
	}{
		{
			name: "return an empty protobuf slug a nil tag slug",
			args: args{
				slug: nil,
			},
			want: &protobuf.Slug{},
		},
		{
			name: "returns an error for an invalid tag slug",
			args: args{
				slug: TagSlug{1, 2, 3},
			},
			wantErr: true,
		},
		{
			name: "returns an error for an invalid tag slug with headers",
			args: args{
				slug: TagSlug{0x1, 0x3, 0x2, 0x2, 1, 2, 3},
			},
			wantErr: true,
		},
		{
			name: "return an empty protobuf slug for an empty tag slug",
			args: args{
				slug: TagSlug{0x1, 0x3, 0x2, 0x2},
			},
			want: &protobuf.Slug{},
		},
		{
			name: "decodes a tag slug with freeform tags",
			args: args{
				slug: TagSlug{0x1, 0x3, 0x2, 0x2, 0xa, 0xe, 0xa, 0xc, 0xa, 0x4, 0x6b, 0x65, 0x79, 0x31, 0x12, 0x4, 0x76, 0x61, 0x6c, 0x31},
			},
			want: &protobuf.Slug{
				FreeformTags: &protobuf.StringMap{
					Entries: FreeformTagSet{"key1": "val1"},
				},
			},
		},
		{
			name: "decodes a tag slug with defined tags",
			args: args{
				slug: TagSlug{0x1, 0x3, 0x2, 0x2, 0x12, 0x19, 0xa, 0x17, 0xa, 0x3, 0x6e, 0x73, 0x31, 0x12, 0x10, 0xa, 0xe, 0xa, 0x4, 0x6b, 0x65, 0x79, 0x31, 0x12, 0x6, 0xa, 0x4, 0x76, 0x61, 0x6c, 0x31},
			},
			want: &protobuf.Slug{
				DefinedTags: &protobuf.NamespaceMap{
					Namespaces: map[string]*protobuf.NamespaceDefinedTags{
						"ns1": {
							Tags: map[string]*protobuf.DefinedTagValue{
								"key1": {Value: []string{"val1"}},
							},
						},
					},
				},
			},
		},
		{
			name: "decodes a tag slug with freeform and defined tags",
			args: args{
				slug: TagSlug{0x1, 0x3, 0x2, 0x2, 0xa, 0xe, 0xa, 0xc, 0xa, 0x4, 0x6b, 0x65, 0x79, 0x31, 0x12, 0x4, 0x76, 0x61, 0x6c, 0x31, 0x12, 0x19, 0xa, 0x17, 0xa, 0x3, 0x6e, 0x73, 0x31, 0x12, 0x10, 0xa, 0xe, 0xa, 0x4, 0x6b, 0x65, 0x79, 0x31, 0x12, 0x6, 0xa, 0x4, 0x76, 0x61, 0x6c, 0x31},
			},
			want: &protobuf.Slug{
				FreeformTags: &protobuf.StringMap{
					Entries: FreeformTagSet{"key1": "val1"},
				},
				DefinedTags: &protobuf.NamespaceMap{
					Namespaces: map[string]*protobuf.NamespaceDefinedTags{
						"ns1": {
							Tags: map[string]*protobuf.DefinedTagValue{
								"key1": {Value: []string{"val1"}},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshal(tt.args.slug)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestAddSlugHeaders(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "adds headers for an empty slug",
			args: args{[]byte{}},
			want: []byte{1, 3, 2, 2},
		},
		{
			name: "adds headers for a non-empty slug",
			args: args{[]byte{0xa, 0xe, 0xa, 0xc, 0xa, 0x4, 0x6b, 0x65, 0x79, 0x31, 0x12, 0x4, 0x76, 0x61, 0x6c, 0x31}},
			want: []byte{1, 3, 2, 2, 0xa, 0xe, 0xa, 0xc, 0xa, 0x4, 0x6b, 0x65, 0x79, 0x31, 0x12, 0x4, 0x76, 0x61, 0x6c, 0x31},
		},
		{
			name: "does not modify an empty slug with existing headers",
			args: args{[]byte{1, 3, 2, 2}},
			want: []byte{1, 3, 2, 2},
		},
		{
			name: "does not modify non-empty slug with existing headers",
			args: args{[]byte{1, 3, 2, 2, 0xa, 0xe, 0xa, 0xc, 0xa, 0x4, 0x6b, 0x65, 0x79, 0x31, 0x12, 0x4, 0x76, 0x61, 0x6c, 0x31}},
			want: []byte{1, 3, 2, 2, 0xa, 0xe, 0xa, 0xc, 0xa, 0x4, 0x6b, 0x65, 0x79, 0x31, 0x12, 0x4, 0x76, 0x61, 0x6c, 0x31},
		},
		{
			name:    "returns an error for an invalid protobuf slug with existing headers",
			args:    args{[]byte{1, 3, 2, 2, 1, 3, 2, 2, 0xa, 0xe, 0xa, 0xc, 0xa, 0x4, 0x6b, 0x65, 0x79, 0x31, 0x12, 0x4, 0x76, 0x61, 0x6c, 0x31}},
			wantErr: true,
		},
		{
			name:    "returns an error for an invalid protobuf slug",
			args:    args{[]byte("an invalid protobuf slug")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddSlugHeaders(tt.args.b)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, TagSlug(tt.want), got)
			}
		})
	}
}
