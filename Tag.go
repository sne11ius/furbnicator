package main

import "strings"

type MatchMode string

const (
	Contains   MatchMode = "Contains"
	StartsWith MatchMode = "StartsWith"
	EndsWith   MatchMode = "EndsWith"
	Equals     MatchMode = "Equals"
)

type Tag struct {
	value     string
	matchMode MatchMode
}

func DoMatch(strs []string, tags []Tag) bool {
	if len(tags) == 0 || (len(tags) == 1 && tags[0].value == "") {
		return true
	}
	for _, tag := range tags {
		found := false
		for _, str := range strs {
			if tag.Matches(str) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (t *Tag) Matches(str string) bool {
	str = strings.ToLower(str)
	tagVal := strings.ToLower(t.value)
	switch {
	case t.matchMode == Contains:
		return strings.Contains(str, tagVal)
	case t.matchMode == StartsWith:
		return strings.HasPrefix(str, tagVal)
	case t.matchMode == EndsWith:
		return strings.HasSuffix(str, tagVal)
	case t.matchMode == Equals:
		return strings.EqualFold(str, tagVal)
	default:
		panic("Unknown matchMode: " + t.matchMode)
	}
}

func TagsFromStrings(strs []string) []Tag {
	var tags []Tag
	for _, str := range strs {
		if strings.HasPrefix(str, "=") {
			tagValue := strings.TrimPrefix(str, "=")
			tags = append(tags, Tag{
				value:     tagValue,
				matchMode: Equals,
			})
		} else if strings.HasPrefix(str, "+") {
			if strings.HasSuffix(str, "+") {
				tagValue := strings.TrimPrefix(strings.TrimSuffix(str, "+"), "+")
				tags = append(tags, Tag{
					value:     tagValue,
					matchMode: Equals,
				})
			} else {
				tagValue := strings.TrimPrefix(str, "+")
				tags = append(tags, Tag{
					value:     tagValue,
					matchMode: StartsWith,
				})
			}
		} else if strings.HasSuffix(str, "+") {
			tagValue := strings.TrimSuffix(str, "+")
			tags = append(tags, Tag{
				value:     tagValue,
				matchMode: EndsWith,
			})
		} else {
			tags = append(tags, Tag{
				value:     str,
				matchMode: Contains,
			})
		}
	}
	return tags
}
