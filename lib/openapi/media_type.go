package openapi

import "regexp"

// MediaKind - represents the sets of media type
type MediaKind int

// MediaKind enums
const (
	MediaKindUnsupported MediaKind = iota
	MediaKindPlaintext   MediaKind = iota
	MediaKindJSON
	MediaKindXML
	MediaKindMultipartRelated
)

var (
	jsonRegex             = regexp.MustCompile(`(?i:(?:application|text)/(?:[a-zA-Z0-9./-]+\+)?json)`)
	xmlRegex              = regexp.MustCompile(`(?i:(?:application|text)/xml)`)
	multipartRelatedRegex = regexp.MustCompile("(?i:multipart/related)")
)

// KindOfMediaType - returns Mediakind of the media type
func KindOfMediaType(mediaType string) MediaKind {
	if jsonRegex.MatchString(mediaType) {
		return MediaKindJSON
	} else if xmlRegex.MatchString(mediaType) {
		return MediaKindXML
	} else if multipartRelatedRegex.MatchString(mediaType) {
		return MediaKindMultipartRelated
	}
	return MediaKindUnsupported
}
