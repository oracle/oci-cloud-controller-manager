package tagging

import (
	"bytes"

	proto "github.com/golang/protobuf/proto"
	"oracle.com/oci/tagging/protobuf"
)

// FreeformTagSet is a map of arbitrary keys to values.
type FreeformTagSet map[string]string

// SwaggerFreeformTagSet is the go-swagger representation of freeform tags.
type SwaggerFreeformTagSet map[string]*string

// DefinedTagSet is a nested map of tag namespaces to defined tag key/value pairs.
type DefinedTagSet map[string]map[string]string

// SwaggerDefinedTagSet is the go-swagger representation of defined tags.
type SwaggerDefinedTagSet map[string]map[string]interface{}

// TagSlug is a protobuf-serialized byte slice of a `FreeformTagSet` and `DefinedTagSet`.
type TagSlug []byte

// ProtobufHeaders is a 4-byte slice prepended to the protobuf tag
// slug before sending to IAM. It declares the version, serialization
// format, encoding algorithm, and compression algorithm.
var ProtobufHeaders = []byte{
	1, // Version
	3, // Serialization format (3 == protobuf)
	2, // Encoding algorithm (2 == none)
	2, // Compression algorithm (2 == none)
}

// NewTagSlug returns a protobuf-marshaled byte slice of the freeform and defined tag sets.
func NewTagSlug(freeformTags FreeformTagSet, definedTags DefinedTagSet) (*TagSlug, error) {
	slug := &protobuf.Slug{}

	if freeformTags != nil {
		slug.FreeformTags = &protobuf.StringMap{Entries: freeformTags}
	}

	if definedTags != nil {
		slug.DefinedTags = &protobuf.NamespaceMap{Namespaces: map[string]*protobuf.NamespaceDefinedTags{}}

		for ns, dts := range definedTags {
			nsTags := &protobuf.NamespaceDefinedTags{Tags: map[string]*protobuf.DefinedTagValue{}}
			for k, v := range dts {
				nsTags.Tags[k] = &protobuf.DefinedTagValue{Value: []string{v}}
			}
			slug.DefinedTags.Namespaces[ns] = nsTags
		}
	}

	return Marshal(slug)
}

// Marshal serializes a `protobuf.Slug` struct into a `*TagSlug` byte slice, prepending a 4-byte
// header required for IAM.
func Marshal(protobufSlug *protobuf.Slug) (*TagSlug, error) {
	slug, err := proto.Marshal(protobufSlug)
	if err != nil {
		return nil, err
	}

	// Add slug headers
	s := TagSlug(append(ProtobufHeaders, slug...))
	return &s, nil
}

// unmarshal unserializes a `TagSlug` byte slice into a
// `*protobuf.Slug` struct. Assumes `slug` has `ProtobufHeaders`
// prefixed.
func unmarshal(slug TagSlug) (*protobuf.Slug, error) {
	slug = bytes.TrimPrefix(slug, ProtobufHeaders)

	s := &protobuf.Slug{}
	if err := proto.Unmarshal([]byte(slug), s); err != nil {
		return nil, err
	}

	return s, nil
}

// FreeformTagSet returns the decoded freeform tag set.
func (ts *TagSlug) FreeformTagSet() (*FreeformTagSet, error) {
	s, err := unmarshal(*ts)
	if err != nil {
		return nil, err
	}

	set := &FreeformTagSet{}
	entries := s.GetFreeformTags().GetEntries()
	if entries != nil {
		*set = entries
	}

	return set, nil
}

// SwaggerFreeformTagSet returns the decoded freeform tags in a type
// compatible with the swagger auto-generated freeform tags type.
func (ts *TagSlug) SwaggerFreeformTagSet() (map[string]*string, error) {
	s, err := unmarshal(*ts)
	if err != nil {
		return nil, err
	}

	set := make(map[string]*string)
	for k, v := range s.GetFreeformTags().GetEntries() {
		// Assign a new var to handle range copying the value
		ptrV := v
		set[k] = &ptrV
	}

	return set, nil
}

// DefinedTagSet returns the decoded defined tag set.
func (ts *TagSlug) DefinedTagSet() (*DefinedTagSet, error) {
	s, err := unmarshal(*ts)
	if err != nil {
		return nil, err
	}

	set := DefinedTagSet{}
	namespaces := s.GetDefinedTags().GetNamespaces()
	if namespaces == nil {
		return &set, nil
	}

	for ns, tags := range namespaces {
		set[ns] = make(map[string]string)
		for k, v := range tags.GetTags() {
			if _, ok := set[ns][k]; !ok {
				set[ns][k] = v.GetValue()[0]
			}
		}
	}

	return &set, nil
}

// SwaggerDefinedTagSet returns the decoded defined tags in a type
// compatible with the swagger auto-generated defined tags type.
func (ts *TagSlug) SwaggerDefinedTagSet() (map[string]map[string]interface{}, error) {
	s, err := unmarshal(*ts)
	if err != nil {
		return nil, err
	}

	set := make(map[string]map[string]interface{})
	for ns, tags := range s.GetDefinedTags().GetNamespaces() {
		m := make(map[string]interface{})
		set[ns] = m
		for k, v := range tags.GetTags() {
			if _, ok := set[ns][k]; !ok {
				set[ns][k] = v.GetValue()[0]
			}
		}
	}

	return set, nil
}

// Copy creates a new copy of a tag slug
func (ts *TagSlug) Copy() *TagSlug {
	if ts == nil {
		return nil
	}
	newSlug := make(TagSlug, len(*ts))
	copy(newSlug, *ts)
	return &newSlug
}

// AddSlugHeaders will add `ProtobufHeaders` if missing, then return
// the byte slice. This function is useful for remdiating bad data,
// such as tag slugs persisted to a db without headers.
func AddSlugHeaders(b []byte) (TagSlug, error) {
	s := &protobuf.Slug{}

	// This function will eventually cause all slugs to converge to
	// the correct state in which a slug is composed of a 4-byte
	// header concatenated with a protobuf-compatible byte slice.

	// Assume it is unlikely that a protobuf slug will start with the
	// same bytes as `ProtobufHeaders` (1, 3, 2, 2) and try to
	// unmarshal it after trimming the headers. If that fails, don't
	// trim the prefix as this may be an instance of that unlikely
	// case. And if that fails, we have something we don't recognize
	// so return the error.
	if bytes.HasPrefix(b, ProtobufHeaders) {
		trimmed := bytes.TrimPrefix(b, ProtobufHeaders)
		if len(trimmed) == 0 {
			return b, nil
		}
		if err := proto.Unmarshal(trimmed, s); err != nil {
			if err := proto.Unmarshal(b, s); err != nil {
				return nil, err
			}
			return TagSlug(append(ProtobufHeaders, b...)), nil
		}
		return b, nil
	}

	if err := proto.Unmarshal(b, s); err != nil {
		return nil, err
	}

	return TagSlug(append(ProtobufHeaders, b...)), nil
}
