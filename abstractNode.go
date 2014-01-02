package asciidocgo

import (
	"strings"
	"unicode/utf8"
)

/* An abstract base class that provides state and methods for managing
a node of AsciiDoc content.
The state and methods on this class are comment to all content segments
in an AsciiDoc document. */
type abstractNode struct {
	parent     *abstractNode
	context    context
	document   *abstractNode
	attributes map[string]interface{}
	*substitutors
}

func newAbstractNode(parent *abstractNode, context context) *abstractNode {
	abstractNode := &abstractNode{parent, context, nil, make(map[string]interface{}), &substitutors{}}
	if context == document {
		abstractNode.parent = nil
		abstractNode.document = parent
	} else if parent != nil {
		abstractNode.document = parent.Document()
	}
	return abstractNode
}

//  Get the element which is the parent of this node
func (an *abstractNode) Parent() *abstractNode {
	return an.parent
}

//  Get the Asciidoctor::Document to which this node belongs
func (an *abstractNode) Document() *abstractNode {
	return an.document
}

// Get the Symbol context for this node
func (an *abstractNode) Context() context {
	return an.context
}

func (an *abstractNode) Attributes() map[string]interface{} {
	return an.attributes
}

// Associate this Block with a new parent Block
// parent: The Block to set as the parent of this Block
func (an *abstractNode) SetParent(parent *abstractNode) {
	an.parent = parent
	an.document = parent.Document()
}

/* Get the value for the specified attribute.

First look in the attributes on this node and return the value
of the attribute if found.
Otherwise, if this node is a child of the Document node, look in
the attributes of the Document node and return the value of the attribute
if found.
Otherwise, return the default value, which defaults to nil.

- name: the String or Symbol name of the attribute to lookup
- default_value: the Object value to return if the attribute is not found
(default: nil)
- inherit: a Boolean indicating whether to check for the attribute on
the AsciiDoctor::Document if not found on this node (default: false)

Return the value of the attribute or the default value if the attribute is
not found in the attributes of this node or the document node
*/
func (an *abstractNode) Attr(name string, defaultValue interface{}, inherit bool) interface{} {
	if an == an.document {
		inherit = false
	}
	if an.attributes[name] != nil {
		return an.attributes[name]
	}
	if inherit {
		if an.document != nil && an.document.attributes[name] != nil {
			return an.document.attributes[name]
		}
	}
	return defaultValue
}

/* Assign the value to the specified key in this block's attributes hash.

- key: The attribute key (or name)
- val: The value to assign to the key

returns a flag indicating whether the assignment was performed
*/
func (an *abstractNode) setAttr(name string, val interface{}, override bool) bool {
	if override {
		an.attributes[name] = val
		return true
	}
	if _, hasName := an.attributes[name]; !hasName {
		an.attributes[name] = val
		return true
	}
	return false
}

/* Enable a specified option attribute on the current node.

This method defines the `%name%-option` attribute on the current node.

- name: the String or Symbol name of the option
*/
func (an *abstractNode) SetOption(option string) {
	val := an.attributes["options"]
	if val == nil {
		valmap := make(map[string]bool)
		an.attributes["options"] = valmap
		val = valmap
	}
	valmap := val.(map[string]bool)
	if _, hasOption := valmap[option]; !hasOption {
		valmap[option] = true
	}
	an.attributes[option+"-option"] = true
}

/*  A convenience method to check if the specified option attribute is enabled
on the current node.

Check if the option is enabled.
This method simply checks to see if the `%name%-option` attribute is defined
on the current node.

- name: the String or Symbol name of the option

return a Boolean indicating whether the option has been specified
*/
func (an *abstractNode) Option(option string) bool {
	_, res := an.attributes[option+"-option"]
	return res
}

/* Update the attributes of this node with the new values
in the attributes argument.

If an attribute already exists with the same key,
it's value will be overridden.

- attributes: A Hash of attributes to assign to this node.
*/
func (an *abstractNode) Update(attrs map[string]interface{}) {
	for key, value := range attrs {
		an.attributes[key] = value
	}
}

// A convenience method that checks if the specified role is present
// in the list of roles on this node
func (an *abstractNode) HasRole(role interface{}) bool {
	if role == nil {
		if _, hasRole := an.attributes["role"]; hasRole {
			return true
		}
		if an.Document() != nil {
			if _, hasRole := an.Document().attributes["role"]; hasRole {
				return true
			}
		}
		return false
	}
	if anAttr := an.Attr("role", nil, true); anAttr == role {
		return true
	}
	return false
}

// A convenience method that checks if the specified role is present
// in the list of roles on this node
func (an *abstractNode) HasARole(name string) bool {
	if name == "" {
		return false
	}
	roles := an.Attr("role", nil, true)
	if roles == nil {
		return false
	}
	rolesString := roles.(string)
	rolesArray := strings.Split(rolesString, " ")
	for _, role := range rolesArray {
		if name == role {
			return true
		}
	}
	return false
}

// A convenience method that returns the value of the role attribute
func (an *abstractNode) Role() interface{} {
	return an.Attr("role", nil, true)
}

// A convenience method that returns the role names as an Array
func (an *abstractNode) RoleNames() []string {
	roles := an.Attr("role", nil, true)
	if roles == nil {
		return []string{}
	}
	rolesString := roles.(string)
	return strings.Split(rolesString, " ")
}

// A convenience method that checks if the reftext attribute is specified
func (an *abstractNode) HasReftext() bool {
	reftext := an.Attr("reftext", nil, true)
	return (reftext != nil)
}

// A convenience method that returns the value of the reftext attribute
func (an *abstractNode) Reftext() interface{} {
	return an.Attr("reftext", nil, true)
}

// Returns a forward slash if the attribute htmlsyntax has the value "xml".
func (an *abstractNode) ShortTagSlash() *rune {
	if an.Document() == nil {
		return nil
	}
	if an.Document().Attr("htmlsyntax", nil, false) == "xml" {
		r, _ := utf8.DecodeLastRuneInString("/")
		return &r
	} else {
		return nil
	}
}

/* Construct a URI reference to the target media.

If the target media is a URI reference, then leave it untouched.

The target media is resolved relative to the directory retrieved from
the specified attribute key, if provided.

The return value can be safely used in a media tag (img, audio, video).

target        - A String reference to the target media
asset_dir_key - The String attribute key used to lookup the directory where

Returns A String reference for the target media
*/
func (an *abstractNode) MediaUri(target string, assetDirKey string) string {
	if strings.Contains(target, ":") && REGEXP[":uri_sniff"].MatchString(target) {
		return target
	} else {
		if assetDirKey != "" && an.HasAttr(assetDirKey, nil, true) {
			// normalize_web_path(target, @document.attr(asset_dir_key))
			return ""
		}
	}
	// normalize_web_path(target)
	return ""
}