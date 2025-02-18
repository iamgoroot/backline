package views

type kind struct {
	Path     string
	Endpoint string
	Name     string
}

var kinds = []kind{
	{Path: "/catalog/view/entities/kind/component", Endpoint: "/catalog/component/entities/kind/Component", Name: "Components"},
	{Path: "/catalog/view/entities/kind/api", Endpoint: "/catalog/component/entities/kind/API", Name: "APIs"},
	{Path: "/catalog/view/entities/kind/group", Endpoint: "/catalog/component/entities/kind/Group", Name: "Groups"},
	{Path: "/catalog/view/entities/kind/user", Endpoint: "/catalog/component/entities/kind/User", Name: "Users"},
	{Path: "/catalog/view/entities/kind/resource", Endpoint: "/catalog/component/entities/kind/Resource", Name: "Resources"},
	{Path: "/catalog/view/entities/kind/system", Endpoint: "/catalog/component/entities/kind/System", Name: "Systems"},
	{Path: "/catalog/view/entities/kind/domain", Endpoint: "/catalog/component/entities/kind/Domain", Name: "Domains"},
}
