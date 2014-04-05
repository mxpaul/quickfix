//Package datadictionary provides support for parsing and organizing FIX Data Dictionaries
package datadictionary

import (
	"encoding/xml"
	"github.com/cbusbey/quickfixgo/tag"
	"os"
)

//DataDictionary models FIX messages, components, and fields.
type DataDictionary struct {
	FieldTypeByTag  map[tag.Tag]*FieldType
	FieldTypeByName map[string]*FieldType
	Messages        map[string]*MessageDef
	Components      map[string]*Component
	Header          *MessageDef
	Trailer         *MessageDef
}

//Component is a grouping of fields.
type Component struct {
	Fields []*FieldDef
}

//TagSet is set for tags.
type TagSet map[tag.Tag]struct{}

//Add adds a tag to the tagset.
func (t TagSet) Add(tag tag.Tag) {
	t[tag] = struct{}{}
}

//FieldDef models a field belonging to a message.
type FieldDef struct {
	*FieldType
	Required    bool
	ChildFields []*FieldDef
}

//IsGroup is true if the field is a repeating group.
func (f FieldDef) IsGroup() bool {
	return len(f.ChildFields) > 0
}

func (f FieldDef) childTags() []tag.Tag {
	tags := make([]tag.Tag, 0, len(f.ChildFields))

	for _, f := range f.ChildFields {
		tags = append(tags, f.Tag)
		for _, t := range f.childTags() {
			tags = append(tags, t)
		}
	}

	return tags
}

func (f FieldDef) requiredChildTags() []tag.Tag {
	tags := make([]tag.Tag, 0)

	for _, f := range f.ChildFields {
		if !f.Required {
			continue
		}

		tags = append(tags, f.Tag)
		for _, t := range f.requiredChildTags() {
			tags = append(tags, t)
		}
	}

	return tags

	return []tag.Tag{}
}

//FieldType holds information relating to a field.  Includes Tag, type, and enums, if defined.
type FieldType struct {
	Name  string
	Tag   tag.Tag
	Type  string
	Enums map[string]Enum
}

//Enum is a container for value and description.
type Enum struct {
	Value       string
	Description string
}

//MessageDef can apply to header, trailer, or body of a FIX Message.
type MessageDef struct {
	Fields map[tag.Tag]*FieldDef

	RequiredTags TagSet
	Tags         TagSet
}

//Parse loads and and build a datadictionary instance from an xml file.
func Parse(path string) (*DataDictionary, error) {
	var xmlFile *os.File
	xmlFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer xmlFile.Close()

	doc := new(XMLDoc)
	decoder := xml.NewDecoder(xmlFile)
	if err := decoder.Decode(doc); err != nil {
		return nil, err
	}

	b := new(builder)
	var dict *DataDictionary
	if dict, err = b.build(doc); err != nil {
		return nil, err
	}

	return dict, nil
}
