package views

import (
	"fmt"
	"net/url"
)

func getContentURL(entityName, contentKey string) string {
	entityName = url.QueryEscape(entityName)
	contentKey = url.QueryEscape(contentKey)

	return fmt.Sprintf("/techdocs/item/by-fullname/%s/by-item-path/%s", entityName, contentKey)
}
