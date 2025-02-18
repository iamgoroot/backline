package mkdocs

import (
	"fmt"

	"github.com/iamgoroot/backline/plugin/documentation/techdocs/model"
	"gopkg.in/yaml.v3"
)

type mkDocs struct {
	Nav      any    `yaml:"nav"`
	SiteName string `yaml:"site_name"`
}

func ParseDocIndex(data []byte, walkFunc func(url string)) (*model.Nav, error) {
	mkdocsDef := &mkDocs{}

	err := yaml.Unmarshal(data, mkdocsDef)
	if err != nil {
		return nil, err
	}

	nav := parseNav(mkdocsDef.SiteName, mkdocsDef.Nav, walkFunc)
	nav.Name = mkdocsDef.SiteName

	return nav, nil
}

func parseNav(parent string, nav any, walkFunc func(url string)) *model.Nav {
	switch navItem := nav.(type) {
	case string:
		return &model.Nav{Name: navItem}
	case []byte:
		return &model.Nav{Name: string(navItem)}
	case map[string]any:
		for navName, navContent := range navItem {
			newParent := fmt.Sprintf("%s/%s", parent, navName)

			if url, ok := navContent.(string); ok {
				var nav = model.Nav{Name: navName}
				nav.URL = url
				walkFunc(url)

				return &nav
			}

			nav := parseNav(newParent, navContent, walkFunc)
			nav.Name = navName

			return nav
		}
	case []any:
		var nav model.Nav
		for _, v := range navItem {
			nav.Children = append(nav.Children, parseNav(parent, v, walkFunc))
		}

		return &nav
	}

	return nil
}
